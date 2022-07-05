package binance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/olafszymanski/arbi/config"
	"github.com/olafszymanski/arbi/pkg/utils"
	"github.com/valyala/fasthttp"
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
	request *fasthttp.Request
}

func NewAPI(cfg *config.Config, factory *URLFactory) *API {
	r := fasthttp.AcquireRequest()
	r.Header.SetMethod("POST")
	r.Header.Add("X-MBX-APIKEY", cfg.Binance.ApiKey)
	return &API{cfg, factory, &http.Client{}, r}
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

func (a *API) NewTestOrder() (bool, error) {
	p := fmt.Sprintf("symbol=BTCUSDT&side=BUY&type=MARKET&quantity=1&recvWindow=10000&timestamp=%v", time.Now().UTC().UnixMilli())
	s := utils.Signature(a.cfg.Binance.SecretKey, p)
	u := a.factory.NewTestOrder(p, s)

	a.request.SetRequestURI(u)
	if err := fasthttp.Do(a.request, nil); err != nil {
		return false, err
	}
	return true, nil
}
