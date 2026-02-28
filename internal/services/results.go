package services

import (
	"context"
	"time"

	"github.com/gjcourt/drift/internal/domain"
	"github.com/gjcourt/drift/internal/ports/outbound"
)

type resultsSvc struct {
	experimentRepo outbound.ExperimentRepository
	simulationRepo outbound.SimulationRepository
}

// NewResultsService constructs a ResultsService backed by the given repositories.
func NewResultsService(er outbound.ExperimentRepository, sr outbound.SimulationRepository) *resultsSvc {
	return &resultsSvc{experimentRepo: er, simulationRepo: sr}
}

func (s *resultsSvc) CreateExperiment(ctx context.Context, exp domain.Experiment) (*domain.Experiment, error) {
	if exp.ID == "" {
		exp.ID = newID("exp")
	}
	exp.CreatedAt = time.Now().UTC()
	exp.UpdatedAt = exp.CreatedAt
	if err := s.experimentRepo.SaveExperiment(ctx, exp); err != nil {
		return nil, err
	}
	return &exp, nil
}

func (s *resultsSvc) GetExperiment(ctx context.Context, id string) (*domain.Experiment, error) {
	return s.experimentRepo.GetExperiment(ctx, id)
}

func (s *resultsSvc) ListExperiments(ctx context.Context) ([]domain.Experiment, error) {
	return s.experimentRepo.ListExperiments(ctx)
}

func (s *resultsSvc) ListRuns(ctx context.Context, experimentID string) ([]domain.Run, error) {
	return s.simulationRepo.ListRuns(ctx, experimentID)
}

func (s *resultsSvc) GetRunStats(ctx context.Context, runID string) (*domain.ResultStats, error) {
	run, err := s.simulationRepo.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	return &run.Stats, nil
}
