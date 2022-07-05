package main

import (
	"github.com/olafszymanski/arbi/config"
	"github.com/olafszymanski/arbi/internal/binance"
)

func main() {
	cfg := config.NewConfig()
	// binance.NewEngine(cfg, []string{"USDT", "DAI", "BTC", "ETH", "BNB"})
	e := binance.NewEngine(cfg, []string{"USDT", "DAI", "BTC", "ETH", "BNB"})
	e.Run()

	// log.SetReportCaller(true)
	// log.SetFormatter(&log.JSONFormatter{})
}
