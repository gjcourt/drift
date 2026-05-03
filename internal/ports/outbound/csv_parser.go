package outbound

import (
	"io"

	"github.com/gjcourt/drift/internal/domain"
)

// CSVParser parses a CSV price-data stream into domain price records.
type CSVParser interface {
	ParseCSV(r io.Reader, filename string) ([]domain.PriceRecord, error)
}
