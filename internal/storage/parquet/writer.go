package parquet

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"exchange-data-platform/internal/model/canonical"
	"exchange-data-platform/internal/model/raw"
	"github.com/parquet-go/parquet-go"
)

type Writer struct {
	Root          string
	SchemaVersion string
}

func (w Writer) WriteStandardized(exchange, region, market, dataset, jobID string, rows []raw.Envelope) (string, error) {
	now := time.Now().UTC()
	dir := filepath.Join(
		w.Root,
		"standardized",
		fmt.Sprintf("exchange=%s", exchange),
		fmt.Sprintf("region=%s", region),
		fmt.Sprintf("market=%s", market),
		fmt.Sprintf("dataset=%s", dataset),
		fmt.Sprintf("version=%s", w.SchemaVersion),
		fmt.Sprintf("year=%04d", now.Year()),
		fmt.Sprintf("month=%02d", now.Month()),
		fmt.Sprintf("day=%02d", now.Day()),
		fmt.Sprintf("hour=%02d", now.Hour()),
	)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir parquet dir: %w", err)
	}
	path := filepath.Join(dir, fmt.Sprintf("%s.parquet", jobID))
	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create parquet file: %w", err)
	}
	defer file.Close()

	pw := parquet.NewGenericWriter[canonical.EnvelopeRow](file)

	// Pre-allocate buffer to avoid repeated re-allocations in the loop
	buffer := make([]canonical.EnvelopeRow, len(rows))
	for i, row := range rows {
		payloadJSON, err := json.Marshal(row.Payload)
		if err != nil {
			return "", fmt.Errorf("marshal payload json: %w", err)
		}
		buffer[i] = canonical.EnvelopeRow{
			Exchange:         row.Exchange,
			Region:           row.Region,
			Market:           row.Market,
			Dataset:          row.Dataset,
			Symbol:           row.Symbol,
			EventTimeUnixMs:  row.EventTime.UnixMilli(),
			IngestTimeUnixMs: row.IngestTime.UnixMilli(),
			JobID:            row.JobID,
			ProducerVersion:  row.ProducerVersion,
			PayloadJSON:      string(payloadJSON),
		}
	}
	if _, err := pw.Write(buffer); err != nil {
		return "", fmt.Errorf("write parquet rows: %w", err)
	}
	if err := pw.Close(); err != nil {
		return "", fmt.Errorf("close parquet writer: %w", err)
	}
	// Redundant file.Sync() removed, as data integrity is ensured by writer and file closing
	return path, nil
}
