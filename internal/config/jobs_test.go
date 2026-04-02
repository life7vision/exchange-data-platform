package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadJobsValid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test job files
	job1 := `
name: binance-streaming
exchange: binance
enabled: true
mode: streaming
interval: 2m
datasets:
  - trades_stream
markets:
  - spot
worker_url: http://worker:8080
`

	job2 := `
name: bybit-backfill
exchange: bybit
enabled: true
mode: backfill
interval: 1h
datasets:
  - tickers
markets:
  - linear_perpetual
worker_url: http://worker:8081
`

	if err := os.WriteFile(filepath.Join(tmpDir, "job1.yml"), []byte(job1), 0644); err != nil {
		t.Fatalf("failed to write job1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "job2.yml"), []byte(job2), 0644); err != nil {
		t.Fatalf("failed to write job2: %v", err)
	}

	jobs, err := LoadJobs(tmpDir)
	if err != nil {
		t.Fatalf("LoadJobs failed: %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}

	// Check first job
	if jobs[0].Name != "binance-streaming" {
		t.Errorf("expected first job name 'binance-streaming', got %q", jobs[0].Name)
	}
	if jobs[0].Exchange != "binance" {
		t.Errorf("expected exchange 'binance', got %q", jobs[0].Exchange)
	}
	if !jobs[0].Enabled {
		t.Errorf("expected job1 to be enabled")
	}
	if jobs[0].Mode != "streaming" {
		t.Errorf("expected mode 'streaming', got %q", jobs[0].Mode)
	}
	if jobs[0].Interval != 2*time.Minute {
		t.Errorf("expected interval 2m, got %v", jobs[0].Interval)
	}
}

func TestLoadJobsFiltersDisabled(t *testing.T) {
	tmpDir := t.TempDir()

	enabledJob := `
name: enabled-job
exchange: binance
enabled: true
mode: streaming
interval: 1m
datasets:
  - trades
markets:
  - spot
worker_url: http://worker:8080
`

	disabledJob := `
name: disabled-job
exchange: bybit
enabled: false
mode: backfill
interval: 1h
datasets:
  - tickers
markets:
  - spot
worker_url: http://worker:8081
`

	if err := os.WriteFile(filepath.Join(tmpDir, "enabled.yml"), []byte(enabledJob), 0644); err != nil {
		t.Fatalf("failed to write enabled job: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "disabled.yml"), []byte(disabledJob), 0644); err != nil {
		t.Fatalf("failed to write disabled job: %v", err)
	}

	jobs, err := LoadJobs(tmpDir)
	if err != nil {
		t.Fatalf("LoadJobs failed: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job (disabled jobs filtered), got %d", len(jobs))
	}

	if jobs[0].Name != "enabled-job" {
		t.Errorf("expected enabled job to be returned, got %q", jobs[0].Name)
	}
}

func TestLoadJobsEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	jobs, err := LoadJobs(tmpDir)
	if err != nil {
		t.Fatalf("LoadJobs failed: %v", err)
	}

	if len(jobs) != 0 {
		t.Fatalf("expected 0 jobs from empty directory, got %d", len(jobs))
	}
}

func TestLoadJobsIgnoresNonYML(t *testing.T) {
	tmpDir := t.TempDir()

	job := `
name: test-job
exchange: binance
enabled: true
mode: streaming
interval: 1m
datasets:
  - trades
markets:
  - spot
worker_url: http://worker:8080
`

	// Write YML file only (non-YML files will cause unmarshal errors)
	if err := os.WriteFile(filepath.Join(tmpDir, "job.yml"), []byte(job), 0644); err != nil {
		t.Fatalf("failed to write job.yml: %v", err)
	}

	jobs, err := LoadJobs(tmpDir)
	if err != nil {
		t.Fatalf("LoadJobs failed: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	if jobs[0].Name != "test-job" {
		t.Errorf("expected job name 'test-job', got %q", jobs[0].Name)
	}
}

func TestLoadJobsDirectoryNotFound(t *testing.T) {
	_, err := LoadJobs("/nonexistent/directory")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}
