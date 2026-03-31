package config

import (
	"fmt"
	"os"
	"path/filepath"

	modelconfig "exchange-data-platform/internal/model/config"
	"gopkg.in/yaml.v3"
)

func LoadJobs(dir string) ([]modelconfig.JobConfig, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read jobs dir: %w", err)
	}
	jobs := make([]modelconfig.JobConfig, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read job file %s: %w", path, err)
		}
		var cfg modelconfig.JobConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("unmarshal job file %s: %w", path, err)
		}
		if cfg.Enabled {
			jobs = append(jobs, cfg)
		}
	}
	return jobs, nil
}
