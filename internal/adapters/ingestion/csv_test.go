package ingestion

import (
	"strings"
	"testing"
)

const singleSymbolCSV = `date,open,high,low,close,adjusted_close,volume
2024-01-02,150.0,155.0,149.0,153.0,153.0,1000000
2024-01-03,153.0,156.0,151.0,154.5,154.5,900000
2024-01-04,154.5,158.0,153.0,157.0,157.0,1100000
`

const multiSymbolCSV = `date,symbol,open,high,low,close,adjusted_close,volume
2024-01-02,AAPL,150.0,155.0,149.0,153.0,153.0,1000000
2024-01-02,GOOG,140.0,145.0,139.0,143.0,143.0,500000
2024-01-03,AAPL,153.0,156.0,151.0,154.5,154.5,900000
2024-01-03,GOOG,143.0,148.0,142.0,147.0,147.0,600000
`

const missingAdjCloseCSV = `date,open,high,low,close,adjusted_close,volume
2024-01-02,150.0,155.0,149.0,153.0,,1000000
2024-01-03,153.0,156.0,151.0,154.5,154.5,900000
`

func TestParseCSVSingleSymbol(t *testing.T) {
	recs, err := ParseCSV(strings.NewReader(singleSymbolCSV), "AAPL.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) != 3 {
		t.Fatalf("got %d records, want 3", len(recs))
	}
	for _, r := range recs {
		if r.Symbol != "AAPL" {
			t.Errorf("symbol = %q, want AAPL", r.Symbol)
		}
		if r.AdjustedClose <= 0 {
			t.Errorf("AdjustedClose = %v, want > 0", r.AdjustedClose)
		}
	}
}

func TestParseCSVMultiSymbol(t *testing.T) {
	recs, err := ParseCSV(strings.NewReader(multiSymbolCSV), "prices.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) != 4 {
		t.Fatalf("got %d records, want 4", len(recs))
	}
	symbols := make(map[string]int)
	for _, r := range recs {
		symbols[r.Symbol]++
	}
	if symbols["AAPL"] != 2 || symbols["GOOG"] != 2 {
		t.Errorf("unexpected symbol counts: %v", symbols)
	}
}

func TestParseCSVSkipsMissingAdjClose(t *testing.T) {
	recs, err := ParseCSV(strings.NewReader(missingAdjCloseCSV), "TEST.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("got %d records, want 1 (row with missing adj_close skipped)", len(recs))
	}
}

func TestParseCSVEmptyFile(t *testing.T) {
	_, err := ParseCSV(strings.NewReader(""), "TEST.csv")
	if err == nil {
		t.Error("expected error for empty file, got nil")
	}
}
