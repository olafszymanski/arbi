package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/olafszymanski/arbi/config"
	"github.com/olafszymanski/arbi/internal/broker/binance"
	"github.com/olafszymanski/arbi/internal/database"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetReportCaller(true)
	log.SetFormatter(&log.JSONFormatter{})

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})

	cfg := config.NewConfig()
	s := database.NewStore(context.Background(), cfg)
	defer s.Disconnect()

	binance := binance.New(cfg, s, map[string][]string{
		"BTC": {
			"USDT",
			"USDC",
			"TUSD",
			"DAI",
			"BUSD",
		},
		"ETH": {
			"USDT",
			"USDC",
			"TUSD",
			"DAI",
			"BUSD",
		},
		"BNB": {
			"USDT",
			"USDC",
			"TUSD",
			"DAI",
			"BUSD",
		},
	})
	binance.Subscribe(context.Background(), done)

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
