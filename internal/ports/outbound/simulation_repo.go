package outbound

import (
	"context"

	"github.com/gjcourt/drift/internal/domain"
)

// SimulationRepository is the outbound port for persisting runs and their results.
type SimulationRepository interface {
	SaveRun(ctx context.Context, run domain.Run) error
	GetRun(ctx context.Context, runID string) (*domain.Run, error)
	ListRuns(ctx context.Context, experimentID string) ([]domain.Run, error)
}
