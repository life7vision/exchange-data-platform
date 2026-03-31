package unit

import (
	"os"
	"path/filepath"
	"testing"

	loader "exchange-data-platform/internal/config"
)

func TestLoadConfigDefaults(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "config.yml")
	content := []byte("exchange: binance\nregion: global\nbase_url: https://api.binance.com\nlake_root: ./lake\ntemp_root: ./lake/tmp\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := loader.Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.SchemaVersion != "v1" {
		t.Fatalf("expected default schema version, got %q", cfg.SchemaVersion)
	}
	if len(cfg.DefaultSymbols) == 0 {
		t.Fatal("expected default symbols")
	}
}
