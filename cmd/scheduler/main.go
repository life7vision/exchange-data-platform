package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	loader "exchange-data-platform/internal/config"
	scheduler "exchange-data-platform/internal/orchestration/scheduler"
)

func main() {
	jobsDir := flag.String("jobs-dir", "configs/jobs", "directory containing job configs")
	dispatch := flag.Bool("dispatch", true, "dispatch jobs to workers")
	flag.Parse()

	jobs, err := loader.LoadJobs(*jobsDir)
	if err != nil {
		log.Fatal(err)
	}
	client := scheduler.NewClient()

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
					log.Printf("initial dispatch failed for %s: %v", job.Name, err)
				} else {
					log.Printf("initial dispatch succeeded for %s", job.Name)
				}
			}

			for range ticker.C {
				if !*dispatch {
					continue
				}
				if err := client.Dispatch(job); err != nil {
					log.Printf("dispatch failed for %s: %v", job.Name, err)
				} else {
					log.Printf("dispatch succeeded for %s", job.Name)
				}
			}
		}()
	}

	select {}
}
