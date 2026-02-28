package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/gjcourt/drift/internal/domain"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(":memory:")
	if err != nil {
		t.Fatalf("new test store: %v", err)
	}
	return s
}

func TestUpsertAndGetAsset(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	asset := domain.Asset{Symbol: "AAPL", Name: "Apple Inc."}
	if err := s.UpsertAsset(ctx, asset); err != nil {
		t.Fatalf("UpsertAsset: %v", err)
	}

	got, err := s.GetAsset(ctx, "AAPL")
	if err != nil {
		t.Fatalf("GetAsset: %v", err)
	}
	if got.Symbol != "AAPL" || got.Name != "Apple Inc." {
		t.Errorf("got %+v, want AAPL / Apple Inc.", got)
	}
}

func TestUpsertAssetUpdatesName(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_ = s.UpsertAsset(ctx, domain.Asset{Symbol: "AAPL", Name: "Old Name"})
	_ = s.UpsertAsset(ctx, domain.Asset{Symbol: "AAPL", Name: "New Name"})

	got, err := s.GetAsset(ctx, "AAPL")
	if err != nil {
		t.Fatalf("GetAsset: %v", err)
	}
	if got.Name != "New Name" {
		t.Errorf("Name = %q, want 'New Name'", got.Name)
	}
}

func TestListAssets(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_ = s.UpsertAsset(ctx, domain.Asset{Symbol: "MSFT"})
	_ = s.UpsertAsset(ctx, domain.Asset{Symbol: "AAPL"})

	assets, err := s.ListAssets(ctx)
	if err != nil {
		t.Fatalf("ListAssets: %v", err)
	}
	if len(assets) != 2 {
		t.Fatalf("got %d assets, want 2", len(assets))
	}
	if assets[0].Symbol != "AAPL" || assets[1].Symbol != "MSFT" {
		t.Errorf("unexpected order: %v %v", assets[0].Symbol, assets[1].Symbol)
	}
}

func TestDeleteAsset(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_ = s.UpsertAsset(ctx, domain.Asset{Symbol: "AAPL"})
	_ = s.UpsertPriceRecords(ctx, []domain.PriceRecord{
		{Symbol: "AAPL", Date: time.Now(), AdjustedClose: 150},
	})

	if err := s.DeleteAsset(ctx, "AAPL"); err != nil {
		t.Fatalf("DeleteAsset: %v", err)
	}

	assets, _ := s.ListAssets(ctx)
	if len(assets) != 0 {
		t.Errorf("asset count = %d, want 0 after delete", len(assets))
	}
}

func TestUpsertAndGetPriceRecords(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	records := []domain.PriceRecord{
		{Symbol: "AAPL", Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), AdjustedClose: 150},
		{Symbol: "AAPL", Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), AdjustedClose: 152},
		{Symbol: "AAPL", Date: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), AdjustedClose: 149},
	}
	if err := s.UpsertPriceRecords(ctx, records); err != nil {
		t.Fatalf("UpsertPriceRecords: %v", err)
	}

	got, err := s.GetPriceRecords(ctx, "AAPL", 10)
	if err != nil {
		t.Fatalf("GetPriceRecords: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d records, want 3", len(got))
	}
}

func TestGetPriceRecordsLimit(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	records := make([]domain.PriceRecord, 10)
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range records {
		records[i] = domain.PriceRecord{
			Symbol:        "GOOG",
			Date:          base.AddDate(0, 0, i),
			AdjustedClose: float64(100 + i),
		}
	}
	_ = s.UpsertPriceRecords(ctx, records)

	got, err := s.GetPriceRecords(ctx, "GOOG", 5)
	if err != nil {
		t.Fatalf("GetPriceRecords: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("limit 5: got %d records", len(got))
	}
}

func TestExperimentRoundTrip(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	exp := domain.Experiment{
		ID:          "exp-001",
		Name:        "Test Experiment",
		Description: "round trip test",
		Portfolio: domain.Portfolio{
			Assets:    []domain.PortfolioAsset{{Symbol: "AAPL", Weight: 1.0}},
			Rebalance: domain.RebalanceAnnual,
		},
		Config: domain.SimulationConfig{
			Model:        domain.ModelGBM,
			NumPaths:     100,
			HorizonDays:  252,
			LookbackDays: 756,
			StartValue:   100_000,
		},
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		UpdatedAt: time.Now().UTC().Truncate(time.Second),
	}

	if err := s.SaveExperiment(ctx, exp); err != nil {
		t.Fatalf("SaveExperiment: %v", err)
	}

	got, err := s.GetExperiment(ctx, "exp-001")
	if err != nil {
		t.Fatalf("GetExperiment: %v", err)
	}
	if got.Name != exp.Name {
		t.Errorf("Name = %q, want %q", got.Name, exp.Name)
	}
	if got.Config.NumPaths != exp.Config.NumPaths {
		t.Errorf("NumPaths = %d, want %d", got.Config.NumPaths, exp.Config.NumPaths)
	}
	if len(got.Portfolio.Assets) != 1 || got.Portfolio.Assets[0].Symbol != "AAPL" {
		t.Errorf("Portfolio assets mismatch: %+v", got.Portfolio.Assets)
	}
}

func TestListExperiments(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	for _, id := range []string{"exp-a", "exp-b", "exp-c"} {
		_ = s.SaveExperiment(ctx, domain.Experiment{
			ID:        id,
			Name:      id,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Config:    domain.SimulationConfig{Model: domain.ModelGBM},
		})
	}

	list, err := s.ListExperiments(ctx)
	if err != nil {
		t.Fatalf("ListExperiments: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("got %d experiments, want 3", len(list))
	}
}

func TestRunRoundTrip(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	run := domain.Run{
		ID:           "run-001",
		ExperimentID: "exp-001",
		StartedAt:    time.Now().UTC().Truncate(time.Second),
		Status:       domain.StatusComplete,
		Stats: domain.ResultStats{
			P50:  120_000,
			Mean: 125_000,
		},
	}
	if err := s.SaveRun(ctx, run); err != nil {
		t.Fatalf("SaveRun: %v", err)
	}

	got, err := s.GetRun(ctx, "run-001")
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if got.Status != domain.StatusComplete {
		t.Errorf("Status = %q, want complete", got.Status)
	}
	if got.Stats.P50 != run.Stats.P50 {
		t.Errorf("Stats.P50 = %v, want %v", got.Stats.P50, run.Stats.P50)
	}
}
