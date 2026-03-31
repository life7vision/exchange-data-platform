package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"exchange-data-platform/internal/app"
	modelconfig "exchange-data-platform/internal/model/config"
	"exchange-data-platform/internal/observability/health"
	"exchange-data-platform/internal/pipeline/publish"
	"exchange-data-platform/internal/storage/checkpoint"
	"exchange-data-platform/internal/storage/manifest"
	"exchange-data-platform/internal/storage/parquet"
	qualitystore "exchange-data-platform/internal/storage/quality"
	"exchange-data-platform/internal/storage/rejects"
	"exchange-data-platform/internal/storage/temp"
)

func TestSecurity_PathTraversal(t *testing.T) {
	root := t.TempDir()
	cfg := modelconfig.AppConfig{
		Exchange:      "stub",
		Region:        "global",
		LakeRoot:      root,
		TempRoot:      filepath.Join(root, "tmp"),
		Markets:       []string{"spot"},
		Datasets:      []string{"trades"},
		SchemaVersion: "v1",
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

	// Attempt path traversal via Datasets
	body, _ := json.Marshal(app.RunRequest{
		Datasets: []string{"../../evil"},
		Markets:  []string{"spot"},
	})
	req := httptest.NewRequest(http.MethodPost, "/run-once", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	controller.RunOnceHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for path traversal attempt, got %d", rec.Code)
	}
}

func TestSecurity_Authentication(t *testing.T) {
	os.Setenv("WORKER_API_TOKEN", "secret-token")
	defer os.Unsetenv("WORKER_API_TOKEN")

	root := t.TempDir()
	cfg := modelconfig.AppConfig{
		Exchange:      "stub",
		Region:        "global",
		LakeRoot:      root,
		TempRoot:      filepath.Join(root, "tmp"),
		Markets:       []string{"spot"},
		Datasets:      []string{"trades"},
		SchemaVersion: "v1",
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

	// 1. Request without token
	body, _ := json.Marshal(app.RunRequest{Datasets: []string{"trades"}, Markets: []string{"spot"}})
	req := httptest.NewRequest(http.MethodPost, "/run-once", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	controller.RunOnceHandler(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized without token, got %d", rec.Code)
	}

	// 2. Request with wrong token
	req = httptest.NewRequest(http.MethodPost, "/run-once", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec = httptest.NewRecorder()
	controller.RunOnceHandler(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized with wrong token, got %d", rec.Code)
	}

	// 3. Request with correct token
	req = httptest.NewRequest(http.MethodPost, "/run-once", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer secret-token")
	rec = httptest.NewRecorder()
	controller.RunOnceHandler(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202 Accepted with correct token, got %d", rec.Code)
	}
}
