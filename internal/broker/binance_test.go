package broker

import (
	"fmt"
	"testing"

	"github.com/olafszymanski/arbi/config"
)

func Test_makeWebsocketUrl(t *testing.T) {
	cfg := config.NewConfig("../../config/config.yml")
	tests := []struct {
		pair string
		want string
	}{
		{"BTCUSDT", fmt.Sprintf("%s://%s/ws/btcusdt@miniTicker", cfg.Binance.WebsocketScheme, cfg.Binance.WebsocketHost)},
		{"btcusdt", fmt.Sprintf("%s://%s/ws/btcusdt@miniTicker", cfg.Binance.WebsocketScheme, cfg.Binance.WebsocketHost)},
		{"btcUSDT", fmt.Sprintf("%s://%s/ws/btcusdt@miniTicker", cfg.Binance.WebsocketScheme, cfg.Binance.WebsocketHost)},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s, %s", tt.pair, tt.want)
		t.Run(name, func(t *testing.T) {
			url := makeWebsocketUrl(cfg, tt.pair)
			if url != tt.want {
				t.Errorf("got: %s, want %s", tt.pair, tt.want)
			}
		})
	}
}

func Test_makeApiUrl(t *testing.T) {
	cfg := config.NewConfig("../../config/config.yml")
	tests := []struct {
		symbols map[string][]string
		want    string
	}{
		{
			map[string][]string{
				"BTC": {
					"USDT",
					"USDC",
					"TUSD",
					"DAI",
				},
			},
			fmt.Sprintf(`%s://%s/api/v3/ticker/price?symbols=["BTCUSDT","BTCUSDC","BTCTUSD","BTCDAI"]`, cfg.Binance.ApiScheme, cfg.Binance.ApiHost),
		},
		{
			map[string][]string{
				"eth": {
					"usdt",
					"usdc",
					"tusd",
					"dai",
				},
			},
			fmt.Sprintf(`%s://%s/api/v3/ticker/price?symbols=["ethusdt","ethusdc","ethtusd","ethdai"]`, cfg.Binance.ApiScheme, cfg.Binance.ApiHost),
		},
		{
			map[string][]string{
				"BNB": {
					"usdt",
					"USDC",
					"tusd",
					"DAI",
				},
			},
			fmt.Sprintf(`%s://%s/api/v3/ticker/price?symbols=["BNBusdt","BNBUSDC","BNBtusd","BNBDAI"]`, cfg.Binance.ApiScheme, cfg.Binance.ApiHost),
		},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s, %s", tt.symbols, tt.want)
		t.Run(name, func(t *testing.T) {
			url := makeApiUrl(cfg, tt.symbols)
			if url != tt.want {
				t.Errorf("got %s, want %s", url, tt.want)
			}
		})
	}
}
