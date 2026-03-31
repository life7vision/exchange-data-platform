package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

var batchesProcessed uint64
var recordsWritten uint64
var rejectCount uint64

func IncBatches() {
	atomic.AddUint64(&batchesProcessed, 1)
}

func AddRecords(n int) {
	atomic.AddUint64(&recordsWritten, uint64(n))
}

func IncRejects() {
	atomic.AddUint64(&rejectCount, 1)
}

func Handler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = fmt.Fprintf(
		w,
		"exchange_batches_processed %d\nexchange_records_written %d\nexchange_reject_count %d\n",
		atomic.LoadUint64(&batchesProcessed),
		atomic.LoadUint64(&recordsWritten),
		atomic.LoadUint64(&rejectCount),
	)
}
