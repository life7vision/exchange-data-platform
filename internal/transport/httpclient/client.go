package httpclient

import (
	"io"
	"net"
	"net/http"
	"time"
)

type retryTransport struct {
	base       http.RoundTripper
	maxRetries int
	backoff    time.Duration
}

func New(timeout time.Duration, maxRetries int, backoff time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: retryTransport{
			base:       http.DefaultTransport,
			maxRetries: maxRetries,
			backoff:    backoff,
		},
	}
}

func (t retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		cloned := req.Clone(req.Context())
		resp, err = t.base.RoundTrip(cloned)
		if err == nil && resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}
		if resp != nil && resp.Body != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
		if attempt < t.maxRetries {
			time.Sleep(t.backoff * time.Duration(attempt+1))
		}
	}
	return resp, err
}

func IsTemporary(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	return err != nil && ((errorAs(err, &netErr) && netErr.Temporary()) || err == io.EOF)
}

func errorAs(err error, target any) bool {
	switch t := target.(type) {
	case *net.Error:
		if ne, ok := err.(net.Error); ok {
			*t = ne
			return true
		}
	}
	return false
}
