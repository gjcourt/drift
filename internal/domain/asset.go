// Package domain defines the core types and business rules for Drift.
// It has no external dependencies and must not import from other internal packages.
package domain

import "time"

// Asset represents a financial instrument identified by its ticker symbol.
type Asset struct {
	ID     string
	Symbol string
	Name   string
}

// PriceRecord holds a single OHLCV row for an asset on a given trading day.
type PriceRecord struct {
	AssetID       string
	Symbol        string
	Date          time.Time
	Open          float64
	High          float64
	Low           float64
	Close         float64
	Volume        int64
	AdjustedClose float64
}
