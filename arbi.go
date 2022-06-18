package main

import (
	"context"

	"github.com/olafszymanski/arbi/config"
	"github.com/olafszymanski/arbi/internal/broker"
	"github.com/olafszymanski/arbi/internal/database"
	"github.com/olafszymanski/arbi/pkg/logger"
)

func main() {
	l := logger.NewLogger()
	cfg := config.NewConfig(&l)
	s := database.NewStore(&l, cfg, context.Background())
	defer s.Disconnect()

	binance := broker.NewBinance(&l, cfg, s, map[string][]string{
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
