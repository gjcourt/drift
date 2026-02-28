// Package inbound defines the service interfaces called by inbound adapters
// (e.g. the HTTP layer). Services implement these interfaces.
package inbound

import (
	"context"
	"io"

	"github.com/gjcourt/drift/internal/domain"
)

// DataIngestionService is the inbound port for loading historical price data.
type DataIngestionService interface {
	IngestCSV(ctx context.Context, r io.Reader, filename string) (int, error)
	ListAssets(ctx context.Context) ([]domain.Asset, error)
	GetAssetPrices(ctx context.Context, symbol string, limit int) ([]domain.PriceRecord, error)
	DeleteAsset(ctx context.Context, symbol string) error
}
