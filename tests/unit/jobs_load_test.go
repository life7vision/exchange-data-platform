package unit

import (
	"os"
	"path/filepath"
	"testing"

	loader "exchange-data-platform/internal/config"
)

func TestLoadJobs(t *testing.T) {
	root := t.TempDir()
	err := os.WriteFile(filepath.Join(root, "job.yml"), []byte("name: sample\nexchange: binance\nenabled: true\nmode: backfill\ninterval: 1m\n"), 0o644)
	if err != nil {
		t.Fatalf("write job config: %v", err)
	}
	jobs, err := loader.LoadJobs(root)
	if err != nil {
		t.Fatalf("load jobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Name != "sample" {
		t.Fatalf("unexpected job name %q", jobs[0].Name)
	}
}
