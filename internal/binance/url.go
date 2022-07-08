package binance

import (
	"fmt"
)

type URLFactory struct {
}

func NewURLFactory() *URLFactory {
	return &URLFactory{}
}

func (u *URLFactory) ExchangeInfo() string {
	return "https://api.binance.com/api/v3/exchangeInfo"
}

func (u *URLFactory) OrderBook() string {
	return "https://api.binance.com/api/v3/ticker/bookTicker"
}

func (u *URLFactory) OrderBookTickers() string {
	return "wss://stream.binance.com:9443/ws/!bookTicker"
}

func (u *URLFactory) UserAssets(params, signature string) string {
	return fmt.Sprintf("https://api.binance.com/sapi/v3/asset/getUserAsset?%s&signature=%s", params, signature)
}

func (u *URLFactory) ListenKey(listenKey string) string {
	if listenKey != "" {
		return fmt.Sprintf("https://api.binance.com/api/v3/userDataStream?listenKey=%s", listenKey)
	}
	return "https://api.binance.com/api/v3/userDataStream"
}

func (u *URLFactory) AccountUpdate(listenKey string) string {
	return fmt.Sprintf("wss://stream.binance.com:9443/ws/%s", listenKey)
}

func (u *URLFactory) NewOrder(params, signature string) string {
	return fmt.Sprintf("https://api.binance.com/api/v3/order?%s&signature=%s", params, signature)
}

func (u *URLFactory) NewTestOrder(params, signature string) string {
	return fmt.Sprintf("https://api.binance.com/api/v3/order/test?%s&signature=%s", params, signature)
}
