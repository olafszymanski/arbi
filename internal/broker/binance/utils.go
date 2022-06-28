package binance

import (
	"fmt"
	"strings"
)

func makeWebsocketUrl(pair string) string {
	return fmt.Sprintf("wss://stream.binance.com/ws/%s@miniTicker", strings.ToLower(pair))
}

func makeApiUrl(symbols map[string][]string) string {
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
