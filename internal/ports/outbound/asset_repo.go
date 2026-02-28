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
