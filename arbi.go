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
	e := binance.NewEngine(cfg, []string{"USDT"})
	e.Run()
}
