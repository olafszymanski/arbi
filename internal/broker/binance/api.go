package binance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/olafszymanski/arbi/config"
	log "github.com/sirupsen/logrus"
)

type API struct {
	cfg    *config.Config
	client *http.Client
}

func NewAPI(cfg *config.Config) *API {
	return &API{cfg, &http.Client{}}
}

func (a *API) ReadPrices(symbols map[string][]string) []Price {
	type tempPrice struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}

	data, err := http.Get(apiPricesUrl(symbols))
	if err != nil {
		log.WithError(err).Panic()
	}
	defer data.Body.Close()
	body, err := ioutil.ReadAll(data.Body)
	if err != nil {
		log.WithError(err).Panic()
	}

	var tmpRes []tempPrice
	if err := json.Unmarshal(body, &tmpRes); err != nil {
		log.WithError(err).Panic()
	}
	var res []Price
	for _, r := range tmpRes {
		res = append(res, Price{r.Symbol, stf64(r.Price)})
	}
	return res
}

func (a *API) ReadBalances(symbols map[string][]string) []Balance {
	type tempBalance struct {
		Asset  string `json:"asset"`
		Amount string `json:"free"`
	}
	type tempBalances struct {
		Balances []tempBalance `json:"balances"`
	}

	req, _ := http.NewRequest("GET", apiBalancesUrl(a.cfg), nil)
	req.Header.Add("X-MBX-APIKEY", a.cfg.Binance.ApiKey)
	data, err := a.client.Do(req)
	if err != nil {
		log.WithError(err).Panic()
	}
	defer data.Body.Close()
	body, err := ioutil.ReadAll(data.Body)
	if err != nil {
		log.WithError(err).Panic()
	}

	var tmpRes tempBalances
	if err := json.Unmarshal(body, &tmpRes); err != nil {
		log.WithError(err).Panic()
	}
	var res []Balance
	for crp, stbs := range symbols {
		for _, r := range tmpRes.Balances {
			amt := stf64(r.Amount)
			if r.Asset == crp {
				res = append(res, Balance{r.Asset, amt})
			} else {
				for _, s := range stbs {
					if r.Asset == s {
						res = append(res, Balance{r.Asset, amt})
					}
				}
			}
		}
	}
	return res
}

func (a *API) ReadListenKey() string {
	type keyResponse struct {
		Key string `json:"listenKey"`
	}

	req, _ := http.NewRequest("POST", apiListenKeyUrl(), nil)
	req.Header.Add("X-MBX-APIKEY", a.cfg.Binance.ApiKey)
	data, err := a.client.Do(req)
	if err != nil {
		log.WithError(err).Panic()
	}
	defer data.Body.Close()
	body, err := ioutil.ReadAll(data.Body)
	if err != nil {
		log.WithError(err).Panic()
	}

	var res keyResponse
	if err := json.Unmarshal(body, &res); err != nil {
		log.WithError(err).Panic()
	}
	return res.Key
}

func (a *API) ExtendListenKey() {
	req, _ := http.NewRequest("PUT", apiListenKeyUrl(), nil)
	req.Header.Add("X-MBX-APIKEY", a.cfg.Binance.ApiKey)
	_, err := a.client.Do(req)
	if err != nil {
		log.WithError(err).Panic()
	}
}

func (a *API) NewOrder(symbol, side string) {
	fmt.Println(symbol)
	req, _ := http.NewRequest("POST", newOrderUrl(a.cfg, symbol, side), nil)
	req.Header.Add("X-MBX-APIKEY", a.cfg.Binance.ApiKey)
	data, err := a.client.Do(req)
	if err != nil {
		log.WithError(err).Panic()
	}
	defer data.Body.Close()
	body, err := ioutil.ReadAll(data.Body)
	if err != nil {
		log.WithError(err).Panic()
	}
	fmt.Println(string(body))
}
