package main

import (
	"github.com/olafszymanski/arbi/config"
	"github.com/olafszymanski/arbi/internal/exchange"
)

func main() {
	cfg := config.NewConfig()
	binance := exchange.NewBinance(cfg, map[string][]string{
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
