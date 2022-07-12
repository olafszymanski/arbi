package binance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/olafszymanski/arbi/config"
	"github.com/olafszymanski/arbi/pkg/utils"
)

type jsonLotSizeFilter struct {
	Type      string `json:"filterType"`
	Precision string `json:"stepSize"`
}

type jsonSymbol struct {
	Symbol      string              `json:"symbol"`
	Base        string              `json:"baseAsset"`
	Quote       string              `json:"quoteAsset"`
	Permissions []string            `json:"permissions"`
	Filters     []jsonLotSizeFilter `json:"filters"`
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

type jsonOrder struct {
	Symbol     string `json:"symbol"`
	Quantity   string `json:"executedQty"`
	Commission string `json:"commission"`
}

type API struct {
	cfg     *config.Config
	factory *URLFactory
	client  *http.Client
}

func NewAPI(cfg *config.Config, factory *URLFactory) *API {
	return &API{cfg, factory, &http.Client{}}
}

func (a *API) GetExchangeInfo() ([]jsonSymbol, error) {
	u := a.factory.ExchangeInfo()

	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	res, err := a.client.Do(r)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var ei jsonExchangeInfo
	if err := json.NewDecoder(res.Body).Decode(&ei); err != nil {
		return nil, err
	}
	return ei.Symbols, nil
}

func (a *API) GetOrderBook() ([]jsonOrderBook, error) {
	u := a.factory.OrderBook()

	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	res, err := a.client.Do(r)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var o []jsonOrderBook
	if err := json.NewDecoder(res.Body).Decode(&o); err != nil {
		return nil, err
	}
	return o, nil
}

func (a *API) GetUserAssets() ([]jsonAsset, error) {
	p := fmt.Sprintf("timestamp=%v", time.Now().UTC().UnixMilli())
	s := utils.Signature(a.cfg.Binance.SecretKey, p)
	u := a.factory.UserAssets(p, s)

	r, err := http.NewRequest("POST", u, nil)
	if err != nil {
		return nil, err
	}
	r.Header.Add("X-MBX-APIKEY", a.cfg.Binance.ApiKey)
	res, err := a.client.Do(r)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var as []jsonAsset
	if err := json.NewDecoder(res.Body).Decode(&as); err != nil {
		return nil, err
	}
	return as, nil
}

func (a *API) GetListenKey() (string, error) {
	u := a.factory.ListenKey("")

	r, err := http.NewRequest("POST", u, nil)
	if err != nil {
		return "", err
	}
	r.Header.Add("X-MBX-APIKEY", a.cfg.Binance.ApiKey)
	res, err := a.client.Do(r)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var l jsonListenKey
	if err := json.NewDecoder(res.Body).Decode(&l); err != nil {
		return "", err
	}
	return l.Key, nil
}

func (a *API) KeepAliveListenKey(listenKey string) error {
	u := a.factory.ListenKey(listenKey)

	r, err := http.NewRequest("PUT", u, nil)
	if err != nil {
		return err
	}
	if _, err := a.client.Do(r); err != nil {
		return err
	}
	return nil
}

func (a *API) NewOrder(symbol, side string, quantity float64, precision int) (*jsonOrder, error) {
	q := utils.Round(quantity, precision)
	p := fmt.Sprintf("symbol=%s&side=%s&type=MARKET&quantity=%v&recvWindow=10000&timestamp=%v", symbol, side, q, time.Now().UTC().UnixMilli())
	s := utils.Signature(a.cfg.Binance.SecretKey, p)
	u := a.factory.NewOrder(p, s)

	r, err := http.NewRequest("POST", u, nil)
	if err != nil {
		return nil, err
	}
	r.Header.Add("X-MBX-APIKEY", a.cfg.Binance.ApiKey)
	res, err := a.client.Do(r)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var o jsonOrder
	if err := json.Unmarshal(b, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

func (a *API) NewTestOrder() error {
	p := fmt.Sprintf("symbol=BTCUSDT&side=BUY&type=MARKET&quantity=1&recvWindow=10000&timestamp=%v", time.Now().UTC().UnixMilli())
	s := utils.Signature(a.cfg.Binance.SecretKey, p)
	u := a.factory.NewTestOrder(p, s)

	r, err := http.NewRequest("POST", u, nil)
	if err != nil {
		return err
	}
	r.Header.Add("X-MBX-APIKEY", a.cfg.Binance.ApiKey)
	if _, err := a.client.Do(r); err != nil {
		return err
	}
	return nil
}
