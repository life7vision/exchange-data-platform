package unit

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	modelconfig "exchange-data-platform/internal/model/config"
	scheduler "exchange-data-platform/internal/orchestration/scheduler"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSchedulerDispatch(t *testing.T) {
	client := scheduler.NewClientWithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/run-once" {
				t.Fatalf("unexpected path %s", r.URL.Path)
			}
			return &http.Response{
				StatusCode: http.StatusAccepted,
				Body:       io.NopCloser(bytes.NewBufferString("{}")),
				Header:     make(http.Header),
			}, nil
		}),
	})
	err := client.Dispatch(modelconfig.JobConfig{
		Name:      "sample",
		WorkerURL: "http://worker",
		Datasets:  []string{"trades_stream"},
		Markets:   []string{"spot"},
	})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
}
