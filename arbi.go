package main

import (
	"github.com/olafszymanski/arbi/config"
	"github.com/olafszymanski/arbi/internal/exchange"
	"github.com/olafszymanski/arbi/internal/postgres"
)

func main() {
	cfg := config.NewConfig()
	s := postgres.NewStore(cfg)
	binance := exchange.NewBinance(cfg, s, map[string][]string{
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
