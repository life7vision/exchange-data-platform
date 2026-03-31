package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
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
}

func (s *State) MarkFailure(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.LastFailureAt = time.Now().UTC()
	if err != nil {
		s.status.LastError = err.Error()
	}
}

func (s *State) Handler(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	code := http.StatusOK
	if s.status.LastSuccessAt.IsZero() || time.Since(s.status.LastSuccessAt) > s.status.MaxAge {
		code = http.StatusServiceUnavailable
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(s.status)
}

func LiveHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "ok")
}
