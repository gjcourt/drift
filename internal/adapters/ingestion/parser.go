package ingestion

import (
	"io"

	"github.com/gjcourt/drift/internal/domain"
)

// Parser implements outbound.CSVParser backed by ParseCSV.
type Parser struct{}

// ParseCSV parses a CSV price-data stream via the ingestion adapter.
func (Parser) ParseCSV(r io.Reader, filename string) ([]domain.PriceRecord, error) {
	return ParseCSV(r, filename)
}
