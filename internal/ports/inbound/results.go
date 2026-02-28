package inbound

import (
	"context"

	"github.com/gjcourt/drift/internal/domain"
)

// ResultsService is the inbound port for querying simulation results and experiments.
type ResultsService interface {
	CreateExperiment(ctx context.Context, exp domain.Experiment) (*domain.Experiment, error)
	GetExperiment(ctx context.Context, id string) (*domain.Experiment, error)
	ListExperiments(ctx context.Context) ([]domain.Experiment, error)
	ListRuns(ctx context.Context, experimentID string) ([]domain.Run, error)
	GetRunStats(ctx context.Context, runID string) (*domain.ResultStats, error)
}
