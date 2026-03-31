package app

import (
	"context"
	"errors"
	"log/slog"
	"time"

	modelconfig "exchange-data-platform/internal/model/config"
	"exchange-data-platform/internal/observability/health"
	"exchange-data-platform/internal/pipeline/publish"
)

func RunWorker(ctx context.Context, cfg modelconfig.AppConfig, svc publish.Service, healthState *health.State) error {
	if cfg.WorkerMode == "once" {
		err := svc.RunOnce(ctx)
		if err != nil {
			healthState.MarkFailure(err)
			return err
		}
		healthState.MarkSuccess()
		return nil
	}
	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()

	for {
		if err := svc.RunOnce(ctx); err != nil {
			healthState.MarkFailure(err)
			slog.Error("worker run failed", "exchange", cfg.Exchange, "err", err)
		} else {
			healthState.MarkSuccess()
		}
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil
			}
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
