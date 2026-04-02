package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"exchange-data-platform/internal/observability/metrics"
)

type Status struct {
	Exchange      string    `json:"exchange"`
	LastSuccessAt time.Time `json:"last_success_at"`
	LastFailureAt time.Time `json:"last_failure_at"`
	LastError     string    `json:"last_error,omitempty"`
	MaxAge        time.Duration
}

type State struct {
	mu     sync.RWMutex
	status Status
}

func NewState(exchange string, maxAge time.Duration) *State {
	return &State{status: Status{Exchange: exchange, MaxAge: maxAge}}
}

func (s *State) MarkSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.LastSuccessAt = time.Now().UTC()
	s.status.LastError = ""
	
	// Update Prometheus metrics
	if metrics.Global != nil {
		metrics.Global.SetWorkerHealth(s.status.Exchange, true)
		metrics.Global.UpdateLastSync()
	}
}

func (s *State) MarkFailure(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.LastFailureAt = time.Now().UTC()
	if err != nil {
		s.status.LastError = err.Error()
	}
	
	// Update Prometheus metrics
	if metrics.Global != nil {
		metrics.Global.SetWorkerHealth(s.status.Exchange, false)
		metrics.Global.DataErrors.Inc()
	}
}

func (s *State) Handler(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	code := http.StatusOK
	healthy := true
	if s.status.LastSuccessAt.IsZero() || time.Since(s.status.LastSuccessAt) > s.status.MaxAge {
		code = http.StatusServiceUnavailable
		healthy = false
	}
	
	// Update Prometheus metrics
	if metrics.Global != nil {
		metrics.Global.SetWorkerHealth(s.status.Exchange, healthy)
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(s.status)
}

func LiveHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "ok")
}
