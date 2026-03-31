package config

import (
	"fmt"
	"os"
	"time"

	modelconfig "exchange-data-platform/internal/model/config"
	"gopkg.in/yaml.v3"
)

func Load(path string) (modelconfig.AppConfig, error) {
	var cfg modelconfig.AppConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal config: %w", err)
	}
	if cfg.SchemaVersion == "" {
		cfg.SchemaVersion = "v1"
	}
	if cfg.ProducerVersion == "" {
		cfg.ProducerVersion = "0.1.0"
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = time.Minute
	}
	if cfg.HTTPTimeout <= 0 {
		cfg.HTTPTimeout = 15 * time.Second
	}
	if cfg.MaxBatchSize <= 0 {
		cfg.MaxBatchSize = 200
	}
	if len(cfg.DefaultSymbols) == 0 {
		cfg.DefaultSymbols = []string{"BTCUSDT"}
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = 2 * time.Second
	}
	if cfg.HealthMaxAge <= 0 {
		cfg.HealthMaxAge = 5 * time.Minute
	}
	return cfg, nil
}
