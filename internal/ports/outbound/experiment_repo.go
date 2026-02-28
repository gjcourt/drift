package outbound

import (
	"context"

	"github.com/gjcourt/drift/internal/domain"
)

// ExperimentRepository is the outbound port for persisting experiment configurations.
type ExperimentRepository interface {
	SaveExperiment(ctx context.Context, exp domain.Experiment) error
	GetExperiment(ctx context.Context, id string) (*domain.Experiment, error)
	ListExperiments(ctx context.Context) ([]domain.Experiment, error)
	DeleteExperiment(ctx context.Context, id string) error
}
