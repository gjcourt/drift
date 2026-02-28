package domain

import (
	"math"
	"testing"
)

func TestTotalWeight(t *testing.T) {
	tests := []struct {
		name   string
		assets []PortfolioAsset
		want   float64
	}{
		{"empty portfolio", nil, 0},
		{"single asset full weight", []PortfolioAsset{{Symbol: "AAPL", Weight: 1.0}}, 1.0},
		{"two equal assets", []PortfolioAsset{
			{Symbol: "AAPL", Weight: 0.5},
			{Symbol: "GOOG", Weight: 0.5},
		}, 1.0},
		{"three assets", []PortfolioAsset{
			{Symbol: "AAPL", Weight: 0.4},
			{Symbol: "GOOG", Weight: 0.35},
			{Symbol: "MSFT", Weight: 0.25},
		}, 1.0},
		{"under-weighted", []PortfolioAsset{
			{Symbol: "AAPL", Weight: 0.3},
			{Symbol: "GOOG", Weight: 0.3},
		}, 0.6},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Portfolio{Assets: tc.assets}
			if got := p.TotalWeight(); math.Abs(got-tc.want) > 1e-9 {
				t.Errorf("TotalWeight() = %v, want %v", got, tc.want)
			}
		})
	}
}
