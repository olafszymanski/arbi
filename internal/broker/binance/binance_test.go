package binance

import (
	"fmt"
	"testing"
)

func Test_makeWebsocketUrl(t *testing.T) {
	tests := []struct {
		pair string
		want string
	}{
		{"BTCUSDT", "wss://stream.binance.com/ws/btcusdt@miniTicker"},
		{"btcusdt", "wss://stream.binance.com/ws/btcusdt@miniTicker"},
		{"btcUSDT", "wss://stream.binance.com/ws/ws/btcusdt@miniTicker"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s, %s", tt.pair, tt.want)
		t.Run(name, func(t *testing.T) {
			url := makeWebsocketUrl(tt.pair)
			if url != tt.want {
				t.Errorf("got: %s, want %s", tt.pair, tt.want)
			}
		})
	}
}

func Test_makeApiUrl(t *testing.T) {
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
			`https://api.binance.com/api/v3/ticker/price?symbols=["BTCUSDT","BTCUSDC","BTCTUSD","BTCDAI"]`,
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
			`https://api.binance.com/api/v3/ticker/price?symbols=["ethusdt","ethusdc","ethtusd","ethdai"]`,
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
			`https://api.binance.com/api/v3/ticker/price?symbols=["BNBusdt","BNBUSDC","BNBtusd","BNBDAI"]`,
		},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s, %s", tt.symbols, tt.want)
		t.Run(name, func(t *testing.T) {
			url := makeApiUrl(tt.symbols)
			if url != tt.want {
				t.Errorf("got %s, want %s", url, tt.want)
			}
		})
	}
}
