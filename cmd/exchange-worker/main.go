package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"exchange-data-platform/internal/app"
	loader "exchange-data-platform/internal/config"
	"exchange-data-platform/internal/connector/common"
	"exchange-data-platform/internal/observability/health"
	"exchange-data-platform/internal/observability/logging"
	"exchange-data-platform/internal/observability/metrics"
	"exchange-data-platform/internal/orchestration/runner"
)

func main() {
	configPath := flag.String("config", "configs/exchanges/binance.yml", "path to worker config")
	addr := flag.String("health-addr", ":8080", "health server listen address")
	flag.Parse()

	logging.Setup()
	metrics.Init()

	cfg, err := loader.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	connector, err := common.NewConnector(cfg)
	if err != nil {
		log.Fatalf("create connector: %v", err)
	}
	healthState := health.NewState(cfg.Exchange, cfg.HealthMaxAge)
	svc, err := runner.NewService(cfg, connector)
	if err != nil {
		log.Fatalf("create service: %v", err)
	}
	controller := app.NewController(cfg, svc, healthState)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", healthState.Handler)
		mux.HandleFunc("/livez", health.LiveHandler)
		mux.Handle("/metrics", metrics.Handler())
		mux.HandleFunc("/run-once", controller.RunOnceHandler)
		if err := http.ListenAndServe(*addr, mux); err != nil {
			log.Printf("health server stopped: %v", err)
		}
	}()

	if err := runner.Start(ctx, cfg, svc, healthState); err != nil {
		log.Fatal(err)
	}
}
