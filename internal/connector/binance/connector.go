package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"exchange-data-platform/internal/connector/api"
	modelconfig "exchange-data-platform/internal/model/config"
	"exchange-data-platform/internal/model/raw"
	"exchange-data-platform/internal/transport/httpclient"
	wsclient "exchange-data-platform/internal/transport/websocket"
)

type Connector struct {
	cfg    modelconfig.AppConfig
	client *http.Client
}

func New(cfg modelconfig.AppConfig) *Connector {
	return &Connector{cfg: cfg, client: httpclient.New(cfg.HTTPTimeout, cfg.MaxRetries, cfg.RetryBackoff)}
}

func (c *Connector) Name() string   { return c.cfg.Exchange }
func (c *Connector) Region() string { return c.cfg.Region }

func (c *Connector) Fetch(ctx context.Context, req api.FetchRequest) ([]raw.Envelope, error) {
	switch req.Dataset {
	case "instruments":
		return c.fetchInstruments(ctx, req)
	case "tickers":
		return c.fetchTickers(ctx, req)
	case "trades":
		return c.fetchTrades(ctx, req)
	case "klines":
		return c.fetchKlines(ctx, req)
	case "orderbook_snapshots":
		return c.fetchOrderbook(ctx, req)
	case "funding_rates":
		return c.fetchFundingRates(ctx, req)
	case "open_interest":
		return c.fetchOpenInterest(ctx, req)
	case "trades_stream":
		return c.fetchTradesStream(ctx, req)
	case "orderbook_deltas":
		return c.fetchOrderbookStream(ctx, req)
	case "liquidations_stream":
		return c.fetchLiquidationsStream(ctx, req)
	default:
		return nil, fmt.Errorf("dataset %s not implemented for %s", req.Dataset, c.cfg.Exchange)
	}
}

func (c *Connector) getJSON(ctx context.Context, path string, out any) error {
	baseURL := c.cfg.BaseURL
	if c.cfg.DerivativesURL != "" && (reqPathUsesDerivatives(path)) {
		baseURL = c.cfg.DerivativesURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func reqPathUsesDerivatives(path string) bool {
	return len(path) >= 6 && path[:6] == "/fapi/"
}

func (c *Connector) defaultSymbol() string {
	if len(c.cfg.DefaultSymbols) > 0 {
		return c.cfg.DefaultSymbols[0]
	}
	return "BTCUSDT"
}

func (c *Connector) isDerivativesMarket(market string) bool {
	return market != "" && market != "spot"
}

func (c *Connector) fetchInstruments(ctx context.Context, req api.FetchRequest) ([]raw.Envelope, error) {
	var payload struct {
		Symbols []map[string]any `json:"symbols"`
	}
	path := "/api/v3/exchangeInfo"
	if c.isDerivativesMarket(req.Market) {
		path = "/fapi/v1/exchangeInfo"
	}
	if err := c.getJSON(ctx, path, &payload); err != nil {
		return nil, err
	}
	rows := make([]raw.Envelope, 0, len(payload.Symbols))
	now := time.Now().UTC()
	for _, item := range payload.Symbols {
		symbol, _ := item["symbol"].(string)
		rows = append(rows, raw.Envelope{
			Exchange:        c.cfg.Exchange,
			Region:          c.cfg.Region,
			Market:          req.Market,
			Dataset:         req.Dataset,
			Symbol:          symbol,
			EventTime:       now,
			IngestTime:      now,
			JobID:           req.JobID,
			ProducerVersion: c.cfg.ProducerVersion,
			Payload:         item,
		})
		if req.Limit > 0 && len(rows) >= req.Limit {
			break
		}
	}
	return rows, nil
}

func (c *Connector) fetchTickers(ctx context.Context, req api.FetchRequest) ([]raw.Envelope, error) {
	var payload []map[string]any
	path := "/api/v3/ticker/bookTicker"
	if c.isDerivativesMarket(req.Market) {
		path = "/fapi/v1/ticker/bookTicker"
	}
	if err := c.getJSON(ctx, path, &payload); err != nil {
		return nil, err
	}
	rows := make([]raw.Envelope, 0, len(payload))
	now := time.Now().UTC()
	for _, item := range payload {
		symbol, _ := item["symbol"].(string)
		rows = append(rows, raw.Envelope{
			Exchange:        c.cfg.Exchange,
			Region:          c.cfg.Region,
			Market:          req.Market,
			Dataset:         req.Dataset,
			Symbol:          symbol,
			EventTime:       now,
			IngestTime:      now,
			JobID:           req.JobID,
			ProducerVersion: c.cfg.ProducerVersion,
			Payload:         item,
		})
		if req.Limit > 0 && len(rows) >= req.Limit {
			break
		}
	}
	return rows, nil
}

func (c *Connector) fetchTrades(ctx context.Context, req api.FetchRequest) ([]raw.Envelope, error) {
	symbol := c.defaultSymbol()
	var payload []map[string]any
	path := fmt.Sprintf("/api/v3/trades?symbol=%s&limit=%d", url.QueryEscape(symbol), req.Limit)
	if c.isDerivativesMarket(req.Market) {
		path = fmt.Sprintf("/fapi/v1/trades?symbol=%s&limit=%d", url.QueryEscape(symbol), req.Limit)
	}
	if err := c.getJSON(ctx, path, &payload); err != nil {
		return nil, err
	}
	rows := make([]raw.Envelope, 0, len(payload))
	for _, item := range payload {
		eventTime := time.Now().UTC()
		if v, ok := item["time"].(float64); ok {
			eventTime = time.UnixMilli(int64(v))
		}
		rows = append(rows, raw.Envelope{
			Exchange:        c.cfg.Exchange,
			Region:          c.cfg.Region,
			Market:          req.Market,
			Dataset:         req.Dataset,
			Symbol:          symbol,
			EventTime:       eventTime,
			IngestTime:      time.Now().UTC(),
			JobID:           req.JobID,
			ProducerVersion: c.cfg.ProducerVersion,
			Payload:         item,
		})
	}
	return rows, nil
}

func (c *Connector) fetchKlines(ctx context.Context, req api.FetchRequest) ([]raw.Envelope, error) {
	symbol := c.defaultSymbol()
	var payload [][]any
	path := fmt.Sprintf("/api/v3/klines?symbol=%s&interval=1m&limit=%d", url.QueryEscape(symbol), req.Limit)
	if c.isDerivativesMarket(req.Market) {
		path = fmt.Sprintf("/fapi/v1/klines?symbol=%s&interval=1m&limit=%d", url.QueryEscape(symbol), req.Limit)
	}
	if err := c.getJSON(ctx, path, &payload); err != nil {
		return nil, err
	}
	rows := make([]raw.Envelope, 0, len(payload))
	for _, item := range payload {
		eventTime := time.Now().UTC()
		if len(item) > 0 {
			if v, ok := item[0].(float64); ok {
				eventTime = time.UnixMilli(int64(v))
			}
		}
		rows = append(rows, raw.Envelope{
			Exchange:        c.cfg.Exchange,
			Region:          c.cfg.Region,
			Market:          req.Market,
			Dataset:         req.Dataset,
			Symbol:          symbol,
			EventTime:       eventTime,
			IngestTime:      time.Now().UTC(),
			JobID:           req.JobID,
			ProducerVersion: c.cfg.ProducerVersion,
			Payload:         map[string]any{"kline": item},
		})
	}
	return rows, nil
}

func (c *Connector) fetchOrderbook(ctx context.Context, req api.FetchRequest) ([]raw.Envelope, error) {
	symbol := c.defaultSymbol()
	var payload map[string]any
	path := fmt.Sprintf("/api/v3/depth?symbol=%s&limit=%d", url.QueryEscape(symbol), req.Limit)
	if c.isDerivativesMarket(req.Market) {
		path = fmt.Sprintf("/fapi/v1/depth?symbol=%s&limit=%d", url.QueryEscape(symbol), req.Limit)
	}
	if err := c.getJSON(ctx, path, &payload); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return []raw.Envelope{{
		Exchange:        c.cfg.Exchange,
		Region:          c.cfg.Region,
		Market:          req.Market,
		Dataset:         req.Dataset,
		Symbol:          symbol,
		EventTime:       now,
		IngestTime:      now,
		JobID:           req.JobID,
		ProducerVersion: c.cfg.ProducerVersion,
		Payload:         payload,
	}}, nil
}

func (c *Connector) fetchFundingRates(ctx context.Context, req api.FetchRequest) ([]raw.Envelope, error) {
	if c.cfg.DerivativesURL == "" {
		return nil, fmt.Errorf("funding_rates require derivatives_url for %s", c.cfg.Exchange)
	}
	symbol := c.defaultSymbol()
	var payload []map[string]any
	if err := c.getJSON(ctx, fmt.Sprintf("/fapi/v1/fundingRate?symbol=%s&limit=%d", url.QueryEscape(symbol), req.Limit), &payload); err != nil {
		return nil, err
	}
	rows := make([]raw.Envelope, 0, len(payload))
	for _, item := range payload {
		eventTime := time.Now().UTC()
		if v, ok := item["fundingTime"].(float64); ok {
			eventTime = time.UnixMilli(int64(v))
		}
		rows = append(rows, raw.Envelope{
			Exchange:        c.cfg.Exchange,
			Region:          c.cfg.Region,
			Market:          req.Market,
			Dataset:         req.Dataset,
			Symbol:          symbol,
			EventTime:       eventTime,
			IngestTime:      time.Now().UTC(),
			JobID:           req.JobID,
			ProducerVersion: c.cfg.ProducerVersion,
			Payload:         item,
		})
	}
	return rows, nil
}

func (c *Connector) fetchOpenInterest(ctx context.Context, req api.FetchRequest) ([]raw.Envelope, error) {
	if c.cfg.DerivativesURL == "" {
		return nil, fmt.Errorf("open_interest require derivatives_url for %s", c.cfg.Exchange)
	}
	symbol := c.defaultSymbol()
	var payload map[string]any
	if err := c.getJSON(ctx, fmt.Sprintf("/fapi/v1/openInterest?symbol=%s", url.QueryEscape(symbol)), &payload); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return []raw.Envelope{{
		Exchange:        c.cfg.Exchange,
		Region:          c.cfg.Region,
		Market:          req.Market,
		Dataset:         req.Dataset,
		Symbol:          symbol,
		EventTime:       now,
		IngestTime:      now,
		JobID:           req.JobID,
		ProducerVersion: c.cfg.ProducerVersion,
		Payload:         payload,
	}}, nil
}

func (c *Connector) fetchTradesStream(ctx context.Context, req api.FetchRequest) ([]raw.Envelope, error) {
	symbol := c.defaultSymbol()
	client := wsclient.Client{
		URL:        c.cfg.WebSocketURL + "/" + strings.ToLower(symbol) + "@trade",
		MaxRetries: c.cfg.MaxRetries,
		Backoff:    c.cfg.RetryBackoff,
	}
	messages, err := client.ReadJSONBatch(ctx, nil, req.Limit, 10*time.Second)
	if err != nil {
		return nil, err
	}
	rows := make([]raw.Envelope, 0, len(messages))
	for _, item := range messages {
		if !isBinanceTradeStreamMessage(item) {
			continue
		}
		eventTime := time.Now().UTC()
		if v, ok := item["E"].(float64); ok {
			eventTime = time.UnixMilli(int64(v))
		}
		rows = append(rows, raw.Envelope{
			Exchange:        c.cfg.Exchange,
			Region:          c.cfg.Region,
			Market:          req.Market,
			Dataset:         req.Dataset,
			Symbol:          symbol,
			EventTime:       eventTime,
			IngestTime:      time.Now().UTC(),
			JobID:           req.JobID,
			ProducerVersion: c.cfg.ProducerVersion,
			Payload:         item,
		})
	}
	return rows, nil
}

func (c *Connector) fetchOrderbookStream(ctx context.Context, req api.FetchRequest) ([]raw.Envelope, error) {
	symbol := c.defaultSymbol()
	client := wsclient.Client{
		URL:        c.cfg.WebSocketURL + "/" + strings.ToLower(symbol) + "@depth@100ms",
		MaxRetries: c.cfg.MaxRetries,
		Backoff:    c.cfg.RetryBackoff,
	}
	messages, err := client.ReadJSONBatch(ctx, nil, req.Limit, 10*time.Second)
	if err != nil {
		return nil, err
	}
	rows := make([]raw.Envelope, 0, len(messages))
	for _, item := range messages {
		if !isBinanceOrderbookDeltaMessage(item) {
			continue
		}
		eventTime := time.Now().UTC()
		if v, ok := item["E"].(float64); ok {
			eventTime = time.UnixMilli(int64(v))
		}
		rows = append(rows, raw.Envelope{
			Exchange:        c.cfg.Exchange,
			Region:          c.cfg.Region,
			Market:          req.Market,
			Dataset:         req.Dataset,
			Symbol:          symbol,
			EventTime:       eventTime,
			IngestTime:      time.Now().UTC(),
			JobID:           req.JobID,
			ProducerVersion: c.cfg.ProducerVersion,
			Payload:         item,
		})
	}
	return rows, nil
}

func (c *Connector) fetchLiquidationsStream(ctx context.Context, req api.FetchRequest) ([]raw.Envelope, error) {
	if !c.isDerivativesMarket(req.Market) {
		return nil, fmt.Errorf("liquidations_stream requires derivatives market")
	}
	symbol := c.defaultSymbol()
	client := wsclient.Client{
		URL:        c.cfg.WebSocketURL + "/" + strings.ToLower(symbol) + "@forceOrder",
		MaxRetries: c.cfg.MaxRetries,
		Backoff:    c.cfg.RetryBackoff,
	}
	messages, err := client.ReadJSONBatch(ctx, nil, req.Limit, 10*time.Second)
	if err != nil {
		return nil, err
	}
	rows := make([]raw.Envelope, 0, len(messages))
	for _, item := range messages {
		if !isBinanceLiquidationMessage(item) {
			continue
		}
		eventTime := time.Now().UTC()
		if v, ok := item["E"].(float64); ok {
			eventTime = time.UnixMilli(int64(v))
		}
		rows = append(rows, raw.Envelope{
			Exchange:        c.cfg.Exchange,
			Region:          c.cfg.Region,
			Market:          req.Market,
			Dataset:         req.Dataset,
			Symbol:          symbol,
			EventTime:       eventTime,
			IngestTime:      time.Now().UTC(),
			JobID:           req.JobID,
			ProducerVersion: c.cfg.ProducerVersion,
			Payload:         item,
		})
	}
	return rows, nil
}

func isBinanceTradeStreamMessage(item map[string]any) bool {
	eventType, ok := item["e"].(string)
	if !ok || eventType != "trade" {
		return false
	}
	_, hasTradeID := item["t"]
	_, hasPrice := item["p"]
	_, hasQty := item["q"]
	return hasTradeID && hasPrice && hasQty
}

func isBinanceOrderbookDeltaMessage(item map[string]any) bool {
	_, hasFirstUpdateID := item["U"]
	_, hasFinalUpdateID := item["u"]
	bids, hasBids := item["b"]
	asks, hasAsks := item["a"]
	return hasFirstUpdateID && hasFinalUpdateID && hasBids && hasAsks && bids != nil && asks != nil
}

func isBinanceLiquidationMessage(item map[string]any) bool {
	eventType, ok := item["e"].(string)
	if !ok || eventType != "forceOrder" {
		return false
	}
	order, ok := item["o"].(map[string]any)
	if !ok {
		return false
	}
	_, hasSymbol := order["s"]
	_, hasPrice := order["p"]
	_, hasQty := order["q"]
	return hasSymbol && hasPrice && hasQty
}
