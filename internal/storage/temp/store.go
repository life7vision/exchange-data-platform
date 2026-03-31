package temp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"exchange-data-platform/internal/model/raw"
)

type Store struct {
	Root string
}

func (s Store) Write(exchange, dataset, jobID string, rows []raw.Envelope) (string, error) {
	dir := filepath.Join(s.Root, exchange, dataset, jobID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir temp dir: %w", err)
	}
	path := filepath.Join(dir, "raw.jsonl")
	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, row := range rows {
		if err := enc.Encode(row); err != nil {
			return "", fmt.Errorf("encode temp row: %w", err)
		}
	}
	return path, f.Sync()
}

func (s Store) Remove(path string) error {
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	_ = os.Remove(filepath.Dir(path))
	return nil
}
