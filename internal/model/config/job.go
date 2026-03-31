package config

import "time"

type JobConfig struct {
	Name      string        `yaml:"name"`
	Exchange  string        `yaml:"exchange"`
	Enabled   bool          `yaml:"enabled"`
	Mode      string        `yaml:"mode"`
	Schedule  string        `yaml:"schedule"`
	Interval  time.Duration `yaml:"interval"`
	Datasets  []string      `yaml:"datasets"`
	Markets   []string      `yaml:"markets"`
	WorkerURL string        `yaml:"worker_url"`
}
