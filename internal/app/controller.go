package app

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	modelconfig "exchange-data-platform/internal/model/config"
	"exchange-data-platform/internal/observability/health"
	"exchange-data-platform/internal/pipeline/publish"
)

type Controller struct {
	mu          sync.Mutex
	cfg         modelconfig.AppConfig
	svc         publish.Service
	healthState *health.State
	running     bool
}

type RunRequest struct {
	Datasets []string `json:"datasets"`
	Markets  []string `json:"markets"`
}

func NewController(cfg modelconfig.AppConfig, svc publish.Service, healthState *health.State) *Controller {
	return &Controller{cfg: cfg, svc: svc, healthState: healthState}
}

var safeNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func (c *Controller) RunOnceHandler(w http.ResponseWriter, r *http.Request) {
	// Security: Require WORKER_API_TOKEN if set
	if token := os.Getenv("WORKER_API_TOKEN"); token != "" {
		auth := r.Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != token {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "unauthorized"})
			return
		}
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid request body"})
		return
	}

	// Security: validate that requested datasets and markets are safe (alphanumeric and underscores)
	// This prevents path traversal vulnerabilities in the storage layer.
	for _, ds := range req.Datasets {
		if !safeNameRegex.MatchString(ds) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid dataset name", "dataset": ds})
			return
		}
	}
	for _, m := range req.Markets {
		if !safeNameRegex.MatchString(m) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid market name", "market": m})
			return
		}
	}

	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "busy", "exchange": c.cfg.Exchange})
		return
	}

	runCfg := c.cfg
	if len(req.Datasets) > 0 {
		runCfg.Datasets = req.Datasets
	}
	if len(req.Markets) > 0 {
		runCfg.Markets = req.Markets
	}

	runSvc := c.svc
	runSvc.Config = runCfg
	c.running = true
	c.mu.Unlock()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		err := runSvc.RunOnce(ctx)

		c.mu.Lock()
		c.running = false
		c.mu.Unlock()

		if err != nil {
			c.healthState.MarkFailure(err)
			return
		}
		c.healthState.MarkSuccess()
	}()

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "accepted", "exchange": runCfg.Exchange})
}
