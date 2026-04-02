package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Global Metrics instance
var Global *Metrics

// Metrics holds all Prometheus metrics for the exchange data platform.
type Metrics struct {
	// Data processing
	DataProcessed   prometheus.Counter   // total records processed
	DataErrors      prometheus.Counter   // total processing errors
	BatchSize       prometheus.Histogram // batch size distribution
	ProcessingTime  prometheus.Histogram // processing time in seconds

	// API interactions
	APIRequests     prometheus.Counter   // total API requests
	APIErrors       prometheus.Counter   // total API errors
	APILatency      prometheus.Histogram // API response latency in seconds

	// Pipeline stages
	TempFilesWritten   prometheus.Counter   // files written to temp spool
	ParquetWritten     prometheus.Counter   // parquet files successfully written
	ManifestWritten    prometheus.Counter   // manifest files written
	RejectsProcessed   prometheus.Counter   // rejected batches
	FilesCleaned       prometheus.Counter   // temp files cleaned up

	// Health & status
	LastSuccessSync prometheus.Gauge     // timestamp of last successful sync
	WorkerHealth    *prometheus.GaugeVec // worker health status by exchange
	QueueDepth      *prometheus.GaugeVec // queue depth by exchange
}

// Init initializes and registers all Prometheus metrics.
func Init() *Metrics {
	m := &Metrics{
		DataProcessed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "exchange_data_records_processed_total",
			Help: "Total number of data records processed",
		}),
		DataErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "exchange_data_errors_total",
			Help: "Total number of processing errors",
		}),
		BatchSize: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "exchange_batch_size",
			Help:    "Batch size distribution",
			Buckets: prometheus.ExponentialBuckets(10, 2, 8),
		}),
		ProcessingTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "exchange_processing_duration_seconds",
			Help:    "Time spent processing batches in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		APIRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: "exchange_api_requests_total",
			Help: "Total number of API requests",
		}),
		APIErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "exchange_api_errors_total",
			Help: "Total number of API errors",
		}),
		APILatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "exchange_api_latency_seconds",
			Help:    "API request latency in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		TempFilesWritten: promauto.NewCounter(prometheus.CounterOpts{
			Name: "exchange_temp_files_written_total",
			Help: "Total number of temp files written",
		}),
		ParquetWritten: promauto.NewCounter(prometheus.CounterOpts{
			Name: "exchange_parquet_written_total",
			Help: "Total number of parquet files written",
		}),
		ManifestWritten: promauto.NewCounter(prometheus.CounterOpts{
			Name: "exchange_manifest_written_total",
			Help: "Total number of manifests written",
		}),
		RejectsProcessed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "exchange_rejects_total",
			Help: "Total number of rejected batches",
		}),
		FilesCleaned: promauto.NewCounter(prometheus.CounterOpts{
			Name: "exchange_files_cleaned_total",
			Help: "Total number of temp files cleaned",
		}),
		LastSuccessSync: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "exchange_last_success_sync_timestamp",
			Help: "Timestamp of the last successful sync",
		}),
		WorkerHealth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "exchange_worker_health",
				Help: "Worker health status (1=healthy, 0=unhealthy)",
			},
			[]string{"exchange"},
		),
		QueueDepth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "exchange_queue_depth",
				Help: "Number of items in worker queue",
			},
			[]string{"exchange"},
		),
	}
	Global = m
	return m
}

// Handler returns the Prometheus metrics HTTP handler.
func Handler() http.Handler {
	return promhttp.Handler()
}

// RecordBatch records batch processing metrics.
func (m *Metrics) RecordBatch(recordCount int, processingTimeMs int64) {
	m.DataProcessed.Add(float64(recordCount))
	m.BatchSize.Observe(float64(recordCount))
	m.ProcessingTime.Observe(float64(processingTimeMs) / 1000.0)
}

// RecordAPICall records API request metrics.
func (m *Metrics) RecordAPICall(latencyMs int64, err bool) {
	m.APIRequests.Inc()
	m.APILatency.Observe(float64(latencyMs) / 1000.0)
	if err {
		m.APIErrors.Inc()
	}
}

// RecordError records a processing error.
func (m *Metrics) RecordError() {
	m.DataErrors.Inc()
}

// RecordReject records a rejected batch.
func (m *Metrics) RecordReject() {
	m.RejectsProcessed.Inc()
}

// UpdateLastSync updates the last successful sync timestamp.
func (m *Metrics) UpdateLastSync() {
	m.LastSuccessSync.Set(float64(time.Now().Unix()))
}

// SetWorkerHealth updates worker health status.
func (m *Metrics) SetWorkerHealth(exchange string, healthy bool) {
	status := 0.0
	if healthy {
		status = 1.0
	}
	m.WorkerHealth.WithLabelValues(exchange).Set(status)
}

// SetQueueDepth updates queue depth for an exchange.
func (m *Metrics) SetQueueDepth(exchange string, depth int) {
	m.QueueDepth.WithLabelValues(exchange).Set(float64(depth))
}
