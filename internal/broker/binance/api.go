package binance

import (
	"encoding/json"
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

func (a *API) Read(symbols map[string][]string) []Result {
	type tempResult struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}

	data, err := http.Get(apiUrl(symbols))
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
	var res []Result
	for _, r := range tmpRes {
		res = append(res, Result(r))
	}
	return res
}
