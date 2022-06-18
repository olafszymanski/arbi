package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/olafszymanski/arbi/config"
	"github.com/olafszymanski/arbi/internal/database"
	"github.com/olafszymanski/arbi/internal/exchange"
	log "github.com/sirupsen/logrus"
)

type Binance struct {
	cfg   *config.Config
	lock  sync.RWMutex
	prs   exchange.Pairs
	store *database.Store
	in    bool
}

func NewBinance(cfg *config.Config, s *database.Store, symbols map[string][]string) *Binance {
	res, err := http.Get(makeApiUrl(cfg, symbols))
	if err != nil {
		log.WithError(err).Panic()
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.WithError(err).Panic()
	}

	type result struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	var prcs []result
	json.Unmarshal(body, &prcs)

	prs := make(exchange.Pairs)
	for key, syms := range symbols {
		for _, sym := range syms {
			s := key + sym
			for _, pr := range prcs {
				if pr.Symbol == s {
					prc, err := strconv.ParseFloat(pr.Price, 64)
					if err != nil {
						log.WithError(err).Panic()
					}
					prs[s] = exchange.Pair{
						Crypto: key,
						Stable: sym,
						Price:  prc,
					}
				}
			}
		}
	}

	return &Binance{
		cfg:   cfg,
		prs:   prs,
		store: s,
		in:    false,
	}
}

func (b *Binance) Subscribe() {
	type result struct {
		Symbol string `json:"s"`
		Price  string `json:"c"`
	}

	for sym, pr := range b.prs {
		sym := sym
		pr := pr
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.WithError(r.(error)).Error("Goroutine - '", sym, "' - panicked")
				}
			}()

			conn, _, err := websocket.DefaultDialer.Dial(makeWebsocketUrl(b.cfg, sym), nil)
			log.Info("Connecting to '", sym, "' websocket...")
			if err != nil {
				log.WithError(err).Panic()
			}
			defer conn.Close()

			for {
				_, data, err := conn.ReadMessage()
				if err != nil {
					log.WithError(err).Panic()
				}

				var res result
				if err := json.Unmarshal(data, &res); err != nil {
					log.WithError(err).Panic()
				}

				prc, err := strconv.ParseFloat(res.Price, 64)
				if err != nil {
					log.WithError(err).Panic()
				}
				b.lock.Lock()
				b.prs[sym] = exchange.Pair{
					Crypto: pr.Crypto,
					Stable: pr.Stable,
					Price:  prc,
				}
				high, low := b.prs.HighestLowest(pr.Crypto)
				b.lock.Unlock()

				val := exchange.Profitability(&high, &low, b.cfg.Binance.Fee, b.cfg.Binance.Conversion)
				log.WithFields(log.Fields{
					"high": high,
					"low":  low,
					"val":  val,
				}).Info("Websocket received")
				if val > b.cfg.Binance.MinProfit && b.cfg.App.UseDB > 0 && !b.in {
					b.lock.Lock()
					b.in = true

					if err := b.store.AddRecord(context.Background(), &high, &low, val); err != nil {
						log.WithError(err).Panic()
					}
					time.Sleep(time.Second * 5)

					b.in = false
					b.lock.Unlock()
				}
			}
		}()
	}
}

func makeWebsocketUrl(cfg *config.Config, pair string) string {
	return fmt.Sprintf("%s://%s/ws/%s@miniTicker", cfg.Binance.WebsocketScheme, cfg.Binance.WebsocketHost, strings.ToLower(pair))
}

func makeApiUrl(cfg *config.Config, symbols map[string][]string) string {
	var url strings.Builder
	url.WriteString(cfg.Binance.ApiScheme + "://" + cfg.Binance.ApiHost + "/api/v3/ticker/price?symbols=[")
	i := 0
	for key, syms := range symbols {
		for j, s := range syms {
			if j == len(syms)-1 {
				url.WriteString(`"` + key + s + `"`)
			} else {
				url.WriteString(`"` + key + s + `",`)
			}
		}
		if i != len(symbols)-1 {
			url.WriteString(",")
		}
		i++
	}
	url.WriteString("]")
	return url.String()
}
