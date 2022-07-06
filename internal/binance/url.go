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
	return "https://api.binance.us/api/v3/exchangeInfo"
}

func (u *URLFactory) OrderBook() string {
	return "https://api.binance.us/api/v3/ticker/bookTicker"
}

func (u *URLFactory) OrderBookTickers() string {
	return "wss://stream.binance.us:9443/ws/!bookTicker"
}

func (u *URLFactory) NewTestOrder(params, signature string) string {
	return fmt.Sprintf("https://api.binance.us/api/v3/order/test?%s&signature=%s", params, signature)
}
