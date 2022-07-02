package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/olafszymanski/arbi/config"
	log "github.com/sirupsen/logrus"
)

func websocketPricesUrl(pair string) string {
	return fmt.Sprintf("wss://stream.binance.com/ws/%s@miniTicker", strings.ToLower(pair))
}

func websocketUserDataUrl(key string) string {
	return fmt.Sprintf("wss://stream.binance.com:9443/ws/%s", key)
}

func apiPricesUrl(symbols map[string][]string) string {
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

func apiBalancesUrl(cfg *config.Config) string {
	t := time.Now().UTC().UnixMilli()
	s := signature(cfg, fmt.Sprintf("timestamp=%v", t))
	return fmt.Sprintf("https://api.binance.com/api/v3/account?timestamp=%v&signature=%s", t, s)
}

func apiListenKeyUrl() string {
	return "https://api.binance.com/api/v3/userDataStream"
}

func newOrderUrl(cfg *config.Config, symbol, side string) string {
	return fmt.Sprintf("https://api.binance.com/api/v3/order?signature=%s&symbol=%s&side=%s&type=MARKET&timestamp=%v", cfg.Binance.SecretKey, symbol, side, time.Now().UTC().UnixMilli())
}

func signature(cfg *config.Config, params string) string {
	mac := hmac.New(sha256.New, []byte(cfg.Binance.SecretKey))
	mac.Write([]byte(params))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

func stf64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.WithError(err).Panic()
	}
	return f
}
