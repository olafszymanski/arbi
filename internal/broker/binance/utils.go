package binance

import (
	"fmt"
	"strings"
	"time"

	"github.com/olafszymanski/arbi/config"
)

func websocketUrl(pair string) string {
	return fmt.Sprintf("wss://stream.binance.com/ws/%s@miniTicker", strings.ToLower(pair))
}

func apiUrl(symbols map[string][]string) string {
	url := "https://api.binance.com/api/v3/ticker/price?symbols=["
	tmpSyms := make([]string, 0, len(symbols))
	for crypto, stables := range symbols {
		for _, stable := range stables {
			tmpSyms = append(tmpSyms, fmt.Sprintf(`"%s"`, crypto+stable))
		}
	}
	syms := strings.Join(tmpSyms, ",")
	url += syms + "]"
	return url
}

func newOrderUrl(cfg *config.Config, symbol, side string) string {
	return fmt.Sprintf("https://api.binance.com/api/v3/order?signature=%s&symbol=%s&side=%s&type=MARKET&timestamp=%v", cfg.Binance.SecretKey, symbol, side, time.Now().UTC().UnixMilli())
}
