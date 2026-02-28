package services

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"

	"github.com/gjcourt/drift/internal/domain"
	"github.com/gjcourt/drift/internal/ports/outbound"
)

type simulationSvc struct {
	assetRepo      outbound.AssetRepository
	simulationRepo outbound.SimulationRepository
	experimentRepo outbound.ExperimentRepository
}

// NewSimulationService constructs a SimulationService backed by the given repositories.
func NewSimulationService(ar outbound.AssetRepository, sr outbound.SimulationRepository, er outbound.ExperimentRepository) *simulationSvc {
	return &simulationSvc{assetRepo: ar, simulationRepo: sr, experimentRepo: er}
}

func (s *simulationSvc) RunExperiment(ctx context.Context, experimentID string) (*domain.Run, error) {
	exp, err := s.experimentRepo.GetExperiment(ctx, experimentID)
	if err != nil {
		return nil, fmt.Errorf("get experiment: %w", err)
	}
	run := domain.Run{
		ID:           newID("run"),
		ExperimentID: experimentID,
		StartedAt:    time.Now().UTC(),
		Status:       domain.StatusRunning,
	}
	if err := s.simulationRepo.SaveRun(ctx, run); err != nil {
		return nil, fmt.Errorf("save run: %w", err)
	}
	paths, simErr := s.simulate(ctx, exp)
	now := time.Now().UTC()
	run.FinishedAt = &now
	if simErr != nil {
		run.Status = domain.StatusFailed
		run.Error = simErr.Error()
		_ = s.simulationRepo.SaveRun(ctx, run)
		return nil, simErr
	}
	run.Stats = domain.ComputeStats(paths, exp.Config.StartValue, float64(exp.Config.HorizonDays)/252.0)
	run.Status = domain.StatusComplete
	if err := s.simulationRepo.SaveRun(ctx, run); err != nil {
		return nil, err
	}
	return &run, nil
}

func (s *simulationSvc) GetRun(ctx context.Context, runID string) (*domain.Run, error) {
	return s.simulationRepo.GetRun(ctx, runID)
}

func (s *simulationSvc) GetRunPaths(_ context.Context, _ string) ([]domain.SimulatedPath, error) {
	return nil, nil
}

func (s *simulationSvc) simulate(ctx context.Context, exp *domain.Experiment) ([]domain.SimulatedPath, error) {
	switch exp.Config.Model {
	case domain.ModelGBM:
		return s.runGBM(ctx, exp)
	case domain.ModelBootstrap, domain.ModelBlockBootstrap:
		return s.runBootstrap(ctx, exp)
	default:
		return nil, fmt.Errorf("unknown model: %s", exp.Config.Model)
	}
}

type assetGBMParams struct{ mu, sigma float64 }

func (s *simulationSvc) runGBM(ctx context.Context, exp *domain.Experiment) ([]domain.SimulatedPath, error) {
	params := make([]assetGBMParams, len(exp.Portfolio.Assets))
	weights := make([]float64, len(exp.Portfolio.Assets))
	for i, pa := range exp.Portfolio.Assets {
		recs, err := s.assetRepo.GetPriceRecords(ctx, pa.Symbol, exp.Config.LookbackDays+1)
		if err != nil {
			return nil, fmt.Errorf("prices %s: %w", pa.Symbol, err)
		}
		if len(recs) < 2 {
			return nil, fmt.Errorf("need >=2 records for %s", pa.Symbol)
		}
		mu, sig := estimateGBMParams(recs)
		params[i] = assetGBMParams{mu: mu, sigma: sig}
		weights[i] = pa.Weight
	}
	return s.workerPool(exp.Config, func(rng *rand.Rand) domain.SimulatedPath {
		return gbmPath(exp.Config, params, weights, rng)
	}), nil
}

func estimateGBMParams(recs []domain.PriceRecord) (mu, sigma float64) {
	var lr []float64
	for i := 1; i < len(recs); i++ {
		p, c := recs[i-1].AdjustedClose, recs[i].AdjustedClose
		if p > 0 && c > 0 {
			lr = append(lr, math.Log(c/p))
		}
	}
	if len(lr) == 0 {
		return 0, 0
	}
	n := float64(len(lr))
	var sum float64
	for _, r := range lr {
		sum += r
	}
	mean := sum / n
	var vsum float64
	for _, r := range lr {
		d := r - mean
		vsum += d * d
	}
	dsig := math.Sqrt(vsum / n)
	return mean*252 + 0.5*dsig*dsig*252, dsig * math.Sqrt(252)
}

func gbmPath(cfg domain.SimulationConfig, params []assetGBMParams, weights []float64, rng *rand.Rand) domain.SimulatedPath {
	dt := 1.0 / 252.0
	vals := make([]float64, cfg.HorizonDays+1)
	vals[0] = cfg.StartValue
	growth := make([]float64, len(params))
	for i := range growth {
		growth[i] = 1.0
	}
	for day := 1; day <= cfg.HorizonDays; day++ {
		for i, p := range params {
			growth[i] *= math.Exp((p.mu-0.5*p.sigma*p.sigma)*dt + p.sigma*math.Sqrt(dt)*rng.NormFloat64())
		}
		var total float64
		for i, w := range weights {
			total += w * growth[i]
		}
		vals[day] = cfg.StartValue * total
		if cfg.AnnualContribution != 0 && day%252 == 0 {
			vals[day] += cfg.AnnualContribution
		}
	}
	return domain.SimulatedPath{Values: vals}
}

func (s *simulationSvc) runBootstrap(ctx context.Context, exp *domain.Experiment) ([]domain.SimulatedPath, error) {
	rs := make([][]float64, len(exp.Portfolio.Assets))
	wts := make([]float64, len(exp.Portfolio.Assets))
	for i, pa := range exp.Portfolio.Assets {
		recs, err := s.assetRepo.GetPriceRecords(ctx, pa.Symbol, exp.Config.LookbackDays+1)
		if err != nil {
			return nil, err
		}
		var rtns []float64
		for j := 1; j < len(recs); j++ {
			p, c := recs[j-1].AdjustedClose, recs[j].AdjustedClose
			if p > 0 && c > 0 {
				rtns = append(rtns, math.Log(c/p))
			}
		}
		rs[i], wts[i] = rtns, pa.Weight
	}
	return s.workerPool(exp.Config, func(rng *rand.Rand) domain.SimulatedPath {
		return bsPath(exp.Config, rs, wts, rng)
	}), nil
}

func bsPath(cfg domain.SimulationConfig, rs [][]float64, wts []float64, rng *rand.Rand) domain.SimulatedPath {
	vals := make([]float64, cfg.HorizonDays+1)
	vals[0] = cfg.StartValue
	for day := 1; day <= cfg.HorizonDays; day++ {
		var lr float64
		for i, s := range rs {
			if len(s) > 0 {
				lr += wts[i] * s[rng.IntN(len(s))]
			}
		}
		vals[day] = vals[day-1] * math.Exp(lr)
		if cfg.AnnualContribution != 0 && day%252 == 0 {
			vals[day] += cfg.AnnualContribution
		}
	}
	return domain.SimulatedPath{Values: vals}
}

func (s *simulationSvc) workerPool(cfg domain.SimulationConfig, gen func(*rand.Rand) domain.SimulatedPath) []domain.SimulatedPath {
	nw := runtime.NumCPU()
	ch := make(chan domain.SimulatedPath, cfg.NumPaths)
	var base uint64
	if cfg.Seed != nil {
		base = uint64(*cfg.Seed)
	} else {
		base = uint64(time.Now().UnixNano())
	}
	batch := cfg.NumPaths / nw
	var wg sync.WaitGroup
	for w := 0; w < nw; w++ {
		cnt := batch
		if w == nw-1 {
			cnt += cfg.NumPaths % nw
		}
		seed := base + uint64(w)*1_000_003
		wg.Add(1)
		go func(n int, seed uint64) {
			defer wg.Done()
			rng := rand.New(rand.NewChaCha8(seedKey(seed)))
			for i := 0; i < n; i++ {
				ch <- gen(rng)
			}
		}(cnt, seed)
	}
	go func() { wg.Wait(); close(ch) }()
	paths := make([]domain.SimulatedPath, 0, cfg.NumPaths)
	for p := range ch {
		paths = append(paths, p)
	}
	return paths
}

func seedKey(seed uint64) [32]byte {
	var k [32]byte
	for i := range 8 {
		k[i] = byte(seed >> (uint(i) * 8))
	}
	return k
}
