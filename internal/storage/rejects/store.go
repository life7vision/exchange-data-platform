package rejects

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Store struct {
	Root string
}

func (s Store) Write(exchange, dataset string, payload any, cause error) (string, error) {
	dir := filepath.Join(s.Root, exchange, dataset)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir reject dir: %w", err)
	}
	path := filepath.Join(dir, fmt.Sprintf("%d.json", time.Now().UnixNano()))
	body := map[string]any{
		"cause":      cause.Error(),
		"payload":    payload,
		"created_at": time.Now().UTC(),
	}
	data, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return "", err
	}
	return path, os.WriteFile(path, data, 0o644)
}
