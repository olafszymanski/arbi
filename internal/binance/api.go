package binance

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/olafszymanski/arbi/config"
	"github.com/olafszymanski/arbi/pkg/utils"
	"golang.org/x/net/http2"
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

type API struct {
	cfg         *config.Config
	factory     *URLFactory
	httpClient  *http.Client
	http2Client *http.Client
}

// TODO: Add error handling {"code":-2014,"msg":"API-key format invalid."}

func NewAPI(cfg *config.Config, factory *URLFactory) *API {
	c := &http.Client{}
	c2 := &http.Client{}
	c2.Transport = &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(netw, addr)
		}}
	return &API{cfg, factory, c, c2}
}

func (a *API) GetExchangeInfo() ([]jsonSymbol, error) {
	u := a.factory.ExchangeInfo()

	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	res, err := a.httpClient.Do(r)
	if err != nil {
		fmt.Println("err")
		return nil, err
	}
	defer res.Body.Close()

	var e jsonExchangeInfo
	if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
		return nil, err
	}
	return e.Symbols, nil
}

func (a *API) GetOrderBook() ([]jsonOrderBook, error) {
	u := a.factory.OrderBook()

	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	res, err := a.httpClient.Do(r)
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
	res, err := a.httpClient.Do(r)
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
	res, err := a.httpClient.Do(r)
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
	if _, err := a.httpClient.Do(r); err != nil {
		return err
	}
	return nil
}

func (a *API) NewOrder(symbol, side string, quantity float64, precision int) error {
	q := utils.Round(quantity, precision)
	p := fmt.Sprintf("symbol=%s&side=%s&type=MARKET&quantity=%v&recvWindow=10000&timestamp=%v", symbol, side, q, time.Now().UTC().UnixMilli())
	s := utils.Signature(a.cfg.Binance.SecretKey, p)
	u := a.factory.NewOrder(p, s)

	r, err := http.NewRequest("POST", u, nil)
	if err != nil {
		return err
	}
	r.Header.Add("X-MBX-APIKEY", a.cfg.Binance.ApiKey)
	if _, err := a.httpClient.Do(r); err != nil {
		return err
	}
	return nil
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
	if _, err := a.httpClient.Do(r); err != nil {
		return err
	}
	return nil
}
