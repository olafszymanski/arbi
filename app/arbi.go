package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/olafszymanski/arbi/app/config"
	"github.com/olafszymanski/arbi/app/internal/exchange"
	"github.com/olafszymanski/arbi/app/internal/postgres"
	"github.com/rs/zerolog"
)

func main() {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("|  %-6s|", i))
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("%s      ", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s: ", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}
	l := zerolog.New(output).With().Timestamp().Logger()

	cfg := config.NewConfig(&l)
	s := postgres.NewStore(&l, cfg)
	binance := exchange.NewBinance(&l, cfg, s, map[string][]string{
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
