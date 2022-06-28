package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/olafszymanski/arbi/config"
	"github.com/olafszymanski/arbi/internal/broker/binance"
	"github.com/olafszymanski/arbi/internal/database"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})

	cfg := config.NewConfig("config/config.yml")
	s := database.NewStore(context.Background(), cfg)
	defer s.Disconnect()

	binance := binance.New(cfg, s, map[string][]string{
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
	binance.Subscribe(done)

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			select {
			case <-done:
			case <-time.After(time.Microsecond):
			}
			return
		}
	}
}
