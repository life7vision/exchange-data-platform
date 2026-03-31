package config

import "time"

type AppConfig struct {
	Exchange        string        `yaml:"exchange"`
	Region          string        `yaml:"region"`
	Environment     string        `yaml:"environment"`
	BaseURL         string        `yaml:"base_url"`
	DerivativesURL  string        `yaml:"derivatives_url"`
	WebSocketURL    string        `yaml:"websocket_url"`
	LakeRoot        string        `yaml:"lake_root"`
	TempRoot        string        `yaml:"temp_root"`
	PollInterval    time.Duration `yaml:"poll_interval"`
	HTTPTimeout     time.Duration `yaml:"http_timeout"`
	MaxBatchSize    int           `yaml:"max_batch_size"`
	Datasets        []string      `yaml:"datasets"`
	Markets         []string      `yaml:"markets"`
	WorkerMode      string        `yaml:"worker_mode"`
	EnableCleanup   bool          `yaml:"enable_cleanup"`
	SchemaVersion   string        `yaml:"schema_version"`
	ProducerVersion string        `yaml:"producer_version"`
	DefaultSymbols  []string      `yaml:"default_symbols"`
	MaxRetries      int           `yaml:"max_retries"`
	RetryBackoff    time.Duration `yaml:"retry_backoff"`
	HealthMaxAge    time.Duration `yaml:"health_max_age"`
}
