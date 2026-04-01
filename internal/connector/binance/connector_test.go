package binance

import "testing"

func TestIsBinanceTradeStreamMessage(t *testing.T) {
	t.Parallel()

	if isBinanceTradeStreamMessage(map[string]any{"result": nil, "id": 1}) {
		t.Fatalf("expected subscribe ack to be filtered")
	}
	if !isBinanceTradeStreamMessage(map[string]any{"e": "trade", "t": 1, "p": "68000.1", "q": "0.01"}) {
		t.Fatalf("expected live trade message to pass")
	}
}

func TestIsBinanceOrderbookDeltaMessage(t *testing.T) {
	t.Parallel()

	if isBinanceOrderbookDeltaMessage(map[string]any{"stream": "btcusdt@depth"}) {
		t.Fatalf("expected non-delta message to be filtered")
	}
	if !isBinanceOrderbookDeltaMessage(map[string]any{"U": 10, "u": 11, "b": []any{}, "a": []any{}}) {
		t.Fatalf("expected depth delta message to pass")
	}
}

func TestIsBinanceLiquidationMessage(t *testing.T) {
	t.Parallel()

	if isBinanceLiquidationMessage(map[string]any{"e": "trade"}) {
		t.Fatalf("expected non-liquidation message to be filtered")
	}
	if !isBinanceLiquidationMessage(map[string]any{
		"e": "forceOrder",
		"o": map[string]any{"s": "BTCUSDT", "p": "68000", "q": "1.5"},
	}) {
		t.Fatalf("expected liquidation message to pass")
	}
}

func BenchmarkIsBinanceTradeStreamMessage(b *testing.B) {
	msg := map[string]any{"e": "trade", "t": 1, "p": "68000.1", "q": "0.01"}
	for i := 0; i < b.N; i++ {
		isBinanceTradeStreamMessage(msg)
	}
}

func BenchmarkIsBinanceOrderbookDeltaMessage(b *testing.B) {
	msg := map[string]any{"U": 10, "u": 11, "b": []any{}, "a": []any{}}
	for i := 0; i < b.N; i++ {
		isBinanceOrderbookDeltaMessage(msg)
	}
}
