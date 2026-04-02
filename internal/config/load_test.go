package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfigDefaults(t *testing.T) {
	// Create a temporary config file with minimal data
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yml")

	content := `
exchange: binance
region: global
base_url: https://api.binance.com
lake_root: ./lake
temp_root: ./lake/tmp
datasets:
  - instruments
markets:
  - spot
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify defaults were set
	if cfg.SchemaVersion != "v1" {
		t.Errorf("expected SchemaVersion 'v1', got %q", cfg.SchemaVersion)
	}
	if cfg.ProducerVersion != "0.1.0" {
		t.Errorf("expected ProducerVersion '0.1.0', got %q", cfg.ProducerVersion)
	}
	if cfg.PollInterval != time.Minute {
		t.Errorf("expected PollInterval 1m, got %v", cfg.PollInterval)
	}
	if cfg.HTTPTimeout != 15*time.Second {
		t.Errorf("expected HTTPTimeout 15s, got %v", cfg.HTTPTimeout)
	}
	if cfg.MaxBatchSize != 200 {
		t.Errorf("expected MaxBatchSize 200, got %d", cfg.MaxBatchSize)
	}
	if len(cfg.DefaultSymbols) == 0 || cfg.DefaultSymbols[0] != "BTCUSDT" {
		t.Errorf("expected DefaultSymbols to include BTCUSDT")
	}
	if cfg.RetryBackoff != 2*time.Second {
		t.Errorf("expected RetryBackoff 2s, got %v", cfg.RetryBackoff)
	}
	if cfg.HealthMaxAge != 5*time.Minute {
		t.Errorf("expected HealthMaxAge 5m, got %v", cfg.HealthMaxAge)
	}
}

func TestLoadConfigWithCustomValues(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yml")

	content := `
exchange: bybit
region: us
base_url: https://api.bybit.com
lake_root: /data/lake
temp_root: /data/temp
poll_interval: 5m
http_timeout: 30s
max_batch_size: 500
max_retries: 5
retry_backoff: 5s
health_max_age: 10m
schema_version: v2
producer_version: 1.0.0
default_symbols:
  - ETHUSDT
  - BTCUSDT
datasets:
  - tickers
  - trades
markets:
  - spot
  - linear_perpetual
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify custom values were loaded
	if cfg.Exchange != "bybit" {
		t.Errorf("expected Exchange 'bybit', got %q", cfg.Exchange)
	}
	if cfg.Region != "us" {
		t.Errorf("expected Region 'us', got %q", cfg.Region)
	}
	if cfg.PollInterval != 5*time.Minute {
		t.Errorf("expected PollInterval 5m, got %v", cfg.PollInterval)
	}
	if cfg.HTTPTimeout != 30*time.Second {
		t.Errorf("expected HTTPTimeout 30s, got %v", cfg.HTTPTimeout)
	}
	if cfg.MaxBatchSize != 500 {
		t.Errorf("expected MaxBatchSize 500, got %d", cfg.MaxBatchSize)
	}
	if cfg.MaxRetries != 5 {
		t.Errorf("expected MaxRetries 5, got %d", cfg.MaxRetries)
	}
	if cfg.RetryBackoff != 5*time.Second {
		t.Errorf("expected RetryBackoff 5s, got %v", cfg.RetryBackoff)
	}
	if cfg.HealthMaxAge != 10*time.Minute {
		t.Errorf("expected HealthMaxAge 10m, got %v", cfg.HealthMaxAge)
	}
	if cfg.SchemaVersion != "v2" {
		t.Errorf("expected SchemaVersion 'v2', got %q", cfg.SchemaVersion)
	}
	if cfg.ProducerVersion != "1.0.0" {
		t.Errorf("expected ProducerVersion '1.0.0', got %q", cfg.ProducerVersion)
	}
	if len(cfg.DefaultSymbols) != 2 {
		t.Errorf("expected 2 default symbols, got %d", len(cfg.DefaultSymbols))
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yml")

	content := `
invalid: [yaml: content:
  - without: proper: structure
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadConfigNegativeDurations(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yml")

	content := `
exchange: binance
region: global
base_url: https://api.binance.com
lake_root: ./lake
temp_root: ./lake/tmp
poll_interval: -1s
http_timeout: -1s
max_batch_size: 100
max_retries: -5
retry_backoff: -1s
health_max_age: -1s
datasets:
  - instruments
markets:
  - spot
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify negative values are corrected with defaults
	if cfg.PollInterval != time.Minute {
		t.Errorf("negative PollInterval should default to 1m, got %v", cfg.PollInterval)
	}
	if cfg.HTTPTimeout != 15*time.Second {
		t.Errorf("negative HTTPTimeout should default to 15s, got %v", cfg.HTTPTimeout)
	}
	if cfg.RetryBackoff != 2*time.Second {
		t.Errorf("negative RetryBackoff should default to 2s, got %v", cfg.RetryBackoff)
	}
	if cfg.HealthMaxAge != 5*time.Minute {
		t.Errorf("negative HealthMaxAge should default to 5m, got %v", cfg.HealthMaxAge)
	}
	// MaxRetries with negative is clamped to 0, not set to default
	if cfg.MaxRetries != 0 {
		t.Errorf("negative MaxRetries should be clamped to 0, got %d", cfg.MaxRetries)
	}
}

func TestLoadConfigEmptyDatasets(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yml")

	content := `
exchange: binance
region: global
base_url: https://api.binance.com
lake_root: ./lake
temp_root: ./lake/tmp
datasets: []
markets: []
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Datasets) > 0 {
		t.Errorf("expected empty datasets, got %v", cfg.Datasets)
	}
	if len(cfg.Markets) > 0 {
		t.Errorf("expected empty markets, got %v", cfg.Markets)
	}
}
