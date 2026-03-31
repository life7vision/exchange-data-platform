package quality

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Report struct {
	Exchange         string    `json:"exchange"`
	Region           string    `json:"region"`
	Market           string    `json:"market"`
	Dataset          string    `json:"dataset"`
	JobID            string    `json:"job_id"`
	RecordCount      int       `json:"record_count"`
	DuplicateSymbols int       `json:"duplicate_symbols"`
	EmptyPayloads    int       `json:"empty_payloads"`
	CreatedAt        time.Time `json:"created_at"`
}

type Store struct {
	Root string
}

func (s Store) Write(report Report) (string, error) {
	dir := filepath.Join(s.Root, report.Exchange, report.Dataset)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir quality dir: %w", err)
	}
	path := filepath.Join(dir, fmt.Sprintf("%s.json", report.JobID))
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal quality report: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write quality report: %w", err)
	}
	return path, nil
}
