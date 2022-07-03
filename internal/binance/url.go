package binance

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
