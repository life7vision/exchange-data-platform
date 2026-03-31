package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	URL        string
	MaxRetries int
	Backoff    time.Duration
}

func (c Client) ReadJSONBatch(ctx context.Context, subscribe any, limit int, timeout time.Duration) ([]map[string]any, error) {
	rows := make([]map[string]any, 0, limit)
	attempts := c.MaxRetries + 1
	if attempts <= 0 {
		attempts = 1
	}
	if c.Backoff <= 0 {
		c.Backoff = time.Second
	}

	for attempt := 0; attempt < attempts && len(rows) < limit; attempt++ {
		conn, _, err := websocket.DefaultDialer.Dial(c.URL, nil)
		if err != nil {
			if attempt == attempts-1 {
				return rows, fmt.Errorf("dial websocket: %w", err)
			}
			time.Sleep(c.Backoff * time.Duration(attempt+1))
			continue
		}
		if subscribe != nil {
			if err := conn.WriteJSON(subscribe); err != nil {
				conn.Close()
				if attempt == attempts-1 {
					return rows, fmt.Errorf("write subscribe: %w", err)
				}
				time.Sleep(c.Backoff * time.Duration(attempt+1))
				continue
			}
		}
		if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			conn.Close()
			if attempt == attempts-1 {
				return rows, fmt.Errorf("set read deadline: %w", err)
			}
			time.Sleep(c.Backoff * time.Duration(attempt+1))
			continue
		}

		for len(rows) < limit {
			select {
			case <-ctx.Done():
				conn.Close()
				return rows, ctx.Err()
			default:
			}

			_, payload, err := conn.ReadMessage()
			if err != nil {
				break
			}
			var item map[string]any
			if err := json.Unmarshal(payload, &item); err != nil {
				continue
			}
			rows = append(rows, item)
		}
		conn.Close()
		if len(rows) >= limit {
			return rows, nil
		}
		if attempt < attempts-1 {
			time.Sleep(c.Backoff * time.Duration(attempt+1))
		}
	}
	if len(rows) > 0 {
		return rows, nil
	}
	return nil, fmt.Errorf("websocket batch ended without messages")
}
