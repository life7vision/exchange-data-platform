package quality

import (
	"exchange-data-platform/internal/model/raw"
	qualitystore "exchange-data-platform/internal/storage/quality"
)

func Analyze(exchange, region, market, dataset, jobID string, rows []raw.Envelope) qualitystore.Report {
	seen := map[string]int{}
	duplicates := 0
	emptyPayloads := 0

	for _, row := range rows {
		if row.Symbol != "" {
			seen[row.Symbol]++
			if seen[row.Symbol] == 2 {
				duplicates++
			}
		}
		if len(row.Payload) == 0 {
			emptyPayloads++
		}
	}

	return qualitystore.Report{
		Exchange:         exchange,
		Region:           region,
		Market:           market,
		Dataset:          dataset,
		JobID:            jobID,
		RecordCount:      len(rows),
		DuplicateSymbols: duplicates,
		EmptyPayloads:    emptyPayloads,
	}
}
