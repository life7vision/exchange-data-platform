package raw

import "time"

type Envelope struct {
	Exchange        string         `json:"exchange"`
	Region          string         `json:"region"`
	Market          string         `json:"market"`
	Dataset         string         `json:"dataset"`
	Symbol          string         `json:"symbol,omitempty"`
	EventTime       time.Time      `json:"event_time"`
	IngestTime      time.Time      `json:"ingest_time"`
	JobID           string         `json:"job_id"`
	ProducerVersion string         `json:"producer_version"`
	Payload         map[string]any `json:"payload"`
}
