package publish

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"exchange-data-platform/internal/connector/api"
	modelconfig "exchange-data-platform/internal/model/config"
	"exchange-data-platform/internal/observability/metrics"
	qualitypipe "exchange-data-platform/internal/pipeline/quality"
	"exchange-data-platform/internal/storage/checkpoint"
	"exchange-data-platform/internal/storage/manifest"
	"exchange-data-platform/internal/storage/parquet"
	qualitystore "exchange-data-platform/internal/storage/quality"
	"exchange-data-platform/internal/storage/rejects"
	"exchange-data-platform/internal/storage/temp"
)

type Service struct {
	Connector       api.Connector
	Config          modelconfig.AppConfig
	TempStore       temp.Store
	ParquetWriter   parquet.Writer
	ManifestStore   manifest.Store
	CheckpointStore checkpoint.Store
	QualityStore    qualitystore.Store
	RejectStore     rejects.Store
}

var pathRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func (s Service) RunOnce(ctx context.Context) error {
	for _, market := range s.Config.Markets {
		if !pathRegex.MatchString(market) {
			return fmt.Errorf("invalid market name: %s", market)
		}
		for _, dataset := range s.Config.Datasets {
			if !pathRegex.MatchString(dataset) {
				return fmt.Errorf("invalid dataset name: %s", dataset)
			}
			jobID := fmt.Sprintf("%s-%s-%d", s.Config.Exchange, dataset, time.Now().UTC().UnixNano())
			rows, err := s.Connector.Fetch(ctx, api.FetchRequest{
				Dataset: dataset,
				Market:  market,
				Limit:   s.Config.MaxBatchSize,
				JobID:   jobID,
			})
			if err != nil {
				metrics.IncRejects()
				_, _ = s.RejectStore.Write(s.Config.Exchange, dataset, map[string]any{"market": market}, err)
				continue
			}
			if len(rows) == 0 {
				continue
			}
			tempPath, err := s.TempStore.Write(s.Config.Exchange, dataset, jobID, rows)
			if err != nil {
				return fmt.Errorf("write temp spool: %w", err)
			}
			parquetPath, err := s.ParquetWriter.WriteStandardized(s.Config.Exchange, s.Config.Region, market, dataset, jobID, rows)
			if err != nil {
				return fmt.Errorf("write parquet: %w", err)
			}
			if _, err := s.ManifestStore.Write(manifest.Entry{
				Exchange:      s.Config.Exchange,
				Region:        s.Config.Region,
				Market:        market,
				Dataset:       dataset,
				JobID:         jobID,
				RecordCount:   len(rows),
				ParquetPath:   parquetPath,
				TempPath:      tempPath,
				CreatedAt:     time.Now().UTC(),
				SchemaVersion: s.Config.SchemaVersion,
			}); err != nil {
				return fmt.Errorf("write manifest: %w", err)
			}
			if _, err := s.CheckpointStore.Save(checkpoint.Record{
				Exchange:    s.Config.Exchange,
				Dataset:     dataset,
				JobID:       jobID,
				LastRunAt:   time.Now().UTC(),
				RecordCount: len(rows),
			}); err != nil {
				return fmt.Errorf("write checkpoint: %w", err)
			}
			report := qualitypipe.Analyze(s.Config.Exchange, s.Config.Region, market, dataset, jobID, rows)
			report.CreatedAt = time.Now().UTC()
			if _, err := s.QualityStore.Write(report); err != nil {
				return fmt.Errorf("write quality report: %w", err)
			}
			if s.Config.EnableCleanup {
				if err := s.TempStore.Remove(tempPath); err != nil {
					return fmt.Errorf("cleanup temp spool: %w", err)
				}
			}
			metrics.IncBatches()
			metrics.AddRecords(len(rows))
		}
	}
	return nil
}
