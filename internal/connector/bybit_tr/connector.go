package bybit_tr

import (
	bybit "exchange-data-platform/internal/connector/bybit"
	modelconfig "exchange-data-platform/internal/model/config"
)

func New(cfg modelconfig.AppConfig) *bybit.Connector {
	return bybit.New(cfg)
}
