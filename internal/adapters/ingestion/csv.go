// Package ingestion parses external data formats (CSV price files and JSON
// experiment configs) into domain objects.
package ingestion

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gjcourt/drift/internal/domain"
)

// ParseCSV parses single-symbol or multi-symbol CSV price files.
func ParseCSV(r io.Reader, filename string) ([]domain.PriceRecord, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	idx := buildIndex(headers)

	// Determine if multi-symbol (has "symbol" column) or single-symbol (filename = SYMBOL.csv).
	defaultSymbol := ""
	if _, ok := idx["symbol"]; !ok {
		// derive symbol from filename: AAPL.csv -> AAPL
		base := filename
		if dot := strings.LastIndex(base, "."); dot >= 0 {
			base = base[:dot]
		}
		defaultSymbol = strings.ToUpper(base)
	}

	var records []domain.PriceRecord
	lineNum := 1
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}
		lineNum++

		adjClose, err := parseFloat(row, idx, "adjusted_close")
		if err != nil || adjClose <= 0 {
			continue // skip rows missing adjusted_close
		}

		date, err := parseDate(row, idx)
		if err != nil {
			continue
		}

		symbol := defaultSymbol
		if i, ok := idx["symbol"]; ok && i < len(row) {
			symbol = strings.ToUpper(strings.TrimSpace(row[i]))
		}

		rec := domain.PriceRecord{
			Symbol:        symbol,
			Date:          date,
			AdjustedClose: adjClose,
		}
		rec.Open, _ = parseFloat(row, idx, "open")
		rec.High, _ = parseFloat(row, idx, "high")
		rec.Low, _ = parseFloat(row, idx, "low")
		rec.Close, _ = parseFloat(row, idx, "close")
		if i, ok := idx["volume"]; ok && i < len(row) {
			v, _ := strconv.ParseInt(strings.TrimSpace(row[i]), 10, 64)
			rec.Volume = v
		}
		records = append(records, rec)
	}
	return records, nil
}

func buildIndex(headers []string) map[string]int {
	m := make(map[string]int, len(headers))
	for i, h := range headers {
		m[strings.ToLower(strings.TrimSpace(h))] = i
	}
	return m
}

func parseFloat(row []string, idx map[string]int, col string) (float64, error) {
	i, ok := idx[col]
	if !ok || i >= len(row) {
		return 0, fmt.Errorf("column %q not found", col)
	}
	v := strings.TrimSpace(row[i])
	if v == "" {
		return 0, fmt.Errorf("empty value for %q", col)
	}
	return strconv.ParseFloat(v, 64)
}

func parseDate(row []string, idx map[string]int) (time.Time, error) {
	i, ok := idx["date"]
	if !ok || i >= len(row) {
		return time.Time{}, fmt.Errorf("date column missing")
	}
	return time.Parse("2006-01-02", strings.TrimSpace(row[i]))
}
