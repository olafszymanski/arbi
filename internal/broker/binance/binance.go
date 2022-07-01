package binance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/olafszymanski/arbi/config"
	"github.com/olafszymanski/arbi/internal/broker"
	"github.com/olafszymanski/arbi/internal/database"
	log "github.com/sirupsen/logrus"
)

type Price struct {
	Symbol string
	Price  float64
}

type Balance struct {
	Asset  string
	Amount float64
}

type Binance struct {
	cfg     *config.Config
	lock    sync.RWMutex
	pairs   broker.Pairs
	account broker.Account
	store   *database.Store
	api     *API
	blocked bool
}

func New(cfg *config.Config, store *database.Store, symbols map[string][]string) *Binance {
	api := NewAPI(cfg)
	prcs := api.ReadPrices(symbols)
	prs := make(broker.Pairs)
	for key, syms := range symbols {
		for _, sym := range syms {
			s := key + sym
			for _, pr := range prcs {
				if pr.Symbol == s {
					prs[s] = broker.Pair{
						Crypto: key,
						Stable: sym,
						Price:  pr.Price,
					}
				}
			}
		}
	}

	blcs := api.ReadBalances(symbols)
	acc := make(broker.Account, 0)
	for _, blc := range blcs {
		acc[blc.Asset] = blc.Amount
	}
	fmt.Println(acc)
	return &Binance{
		cfg:     cfg,
		pairs:   prs,
		account: acc,
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
				res := ws.ReadPrice()
				b.lock.Lock()
				b.pairs[sym] = broker.Pair{
					Crypto: pr.Crypto,
					Stable: pr.Stable,
					Price:  res.Price,
				}
				high, low := b.pairs.HighestLowest(pr.Crypto)
				b.lock.Unlock()

				val := broker.Profitability(&high, &low, b.cfg.Binance.Fee, b.cfg.Binance.Conversion)
				if val > b.cfg.Binance.MinProfit && b.cfg.App.UseDB < 1 && !b.blocked {
					b.lock.Lock()
					b.blocked = true

					// b.api.NewOrder(high.Crypto+high.Stable, b.account[high.Crypto], "SELL")
					// if low.Stable == "dai" {
					//     	b.api.NewOrder(high.Stable+low.Stable, b.account[high.Stable], "SELL")
					// } else {
					//		b.api.NewOrder(low.Stable+high.Stable, b.account[low.Stable], "BUY")
					// }
					// b.api.NewOrder(high.Crypto+low.Stable, b.account[low.Stable], "BUY")

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
