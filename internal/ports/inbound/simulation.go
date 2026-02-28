package inbound

import (
	"context"

	"github.com/gjcourt/drift/internal/domain"
)

// SimulationService is the inbound port for running Monte Carlo simulations.
type SimulationService interface {
	RunExperiment(ctx context.Context, experimentID string) (*domain.Run, error)
	GetRun(ctx context.Context, runID string) (*domain.Run, error)
	GetRunPaths(ctx context.Context, runID string) ([]domain.SimulatedPath, error)
}
