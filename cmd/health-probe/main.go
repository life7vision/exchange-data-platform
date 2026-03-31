package main

import (
	"flag"
	"net/http"
	"os"
	"time"
)

func main() {
	addr := flag.String("addr", "http://127.0.0.1:8080/healthz", "probe target")
	flag.Parse()

	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(*addr)
	if err != nil || resp.StatusCode >= 300 {
		os.Exit(1)
	}
	os.Exit(0)
}
