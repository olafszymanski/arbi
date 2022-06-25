package main

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/olafszymanski/arbi/config"
	broker "github.com/olafszymanski/arbi/internal/broker/binance"
	"github.com/olafszymanski/arbi/internal/database"
)

func main() {
	cfg := config.NewConfig("config/config.yml")
	s := database.NewStore(context.Background(), cfg)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.WithError(r.(error)).Error("Goroutine panicked while commiting to the firestore")
			}
		}()

		for {
			err := s.Commit(context.Background())
			if err != nil {
				log.WithError(err).Panic()
			}
		}
	}()
	defer s.Disconnect()

	binance := broker.NewBinance(cfg, s, map[string][]string{
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
