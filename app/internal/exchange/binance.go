package exchange

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/olafszymanski/arbi/app/config"
	"github.com/olafszymanski/arbi/app/internal/postgres"
	"github.com/rs/zerolog"
)

type Binance struct {
	l     *zerolog.Logger
	cfg   *config.Config
	lock  sync.RWMutex
	prs   Pairs
	store *postgres.Store
	in    bool
}

func NewBinance(l *zerolog.Logger, cfg *config.Config, s *postgres.Store, symbols map[string][]string) *Binance {
	res, err := http.Get(makeApiUrl(cfg, symbols))
	if err != nil {
		l.Panic().Msg(err.Error())
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		l.Panic().Msg(err.Error())
	}

	type Result struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	var prcs []Result
	json.Unmarshal(body, &prcs)

	prs := make(Pairs)
	for key, syms := range symbols {
		for _, sym := range syms {
			s := key + sym
			for _, pr := range prcs {
				if pr.Symbol == s {
					prc, err := strconv.ParseFloat(pr.Price, 64)
					if err != nil {
						l.Panic().Msg(err.Error())
					}
					prs[s] = Pair{key, sym, prc}
				}
			}
		}
	}

	return &Binance{
		l:     l,
		cfg:   cfg,
		prs:   prs,
		store: s,
		in:    false,
	}
}

func (b *Binance) Subscribe() {
	type Result struct {
		Symbol string `json:"s"`
		Price  string `json:"c"`
	}

	for sym, pr := range b.prs {
		sym := sym
		pr := pr
		go func() {
			defer func() {
				if r := recover(); r != nil {
					b.l.Error().Msg(fmt.Sprintf("Goroutine: %s panicked, error: %v", sym, r))
				}
			}()

			conn, _, err := websocket.DefaultDialer.Dial(makeWebsocketUrl(b.cfg, sym), nil)
			b.l.Info().Msg(fmt.Sprintf("Connecting to %s websocket...", sym))
			if err != nil {
				b.l.Panic().Msg(err.Error())
			}
			defer conn.Close()

			for {
				_, data, err := conn.ReadMessage()
				if err != nil {
					b.l.Panic().Msg(err.Error())
				}

				var res Result
				if err := json.Unmarshal(data, &res); err != nil {
					b.l.Panic().Msg(err.Error())
				}

				prc, err := strconv.ParseFloat(res.Price, 64)
				if err != nil {
					b.l.Panic().Msg(err.Error())
				}
				b.lock.Lock()
				b.prs[sym] = Pair{pr.Crypto, pr.Stable, prc}
				high, low := b.prs.HighestLowest(pr.Crypto)
				b.lock.Unlock()

				val := Profitability(&high, &low, b.cfg.Binance.Fee, b.cfg.Binance.Conversion)
				b.l.Info().Str("High", high.String()).Str("Low", low.String()).Str("Val", fmt.Sprintf("%f", val)).Msg("Websocket received")
				if val > b.cfg.Binance.MinProfit && b.cfg.App.UseDB > 0 && !b.in {
					b.lock.Lock()
					b.in = true

					rec := postgres.Record{
						Low: postgres.RecordPair{
							Symbol: low.Crypto + low.Stable,
							Price:  low.Price,
						},
						High: postgres.RecordPair{
							Symbol: high.Crypto + high.Stable,
							Price:  high.Price,
						},
						Value:     val,
						Timestamp: time.Now(),
					}
					if err := b.store.Binance.CreateRecord(rec); err != nil {
						b.l.Panic().Msg(err.Error())
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
