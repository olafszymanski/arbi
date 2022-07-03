package binance

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/olafszymanski/arbi/config"
)

type jsonSymbol struct {
	Symbol      string   `json:"symbol"`
	Base        string   `json:"baseAsset"`
	Quote       string   `json:"quoteAsset"`
	Precision   uint8    `json:"quoteAssetPrecision"`
	Permissions []string `json:"permissions"`
}

type jsonExchangeInfo struct {
	Symbols []jsonSymbol `json:"symbols"`
}

type jsonOrderBook struct {
	Symbol string `json:"symbol"`
	Bid    string `json:"bidPrice"`
	Ask    string `json:"askPrice"`
}

type API struct {
	cfg     *config.Config
	factory *URLFactory
	client  *http.Client
}

func NewAPI(cfg *config.Config) *API {
	return &API{cfg, NewURLFactory(), &http.Client{}}
}

func (a *API) GetExchangeInfo() ([]jsonSymbol, error) {
	u := a.factory.ExchangeInfo()
	r, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	d, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var e jsonExchangeInfo
	if err := json.Unmarshal(d, &e); err != nil {
		return nil, err
	}
	return e.Symbols, nil
}

func (a *API) GetOrderBook() ([]jsonOrderBook, error) {
	u := a.factory.OrderBook()
	r, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	d, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var o []jsonOrderBook
	if err := json.Unmarshal(d, &o); err != nil {
		return nil, err
	}
	return o, nil
}
