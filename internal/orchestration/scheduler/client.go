package scheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	modelconfig "exchange-data-platform/internal/model/config"
)

type Client struct {
	http *http.Client
}

func NewClient() Client {
	return NewClientWithHTTPClient(&http.Client{Timeout: 30 * time.Second})
}

func NewClientWithHTTPClient(client *http.Client) Client {
	return Client{http: client}
}

func (c Client) Dispatch(job modelconfig.JobConfig) error {
	body, err := json.Marshal(map[string]any{
		"datasets": job.Datasets,
		"markets":  job.Markets,
	})
	if err != nil {
		return fmt.Errorf("marshal job payload: %w", err)
	}
	resp, err := c.http.Post(job.WorkerURL+"/run-once", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("dispatch job %s: %w", job.Name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("dispatch job %s got status %d", job.Name, resp.StatusCode)
	}
	return nil
}
