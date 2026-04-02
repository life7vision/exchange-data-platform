package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	loader "exchange-data-platform/internal/config"
	"exchange-data-platform/internal/observability/logging"
	"exchange-data-platform/internal/observability/metrics"
	scheduler "exchange-data-platform/internal/orchestration/scheduler"
)

func main() {
	jobsDir := flag.String("jobs-dir", "configs/jobs", "directory containing job configs")
	dispatch := flag.Bool("dispatch", true, "dispatch jobs to workers")
	metricsAddr := flag.String("metrics-addr", ":9090", "metrics server listen address")
	flag.Parse()

	logging.Setup()
	metrics.Init()
	logger := slog.Default()

	jobs, err := loader.LoadJobs(*jobsDir)
	if err != nil {
		log.Fatal(err)
	}
	client := scheduler.NewClient()

	// Start metrics server
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", metrics.Handler())
		logger.Info("starting metrics server", "addr", *metricsAddr)
		if err := http.ListenAndServe(*metricsAddr, mux); err != nil {
			logger.Error("metrics server stopped", "err", err)
		}
	}()

	for _, job := range jobs {
		fmt.Printf("{\"scheduler\":\"loaded\",\"job\":\"%s\",\"exchange\":\"%s\",\"mode\":\"%s\",\"interval\":\"%s\"}\n", job.Name, job.Exchange, job.Mode, job.Interval)
	}

	for _, job := range jobs {
		job := job
		if job.Interval <= 0 {
			job.Interval = time.Minute
		}
		go func() {
			ticker := time.NewTicker(job.Interval)
			defer ticker.Stop()

			if *dispatch {
				if err := client.Dispatch(job); err != nil {
					logger.Error("initial dispatch failed", "job", job.Name, "err", err)
				} else {
					logger.Info("initial dispatch succeeded", "job", job.Name)
				}
			}

			for range ticker.C {
				if !*dispatch {
					continue
				}
				if err := client.Dispatch(job); err != nil {
					logger.Error("dispatch failed", "job", job.Name, "err", err)
				} else {
					logger.Info("dispatch succeeded", "job", job.Name)
				}
			}
		}()
	}

	select {}
}
