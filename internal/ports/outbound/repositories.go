// Package outbound defines the repository and infrastructure interfaces used by the app layer.
// Storage adapters (e.g. SQLite) implement these interfaces.
package outbound

import (
	"context"

	"github.com/gjcourt/drift/internal/domain"
)

// AssetRepository is the outbound port for persisting and querying assets and price records.
type AssetRepository interface {
	UpsertAsset(ctx context.Context, asset domain.Asset) error
	GetAsset(ctx context.Context, symbol string) (*domain.Asset, error)
	ListAssets(ctx context.Context) ([]domain.Asset, error)
	DeleteAsset(ctx context.Context, symbol string) error
	UpsertPriceRecords(ctx context.Context, records []domain.PriceRecord) error
	GetPriceRecords(ctx context.Context, symbol string, limit int) ([]domain.PriceRecord, error)
}

// ExperimentRepository is the outbound port for persisting experiment configurations.
type ExperimentRepository interface {
	SaveExperiment(ctx context.Context, exp domain.Experiment) error
	GetExperiment(ctx context.Context, id string) (*domain.Experiment, error)
	ListExperiments(ctx context.Context) ([]domain.Experiment, error)
	DeleteExperiment(ctx context.Context, id string) error
}

// SimulationRepository is the outbound port for persisting runs and their results.
type SimulationRepository interface {
	SaveRun(ctx context.Context, run domain.Run) error
	GetRun(ctx context.Context, runID string) (*domain.Run, error)
	ListRuns(ctx context.Context, experimentID string) ([]domain.Run, error)
}
