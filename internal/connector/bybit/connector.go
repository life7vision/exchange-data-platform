package bybit

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"exchange-data-platform/internal/connector/api"
	modelconfig "exchange-data-platform/internal/model/config"
	"exchange-data-platform/internal/model/raw"
	"exchange-data-platform/internal/observability/metrics"
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
	slog.Debug("fetching dataset", "exchange", c.cfg.Exchange, "dataset", req.Dataset, "market", req.Market)
	rows, err := c.fetch(ctx, req)
	if err != nil {
		slog.Error("fetch failed", "exchange", c.cfg.Exchange, "dataset", req.Dataset, "market", req.Market, "err", err)
		metrics.IncExchangeError(c.cfg.Exchange)
		return nil, err
	}
	slog.Info("fetch successful", "exchange", c.cfg.Exchange, "dataset", req.Dataset, "market", req.Market, "count", len(rows))
	return rows, nil
}

func (c *Connector) fetch(ctx context.Context, req api.FetchRequest) ([]raw.Envelope, error) {
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

func (c *Connector) defaultSymbol() string {
	if len(c.cfg.DefaultSymbols) > 0 {
		return c.cfg.DefaultSymbols[0]
	}
	return "BTCUSDT"
}

func (c *Connector) categoryForMarket(market string) string {
	switch market {
	case "linear_perpetual":
		return "linear"
	case "inverse_perpetual":
		return "inverse"
	case "option":
		return "option"
	default:
		return "spot"
	}
}

func (c *Connector) getResult(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.BaseURL+path, nil)
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
	var body struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}
	return json.Unmarshal(body.Result, out)
}

func (c *Connector) fetchInstruments(ctx context.Context, req api.FetchRequest) ([]raw.Envelope, error) {
	var payload struct {
		List []map[string]any `json:"list"`
	}
	if err := c.getResult(ctx, fmt.Sprintf("/v5/market/instruments-info?category=%s&limit=1000", c.categoryForMarket(req.Market)), &payload); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	rows := make([]raw.Envelope, 0, len(payload.List))
	for _, item := range payload.List {
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
	var payload struct {
		List []map[string]any `json:"list"`
	}
	if err := c.getResult(ctx, fmt.Sprintf("/v5/market/tickers?category=%s", c.categoryForMarket(req.Market)), &payload); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	rows := make([]raw.Envelope, 0, len(payload.List))
	for _, item := range payload.List {
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
	var payload struct {
		List []map[string]any `json:"list"`
	}
	if err := c.getResult(ctx, fmt.Sprintf("/v5/market/recent-trade?category=%s&symbol=%s&limit=%d", c.categoryForMarket(req.Market), url.QueryEscape(symbol), req.Limit), &payload); err != nil {
		return nil, err
	}
	rows := make([]raw.Envelope, 0, len(payload.List))
	for _, item := range payload.List {
		eventTime := time.Now().UTC()
		if v, ok := item["time"].(string); ok {
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				eventTime = time.UnixMilli(parsed)
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
			Payload:         item,
		})
	}
	return rows, nil
}

func (c *Connector) fetchKlines(ctx context.Context, req api.FetchRequest) ([]raw.Envelope, error) {
	symbol := c.defaultSymbol()
	var payload struct {
		List [][]any `json:"list"`
	}
	if err := c.getResult(ctx, fmt.Sprintf("/v5/market/kline?category=%s&symbol=%s&interval=1&limit=%d", c.categoryForMarket(req.Market), url.QueryEscape(symbol), req.Limit), &payload); err != nil {
		return nil, err
	}
	rows := make([]raw.Envelope, 0, len(payload.List))
	for _, item := range payload.List {
		eventTime := time.Now().UTC()
		if len(item) > 0 {
			if v, ok := item[0].(string); ok {
				if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
					eventTime = time.UnixMilli(parsed)
				}
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
	if err := c.getResult(ctx, fmt.Sprintf("/v5/market/orderbook?category=%s&symbol=%s&limit=%d", c.categoryForMarket(req.Market), url.QueryEscape(symbol), req.Limit), &payload); err != nil {
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
	symbol := c.defaultSymbol()
	var payload struct {
		List []map[string]any `json:"list"`
	}
	if err := c.getResult(ctx, fmt.Sprintf("/v5/market/funding/history?category=linear&symbol=%s&limit=%d", url.QueryEscape(symbol), req.Limit), &payload); err != nil {
		return nil, err
	}
	rows := make([]raw.Envelope, 0, len(payload.List))
	for _, item := range payload.List {
		eventTime := time.Now().UTC()
		if v, ok := item["fundingRateTimestamp"].(string); ok {
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				eventTime = time.UnixMilli(parsed)
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
			Payload:         item,
		})
	}
	return rows, nil
}

func (c *Connector) fetchOpenInterest(ctx context.Context, req api.FetchRequest) ([]raw.Envelope, error) {
	symbol := c.defaultSymbol()
	var payload struct {
		List []map[string]any `json:"list"`
	}
	if err := c.getResult(ctx, fmt.Sprintf("/v5/market/open-interest?category=linear&symbol=%s&intervalTime=5min&limit=%d", url.QueryEscape(symbol), req.Limit), &payload); err != nil {
		return nil, err
	}
	rows := make([]raw.Envelope, 0, len(payload.List))
	for _, item := range payload.List {
		now := time.Now().UTC()
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
	}
	return rows, nil
}

func (c *Connector) fetchTradesStream(ctx context.Context, req api.FetchRequest) ([]raw.Envelope, error) {
	symbol := c.defaultSymbol()
	client := wsclient.Client{
		URL:        c.cfg.WebSocketURL,
		MaxRetries: c.cfg.MaxRetries,
		Backoff:    c.cfg.RetryBackoff,
	}
	subscribe := map[string]any{
		"op":   "subscribe",
		"args": []string{"publicTrade." + symbol},
	}
	messages, err := client.ReadJSONBatch(ctx, subscribe, req.Limit, 10*time.Second)
	if err != nil {
		return nil, err
	}
	rows := make([]raw.Envelope, 0, len(messages))
	for _, item := range messages {
		if !isBybitTradeStreamMessage(item) {
			continue
		}
		rows = append(rows, raw.Envelope{
			Exchange:        c.cfg.Exchange,
			Region:          c.cfg.Region,
			Market:          req.Market,
			Dataset:         req.Dataset,
			Symbol:          symbol,
			EventTime:       bybitEventTime(item),
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
		URL:        c.cfg.WebSocketURL,
		MaxRetries: c.cfg.MaxRetries,
		Backoff:    c.cfg.RetryBackoff,
	}
	subscribe := map[string]any{
		"op":   "subscribe",
		"args": []string{"orderbook.50." + symbol},
	}
	messages, err := client.ReadJSONBatch(ctx, subscribe, req.Limit, 10*time.Second)
	if err != nil {
		return nil, err
	}
	rows := make([]raw.Envelope, 0, len(messages))
	for _, item := range messages {
		if !isBybitOrderbookMessage(item) {
			continue
		}
		rows = append(rows, raw.Envelope{
			Exchange:        c.cfg.Exchange,
			Region:          c.cfg.Region,
			Market:          req.Market,
			Dataset:         req.Dataset,
			Symbol:          symbol,
			EventTime:       bybitEventTime(item),
			IngestTime:      time.Now().UTC(),
			JobID:           req.JobID,
			ProducerVersion: c.cfg.ProducerVersion,
			Payload:         item,
		})
	}
	return rows, nil
}

func (c *Connector) fetchLiquidationsStream(ctx context.Context, req api.FetchRequest) ([]raw.Envelope, error) {
	if c.categoryForMarket(req.Market) == "spot" {
		return nil, fmt.Errorf("liquidations_stream requires derivatives market")
	}
	symbol := c.defaultSymbol()
	client := wsclient.Client{
		URL:        c.cfg.WebSocketURL,
		MaxRetries: c.cfg.MaxRetries,
		Backoff:    c.cfg.RetryBackoff,
	}
	subscribe := map[string]any{
		"op":   "subscribe",
		"args": []string{"liquidation." + symbol},
	}
	messages, err := client.ReadJSONBatch(ctx, subscribe, req.Limit, 10*time.Second)
	if err != nil {
		return nil, err
	}
	rows := make([]raw.Envelope, 0, len(messages))
	for _, item := range messages {
		if !isBybitLiquidationMessage(item) {
			continue
		}
		rows = append(rows, raw.Envelope{
			Exchange:        c.cfg.Exchange,
			Region:          c.cfg.Region,
			Market:          req.Market,
			Dataset:         req.Dataset,
			Symbol:          symbol,
			EventTime:       bybitEventTime(item),
			IngestTime:      time.Now().UTC(),
			JobID:           req.JobID,
			ProducerVersion: c.cfg.ProducerVersion,
			Payload:         item,
		})
	}
	return rows, nil
}

func isBybitTradeStreamMessage(item map[string]any) bool {
	if !hasBybitTopicPrefix(item, "publicTrade.") {
		return false
	}
	data, ok := item["data"].([]any)
	return ok && len(data) > 0
}

func isBybitOrderbookMessage(item map[string]any) bool {
	if !hasBybitTopicPrefix(item, "orderbook.") {
		return false
	}
	data, ok := item["data"].(map[string]any)
	if !ok {
		return false
	}
	_, hasBids := data["b"]
	_, hasAsks := data["a"]
	return hasBids && hasAsks
}

func isBybitLiquidationMessage(item map[string]any) bool {
	if !hasBybitTopicPrefix(item, "liquidation.") {
		return false
	}
	data, ok := item["data"].([]any)
	return ok && len(data) > 0
}

func hasBybitTopicPrefix(item map[string]any, prefix string) bool {
	topic, ok := item["topic"].(string)
	return ok && strings.HasPrefix(topic, prefix)
}

func bybitEventTime(item map[string]any) time.Time {
	now := time.Now().UTC()
	if v, ok := item["ts"].(float64); ok {
		return time.UnixMilli(int64(v))
	}
	if v, ok := item["cts"].(float64); ok {
		return time.UnixMilli(int64(v))
	}
	return now
}
