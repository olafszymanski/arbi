package main

import (
	"github.com/olafszymanski/arbi/app/config"
	"github.com/olafszymanski/arbi/app/internal/exchange"
	"github.com/olafszymanski/arbi/app/internal/postgres"
	"github.com/olafszymanski/arbi/app/pkg/logger"
)

func main() {
	l := logger.NewLogger()
	cfg := config.NewConfig(&l)
	s := postgres.NewStore(&l, cfg)
	binance := exchange.NewBinance(&l, cfg, s, map[string][]string{
		"BTC": {
			"USDT",
			"USDC",
			"TUSD",
			"DAI",
		},
		"ETH": {
			"USDT",
			"USDC",
			"TUSD",
			"DAI",
		},
		"BNB": {
			"USDT",
			"USDC",
			"TUSD",
			"DAI",
		},
	})
	binance.Subscribe()

	for {
	}
}
