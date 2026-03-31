package bybit

import "testing"

func TestIsBybitTradeStreamMessage(t *testing.T) {
	t.Parallel()

	if isBybitTradeStreamMessage(map[string]any{"op": "subscribe", "success": true}) {
		t.Fatalf("expected subscribe ack to be filtered")
	}
	if !isBybitTradeStreamMessage(map[string]any{
		"topic": "publicTrade.BTCUSDT",
		"data":  []any{map[string]any{"p": "68000.1", "v": "0.01"}},
	}) {
		t.Fatalf("expected public trade message to pass")
	}
}

func TestIsBybitOrderbookMessage(t *testing.T) {
	t.Parallel()

	if isBybitOrderbookMessage(map[string]any{"topic": "orderbook.50.BTCUSDT"}) {
		t.Fatalf("expected malformed orderbook payload to be filtered")
	}
	if !isBybitOrderbookMessage(map[string]any{
		"topic": "orderbook.50.BTCUSDT",
		"data":  map[string]any{"a": []any{}, "b": []any{}},
	}) {
		t.Fatalf("expected orderbook payload to pass")
	}
}

func TestIsBybitLiquidationMessage(t *testing.T) {
	t.Parallel()

	if isBybitLiquidationMessage(map[string]any{"topic": "publicTrade.BTCUSDT", "data": []any{}}) {
		t.Fatalf("expected non-liquidation payload to be filtered")
	}
	if !isBybitLiquidationMessage(map[string]any{
		"topic": "liquidation.BTCUSDT",
		"data":  []any{map[string]any{"p": "68000.1"}},
	}) {
		t.Fatalf("expected liquidation payload to pass")
	}
}
