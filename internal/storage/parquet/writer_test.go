package parquet

import (
	"exchange-data-platform/internal/model/raw"
	"os"
	"testing"
	"time"
)

func BenchmarkWriteStandardized(b *testing.B) {
	root, err := os.MkdirTemp("", "parquet-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(root)

	w := Writer{Root: root, SchemaVersion: "v1"}
	rows := make([]raw.Envelope, 1000)
	for i := 0; i < 1000; i++ {
		rows[i] = raw.Envelope{
			Exchange:        "binance",
			Region:          "us",
			Market:          "spot",
			Dataset:         "tickers",
			Symbol:          "BTCUSDT",
			EventTime:       time.Now(),
			IngestTime:      time.Now(),
			JobID:           "job-123",
			ProducerVersion: "v1.0.0",
			Payload:         map[string]any{"price": "50000.00", "volume": "100.0"},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := w.WriteStandardized("binance", "us", "spot", "tickers", "job-123", rows)
		if err != nil {
			b.Fatal(err)
		}
	}
}
