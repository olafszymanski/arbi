package main

import (
	"github.com/olafszymanski/arbi/cmd/config"
	"github.com/olafszymanski/arbi/internal/exchange"
)

func main() {
	cfg := config.NewConfig()
	binance := exchange.NewBinance(cfg, map[string][]string{
		"BTC": []string{
			"USDT",
			"USDC",
			"TUSD",
			"DAI",
		},
		"ETH": []string{
			"USDT",
			"USDC",
			"TUSD",
			"DAI",
		},
		"BNB": []string{
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
