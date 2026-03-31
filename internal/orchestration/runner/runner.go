package runner

import (
	"context"
	"fmt"
	"path/filepath"

	"exchange-data-platform/internal/app"
	"exchange-data-platform/internal/connector/api"
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

func NewService(cfg modelconfig.AppConfig, connector api.Connector) (publish.Service, error) {
	if cfg.LakeRoot == "" {
		return publish.Service{}, fmt.Errorf("lake_root is required")
	}
	svc := publish.Service{
		Connector: connector,
		Config:    cfg,
		TempStore: temp.Store{Root: cfg.TempRoot},
		ParquetWriter: parquet.Writer{
			Root:          cfg.LakeRoot,
			SchemaVersion: cfg.SchemaVersion,
		},
		ManifestStore:   manifest.Store{Root: filepath.Join(cfg.LakeRoot, "manifests")},
		CheckpointStore: checkpoint.Store{Root: filepath.Join(cfg.LakeRoot, "checkpoints")},
		QualityStore:    qualitystore.Store{Root: filepath.Join(cfg.LakeRoot, "quality")},
		RejectStore:     rejects.Store{Root: filepath.Join(cfg.LakeRoot, "rejects")},
	}
	return svc, nil
}

func Start(ctx context.Context, cfg modelconfig.AppConfig, svc publish.Service, healthState *health.State) error {
	return app.RunWorker(ctx, cfg, svc, healthState)
}
