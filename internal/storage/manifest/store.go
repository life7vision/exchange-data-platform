package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	Exchange      string    `json:"exchange"`
	Region        string    `json:"region"`
	Market        string    `json:"market"`
	Dataset       string    `json:"dataset"`
	JobID         string    `json:"job_id"`
	RecordCount   int       `json:"record_count"`
	ParquetPath   string    `json:"parquet_path"`
	TempPath      string    `json:"temp_path"`
	CreatedAt     time.Time `json:"created_at"`
	SchemaVersion string    `json:"schema_version"`
}

type Store struct {
	Root string
}

func (s Store) Write(entry Entry) (string, error) {
	dir := filepath.Join(s.Root, entry.Exchange, entry.Dataset)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir manifest dir: %w", err)
	}
	path := filepath.Join(dir, fmt.Sprintf("%s.json", entry.JobID))
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal manifest: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write manifest: %w", err)
	}
	return path, nil
}
