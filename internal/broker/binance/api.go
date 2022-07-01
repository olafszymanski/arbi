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

func (a *API) ReadPrices(symbols map[string][]string) []PairResult {
	type tempResult struct {
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

	var tmpRes []tempResult
	if err := json.Unmarshal(body, &tmpRes); err != nil {
		log.WithError(err).Panic()
	}
	var res []PairResult
	for _, r := range tmpRes {
		res = append(res, PairResult(r))
	}
	return res
}

func (a *API) ReadBalances(symbols map[string][]string) []BalanceResult {
	type tempBalanceResult struct {
		Balances []BalanceResult `json:"balances"`
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

	var tmpRes tempBalanceResult
	if err := json.Unmarshal(body, &tmpRes); err != nil {
		log.WithError(err).Panic()
	}
	var res []BalanceResult
	for crp, stbs := range symbols {
		for _, r := range tmpRes.Balances {
			if r.Asset == crp {
				res = append(res, BalanceResult{r.Asset, r.Free})
			} else {
				for _, s := range stbs {
					if r.Asset == s {
						res = append(res, BalanceResult{r.Asset, r.Free})
					}
				}
			}
		}
	}
	return res
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
