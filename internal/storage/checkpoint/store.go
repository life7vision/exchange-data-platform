package checkpoint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Record struct {
	Exchange    string    `json:"exchange"`
	Dataset     string    `json:"dataset"`
	JobID       string    `json:"job_id"`
	LastRunAt   time.Time `json:"last_run_at"`
	RecordCount int       `json:"record_count"`
}

type Store struct {
	Root string
}

func (s Store) Save(r Record) (string, error) {
	dir := filepath.Join(s.Root, r.Exchange)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir checkpoint dir: %w", err)
	}
	path := filepath.Join(dir, fmt.Sprintf("%s.json", r.Dataset))
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal checkpoint: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write checkpoint: %w", err)
	}
	return path, nil
}
