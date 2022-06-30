package binance

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/olafszymanski/arbi/config"
	"github.com/olafszymanski/arbi/internal/broker"
	"github.com/olafszymanski/arbi/internal/database"
	log "github.com/sirupsen/logrus"
)

type Result struct {
	Symbol string `json:"s"`
	Price  string `json:"c"`
}

type Binance struct {
	cfg     *config.Config
	lock    sync.RWMutex
	pairs   broker.Pairs
	store   *database.Store
	api     *API
	blocked bool
}

func New(cfg *config.Config, store *database.Store, symbols map[string][]string) *Binance {
	api := NewAPI(cfg)
	res := api.Read(symbols)
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
		cfg:     cfg,
		pairs:   prs,
		store:   store,
		api:     api,
		blocked: false,
	}
}

func (b *Binance) Subscribe(ctx context.Context, done chan struct{}) {
	for sym, pr := range b.pairs {
		sym := sym
		pr := pr

		go func() {
			defer close(done)
			defer func() {
				if r := recover(); r != nil {
					log.Error("Goroutine - '", sym, "' - panicked")
				}
			}()

			ws := NewWebsocket(b.cfg, sym)
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
				if val > b.cfg.Binance.MinProfit && b.cfg.App.UseDB < 1 && !b.blocked {
					b.lock.Lock()
					b.blocked = true

					// b.api.NewOrder(high.Crypto+high.Stable, "SELL")
					// b.api.NewOrder(high.Stable+low.Stable, "BUY")
					// b.api.NewOrder(high.Crypto+low.Stable, "SELL")

					b.store.PushRecord(&high, &low, val)
					log.WithFields(log.Fields{
						"high": high,
						"low":  low,
						"val":  val,
					}).Info("Pushed to store queue")

					time.Sleep(time.Second * time.Duration(b.cfg.Binance.Cooldown))

					b.blocked = false
					b.lock.Unlock()
				}

				if err := b.store.Commit(ctx); err != nil {
					log.WithError(err).Panic()
				}
			}
		}()
	}
}
