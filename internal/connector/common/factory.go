package common

import (
	"fmt"

	"exchange-data-platform/internal/connector/api"
	"exchange-data-platform/internal/connector/binance"
	binancetr "exchange-data-platform/internal/connector/binance_tr"
	"exchange-data-platform/internal/connector/bybit"
	bybittr "exchange-data-platform/internal/connector/bybit_tr"
	modelconfig "exchange-data-platform/internal/model/config"
)

func NewConnector(cfg modelconfig.AppConfig) (api.Connector, error) {
	switch cfg.Exchange {
	case "binance":
		return binance.New(cfg), nil
	case "bybit":
		return bybit.New(cfg), nil
	case "binance_tr":
		return binancetr.New(cfg), nil
	case "bybit_tr":
		return bybittr.New(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported exchange %s", cfg.Exchange)
	}
}
