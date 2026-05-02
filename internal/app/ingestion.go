// Package app contains use-case orchestration for simulation, ingestion, and results.
// It depends only on internal/domain and internal/ports; it must not import concrete adapter packages.
package app

import (
	"context"
	"io"

	"github.com/gjcourt/drift/internal/domain"
	"github.com/gjcourt/drift/internal/ports/outbound"
)

type ingestionSvc struct {
	csvParser outbound.CSVParser
	assetRepo outbound.AssetRepository
}

// NewIngestionService constructs an IngestionService backed by the given parser and asset repository.
func NewIngestionService(cp outbound.CSVParser, ar outbound.AssetRepository) *ingestionSvc {
	return &ingestionSvc{csvParser: cp, assetRepo: ar}
}

func (s *ingestionSvc) IngestCSV(ctx context.Context, r io.Reader, filename string) (int, error) {
	records, err := s.csvParser.ParseCSV(r, filename)
	if err != nil {
		return 0, err
	}
	symbolsSeen := map[string]bool{}
	for _, rec := range records {
		if !symbolsSeen[rec.Symbol] {
			if err := s.assetRepo.UpsertAsset(ctx, domain.Asset{Symbol: rec.Symbol, Name: rec.Symbol}); err != nil {
				return 0, err
			}
			symbolsSeen[rec.Symbol] = true
		}
	}
	if err := s.assetRepo.UpsertPriceRecords(ctx, records); err != nil {
		return 0, err
	}
	return len(records), nil
}

func (s *ingestionSvc) ListAssets(ctx context.Context) ([]domain.Asset, error) {
	return s.assetRepo.ListAssets(ctx)
}

func (s *ingestionSvc) GetAssetPrices(ctx context.Context, symbol string, limit int) ([]domain.PriceRecord, error) {
	return s.assetRepo.GetPriceRecords(ctx, symbol, limit)
}

func (s *ingestionSvc) DeleteAsset(ctx context.Context, symbol string) error {
	return s.assetRepo.DeleteAsset(ctx, symbol)
}
