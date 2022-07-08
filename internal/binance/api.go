package binance

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/olafszymanski/arbi/config"
	"github.com/olafszymanski/arbi/pkg/utils"
	"github.com/valyala/fasthttp"
)

type jsonSymbol struct {
	Symbol         string   `json:"symbol"`
	Base           string   `json:"baseAsset"`
	BasePrecision  int      `json:"baseAssetPrecision"`
	Quote          string   `json:"quoteAsset"`
	QuotePrecision int      `json:"quoteAssetPrecision"`
	Permissions    []string `json:"permissions"`
}

type jsonExchangeInfo struct {
	Symbols []jsonSymbol `json:"symbols"`
}

type jsonOrderBook struct {
	Symbol string `json:"symbol"`
	Bid    string `json:"bidPrice"`
	Ask    string `json:"askPrice"`
}

type jsonAsset struct {
	Asset string `json:"asset"`
	Free  string `json:"free"`
}

type jsonListenKey struct {
	Key string `json:"listenKey"`
}

type API struct {
	cfg             *config.Config
	factory         *URLFactory
	request         *fasthttp.Request
	newOrderRequest *fasthttp.Request
}

func NewAPI(cfg *config.Config, factory *URLFactory) *API {
	r := fasthttp.AcquireRequest()
	r.Header.SetMethod("POST")
	r.Header.Add("X-MBX-APIKEY", cfg.Binance.ApiKey)
	return &API{cfg, factory, fasthttp.AcquireRequest(), r}
}

func (a *API) GetExchangeInfo() ([]jsonSymbol, error) {
	u := a.factory.ExchangeInfo()

	a.request.Header.SetMethod("GET")
	a.request.SetRequestURI(u)
	r := fasthttp.Response{}
	if err := fasthttp.Do(a.request, &r); err != nil {
		return nil, err
	}

	var e jsonExchangeInfo
	if err := json.Unmarshal(r.Body(), &e); err != nil {
		return nil, err
	}
	return e.Symbols, nil
}

func (a *API) GetOrderBook() ([]jsonOrderBook, error) {
	u := a.factory.OrderBook()

	a.request.Header.SetMethod("GET")
	a.request.SetRequestURI(u)
	r := fasthttp.Response{}
	if err := fasthttp.Do(a.request, &r); err != nil {
		return nil, err
	}

	var o []jsonOrderBook
	if err := json.Unmarshal(r.Body(), &o); err != nil {
		return nil, err
	}
	return o, nil
}

func (a *API) GetUserAssets() ([]jsonAsset, error) {
	p := fmt.Sprintf("timestamp=%v", time.Now().UTC().UnixMilli())
	s := utils.Signature(a.cfg.Binance.SecretKey, p)
	u := a.factory.UserAssets(p, s)

	a.request.Header.SetMethod("POST")
	a.request.Header.Add("X-MBX-APIKEY", a.cfg.Binance.ApiKey)
	a.request.SetRequestURI(u)
	r := fasthttp.Response{}
	if err := fasthttp.Do(a.request, &r); err != nil {
		return nil, err
	}

	var as []jsonAsset
	if err := json.Unmarshal(r.Body(), &as); err != nil {
		return nil, err
	}
	return as, nil
}

func (a *API) GetListenKey() (string, error) {
	u := a.factory.ListenKey("")

	a.request.Header.SetMethod("POST")
	a.request.SetRequestURI(u)
	r := fasthttp.Response{}
	if err := fasthttp.Do(a.request, &r); err != nil {
		return "", err
	}

	var l jsonListenKey
	if err := json.Unmarshal(r.Body(), &l); err != nil {
		return "", err
	}
	return l.Key, nil
}

func (a *API) KeepAliveListenKey(listenKey string) error {
	u := a.factory.ListenKey(listenKey)

	a.request.Header.SetMethod("PUT")
	a.request.SetRequestURI(u)
	if err := fasthttp.Do(a.request, nil); err != nil {
		return err
	}
	return nil
}

func (a *API) NewOrder(symbol, side string, quantity float64, precision int) (bool, error) {
	q := utils.Round(quantity, precision)
	p := fmt.Sprintf("symbol=%s&side=%s&type=MARKET&quantity=%v&recvWindow=10000&timestamp=%v", symbol, side, q, time.Now().UTC().UnixMilli())
	s := utils.Signature(a.cfg.Binance.SecretKey, p)
	u := a.factory.NewOrder(p, s)

	a.newOrderRequest.SetRequestURI(u)
	if err := fasthttp.Do(a.request, nil); err != nil {
		return false, err
	}
	return true, nil
}

func (a *API) NewTestOrder() (bool, error) {
	p := fmt.Sprintf("symbol=BTCUSDT&side=BUY&type=MARKET&quantity=1&recvWindow=10000&timestamp=%v", time.Now().UTC().UnixMilli())
	s := utils.Signature(a.cfg.Binance.SecretKey, p)
	u := a.factory.NewTestOrder(p, s)

	a.newOrderRequest.SetRequestURI(u)
	if err := fasthttp.Do(a.request, nil); err != nil {
		return false, err
	}
	return true, nil
}
