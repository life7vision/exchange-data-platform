package binance_tr

import (
	binance "exchange-data-platform/internal/connector/binance"
	modelconfig "exchange-data-platform/internal/model/config"
)

func New(cfg modelconfig.AppConfig) *binance.Connector {
	return binance.New(cfg)
}
