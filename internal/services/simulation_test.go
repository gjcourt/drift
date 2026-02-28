package services

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/gjcourt/drift/internal/domain"
)

func TestEstimateGBMParamsFlat(t *testing.T) {
	recs := make([]domain.PriceRecord, 50)
	for i := range recs {
		recs[i] = domain.PriceRecord{AdjustedClose: 100.0}
	}
	mu, sigma := estimateGBMParams(recs)
	if math.Abs(mu) > 1e-9 || math.Abs(sigma) > 1e-9 {
		t.Errorf("flat prices: mu=%v sigma=%v, both want ~0", mu, sigma)
	}
}

func TestEstimateGBMParamsTwoRecords(t *testing.T) {
	recs := []domain.PriceRecord{
		{AdjustedClose: 100},
		{AdjustedClose: 110},
	}
	_, sigma := estimateGBMParams(recs)
	if sigma < 0 {
		t.Errorf("sigma = %v, must be non-negative", sigma)
	}
}

func TestGBMPathShape(t *testing.T) {
	cfg := domain.SimulationConfig{
		NumPaths:    10,
		HorizonDays: 252,
		StartValue:  100_000,
	}
	params := []assetGBMParams{{mu: 0.07, sigma: 0.15}}
	weights := []float64{1.0}
	rng := rand.New(rand.NewChaCha8([32]byte{}))

	path := gbmPath(cfg, params, weights, rng)
	if len(path.Values) != cfg.HorizonDays+1 {
		t.Errorf("path length = %d, want %d", len(path.Values), cfg.HorizonDays+1)
	}
	if path.Values[0] != cfg.StartValue {
		t.Errorf("path[0] = %v, want %v", path.Values[0], cfg.StartValue)
	}
	if path.Values[len(path.Values)-1] <= 0 {
		t.Errorf("terminal value %v must be positive", path.Values[len(path.Values)-1])
	}
}

func TestGBMPathDeterministic(t *testing.T) {
	cfg := domain.SimulationConfig{
		NumPaths:    1,
		HorizonDays: 100,
		StartValue:  50_000,
	}
	params := []assetGBMParams{{mu: 0.08, sigma: 0.20}}
	weights := []float64{1.0}

	key := [32]byte{1, 2, 3}
	rng1 := rand.New(rand.NewChaCha8(key))
	rng2 := rand.New(rand.NewChaCha8(key))

	p1 := gbmPath(cfg, params, weights, rng1)
	p2 := gbmPath(cfg, params, weights, rng2)

	for i := range p1.Values {
		if p1.Values[i] != p2.Values[i] {
			t.Errorf("path[%d]: %v != %v (not deterministic)", i, p1.Values[i], p2.Values[i])
		}
	}
}

func TestBsPathShape(t *testing.T) {
	cfg := domain.SimulationConfig{
		HorizonDays: 100,
		StartValue:  10_000,
	}
	returns := [][]float64{{0.001, -0.001, 0.002, -0.002}}
	weights := []float64{1.0}
	rng := rand.New(rand.NewChaCha8([32]byte{}))

	path := bsPath(cfg, returns, weights, rng)
	if len(path.Values) != cfg.HorizonDays+1 {
		t.Errorf("path length = %d, want %d", len(path.Values), cfg.HorizonDays+1)
	}
	if path.Values[0] != cfg.StartValue {
		t.Errorf("path[0] = %v, want %v", path.Values[0], cfg.StartValue)
	}
}

func TestSeedKey(t *testing.T) {
	k1 := seedKey(42)
	k2 := seedKey(43)
	if k1 == k2 {
		t.Error("different seeds must produce different keys")
	}
}
