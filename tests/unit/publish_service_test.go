package unit

import (
	"context"
	"io/fs"
	"path/filepath"
	"testing"
	"time"

	"exchange-data-platform/internal/connector/api"
	modelconfig "exchange-data-platform/internal/model/config"
	"exchange-data-platform/internal/model/raw"
	"exchange-data-platform/internal/pipeline/publish"
	"exchange-data-platform/internal/storage/checkpoint"
	"exchange-data-platform/internal/storage/manifest"
	"exchange-data-platform/internal/storage/parquet"
	qualitystore "exchange-data-platform/internal/storage/quality"
	"exchange-data-platform/internal/storage/rejects"
	"exchange-data-platform/internal/storage/temp"
)

type stubConnector struct{}

func (s stubConnector) Name() string   { return "stub" }
func (s stubConnector) Region() string { return "global" }
func (s stubConnector) Fetch(context.Context, api.FetchRequest) ([]raw.Envelope, error) {
	return []raw.Envelope{{
		Exchange:        "stub",
		Region:          "global",
		Market:          "spot",
		Dataset:         "tickers",
		Symbol:          "BTCUSDT",
		EventTime:       time.Now().UTC(),
		IngestTime:      time.Now().UTC(),
		JobID:           "job-1",
		ProducerVersion: "0.1.0",
		Payload:         map[string]any{"price": "1"},
	}}, nil
}

func TestPublishServiceCleansTempAfterParquetWrite(t *testing.T) {
	root := t.TempDir()
	cfg := modelconfig.AppConfig{
		Exchange:      "stub",
		Region:        "global",
		LakeRoot:      root,
		TempRoot:      filepath.Join(root, "tmp"),
		Markets:       []string{"spot"},
		Datasets:      []string{"tickers"},
		SchemaVersion: "v1",
		EnableCleanup: true,
	}
	svc := publish.Service{
		Connector:       stubConnector{},
		Config:          cfg,
		TempStore:       temp.Store{Root: cfg.TempRoot},
		ParquetWriter:   parquet.Writer{Root: cfg.LakeRoot, SchemaVersion: cfg.SchemaVersion},
		ManifestStore:   manifest.Store{Root: filepath.Join(cfg.LakeRoot, "manifests")},
		CheckpointStore: checkpoint.Store{Root: filepath.Join(cfg.LakeRoot, "checkpoints")},
		QualityStore:    qualitystore.Store{Root: filepath.Join(cfg.LakeRoot, "quality")},
		RejectStore:     rejects.Store{Root: filepath.Join(cfg.LakeRoot, "rejects")},
	}
	if err := svc.RunOnce(context.Background()); err != nil {
		t.Fatalf("run once: %v", err)
	}
	matches, err := filepath.Glob(filepath.Join(cfg.TempRoot, "*", "*", "*", "raw.jsonl"))
	if err != nil {
		t.Fatalf("glob temp: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected temp spool cleanup, found %v", matches)
	}
	parquetCount := 0
	err = filepath.WalkDir(cfg.LakeRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() && filepath.Ext(path) == ".parquet" {
			parquetCount++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk parquet tree: %v", err)
	}
	if parquetCount == 0 {
		t.Fatal("expected at least one parquet file")
	}
	qualityMatches, err := filepath.Glob(filepath.Join(cfg.LakeRoot, "quality", "*", "*", "*.json"))
	if err != nil {
		t.Fatalf("glob quality: %v", err)
	}
	if len(qualityMatches) == 0 {
		t.Fatal("expected quality report")
	}
}
