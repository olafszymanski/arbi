package main

import (
	"github.com/olafszymanski/arbi/config"
	"github.com/olafszymanski/arbi/internal/binance"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetReportCaller(true)
	log.SetFormatter(&log.JSONFormatter{})

	cfg := config.NewConfig()
	e, err := binance.NewEngine(cfg, []string{"USDT", "BTC", "ETH", "DAI", "BUSD", "BNB"})
	if err != nil {
		log.WithError(err).Panic()
	}
	e.Run()
}
