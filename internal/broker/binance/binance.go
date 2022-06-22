package binance

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
	"github.com/olafszymanski/arbi/internal/broker"
	"github.com/olafszymanski/arbi/internal/database"
	log "github.com/sirupsen/logrus"
)

type BinanceResult struct {
	Symbol string `json:"s"`
	Price  string `json:"c"`
}

type BinanceWebsocket struct {
	conn *websocket.Conn
}

func NewBinanceWebsocket(cfg *config.Config, symbol string) *BinanceWebsocket {
	conn, _, err := websocket.DefaultDialer.Dial(makeWebsocketUrl(symbol), nil)
	if err != nil {
		log.WithError(err).Panic()
	}
	return &BinanceWebsocket{conn}
}

func (b *BinanceWebsocket) Read() BinanceResult {
	_, data, err := b.conn.ReadMessage()
	if err != nil {
		log.WithError(err).Panic()
	}

	var res BinanceResult
	if err := json.Unmarshal(data, &res); err != nil {
		log.WithError(err).Panic()
	}
	return res
}

func (b *BinanceWebsocket) Close() {
	b.conn.Close()
}

type BinancePricesAPI struct {
}

func NewBinancePricesAPI() *BinancePricesAPI {
	return &BinancePricesAPI{}
}

func (b *BinancePricesAPI) Read(cfg *config.Config, symbols map[string][]string) []BinanceResult {
	type result struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}

	data, err := http.Get(makeApiUrl(symbols))
	if err != nil {
		log.WithError(err).Panic()
	}
	defer data.Body.Close()
	body, err := ioutil.ReadAll(data.Body)
	if err != nil {
		log.WithError(err).Panic()
	}

	var tmpRes []result
	if err := json.Unmarshal(body, &tmpRes); err != nil {
		log.WithError(err).Panic()
	}
	var res []BinanceResult
	for _, r := range tmpRes {
		res = append(res, BinanceResult(r))
	}
	return res
}

type Binance struct {
	cfg   *config.Config
	lock  sync.RWMutex
	pairs broker.Pairs
	store *database.Store
	in    bool
}

func NewBinance(cfg *config.Config, store *database.Store, symbols map[string][]string) *Binance {
	api := NewBinancePricesAPI()
	res := api.Read(cfg, symbols)
	prs := make(broker.Pairs)
	for key, syms := range symbols {
		for _, sym := range syms {
			s := key + sym
			for _, pr := range res {
				if pr.Symbol == s {
					prc, err := strconv.ParseFloat(pr.Price, 64)
					if err != nil {
						log.WithError(err).Panic()
					}
					prs[s] = broker.Pair{
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
		pairs: prs,
		store: store,
		in:    false,
	}
}

func (b *Binance) Subscribe() {
	for sym, pr := range b.pairs {
		sym := sym
		pr := pr
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.WithError(r.(error)).Error("Goroutine - '", sym, "' - panicked")
				}
			}()

			ws := NewBinanceWebsocket(b.cfg, sym)
			defer ws.Close()
			log.Info("Connected to '", sym, "' websocket")

			for {
				res := ws.Read()
				prc, err := strconv.ParseFloat(res.Price, 64)
				if err != nil {
					log.WithError(err).Panic()
				}
				b.lock.Lock()
				b.pairs[sym] = broker.Pair{
					Crypto: pr.Crypto,
					Stable: pr.Stable,
					Price:  prc,
				}
				high, low := b.pairs.HighestLowest(pr.Crypto)
				b.lock.Unlock()

				val := broker.Profitability(&high, &low, b.cfg.Binance.Fee, b.cfg.Binance.Conversion)
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

func makeWebsocketUrl(pair string) string {
	return fmt.Sprintf("wss://stream.binance.com/ws/%s@miniTicker", strings.ToLower(pair))
}

func makeApiUrl(symbols map[string][]string) string {
	url := "https://api.binance.com/api/v3/ticker/price?symbols=["
	tmpSyms := make([]string, 0, len(symbols))
	for crypto, stables := range symbols {
		for _, stable := range stables {
			tmpSyms = append(tmpSyms, fmt.Sprintf(`"%s"`, crypto+stable))
		}
	}
	syms := strings.Join(tmpSyms, ",")
	url += syms + "]"
	return url
}
