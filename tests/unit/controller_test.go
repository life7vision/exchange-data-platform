package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"exchange-data-platform/internal/app"
	"exchange-data-platform/internal/connector/api"
	modelconfig "exchange-data-platform/internal/model/config"
	"exchange-data-platform/internal/model/raw"
	"exchange-data-platform/internal/observability/health"
	"exchange-data-platform/internal/pipeline/publish"
	"exchange-data-platform/internal/storage/checkpoint"
	"exchange-data-platform/internal/storage/manifest"
	"exchange-data-platform/internal/storage/parquet"
	qualitystore "exchange-data-platform/internal/storage/quality"
	"exchange-data-platform/internal/storage/rejects"
	"exchange-data-platform/internal/storage/temp"
)

type controllerStubConnector struct{}

func (s controllerStubConnector) Name() string   { return "stub" }
func (s controllerStubConnector) Region() string { return "global" }
func (s controllerStubConnector) Fetch(context.Context, api.FetchRequest) ([]raw.Envelope, error) {
	return []raw.Envelope{{
		Exchange:        "stub",
		Region:          "global",
		Market:          "spot",
		Dataset:         "trades",
		Symbol:          "BTCUSDT",
		EventTime:       time.Now().UTC(),
		IngestTime:      time.Now().UTC(),
		JobID:           "job-1",
		ProducerVersion: "0.1.0",
		Payload:         map[string]any{"ok": true},
	}}, nil
}

func TestRunOnceHandler(t *testing.T) {
	root := t.TempDir()
	cfg := modelconfig.AppConfig{
		Exchange:      "stub",
		Region:        "global",
		LakeRoot:      root,
		TempRoot:      filepath.Join(root, "tmp"),
		Markets:       []string{"spot"},
		Datasets:      []string{"trades"},
		SchemaVersion: "v1",
		EnableCleanup: true,
	}
	svc := publish.Service{
		Connector:       controllerStubConnector{},
		Config:          cfg,
		TempStore:       temp.Store{Root: cfg.TempRoot},
		ParquetWriter:   parquet.Writer{Root: cfg.LakeRoot, SchemaVersion: cfg.SchemaVersion},
		ManifestStore:   manifest.Store{Root: filepath.Join(cfg.LakeRoot, "manifests")},
		CheckpointStore: checkpoint.Store{Root: filepath.Join(cfg.LakeRoot, "checkpoints")},
		QualityStore:    qualitystore.Store{Root: filepath.Join(cfg.LakeRoot, "quality")},
		RejectStore:     rejects.Store{Root: filepath.Join(cfg.LakeRoot, "rejects")},
	}
	controller := app.NewController(cfg, svc, health.NewState("stub", time.Minute))
	body, _ := json.Marshal(app.RunRequest{Datasets: []string{"trades"}, Markets: []string{"spot"}})
	req := httptest.NewRequest(http.MethodPost, "/run-once", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	controller.RunOnceHandler(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
}
