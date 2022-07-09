package binance

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/olafszymanski/arbi/config"
	"github.com/olafszymanski/arbi/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type Symbol struct {
	Symbol         string
	Base           string
	BasePrecision  int
	Quote          string
	QuotePrecision int
	Bid            float64
	Ask            float64
}

type Wallet map[string]float64

func (w Wallet) Update(assets []jsonAsset) error {
	for _, a := range assets {
		f, err := utils.Stf(a.Free)
		if err != nil {
			return err
		}
		w[a.Asset] = f
	}
	return nil
}

type Engine struct {
	*sync.RWMutex
	cfg                *config.Config
	api                *API
	orderBookWebsocket *OrderBookWebsocket
	walletWebsocket    *WalletWebsocket
	triangles          []Triangle
	symbols            map[string]Symbol
	wallet             Wallet
}

func NewEngine(cfg *config.Config, bases []string) *Engine {
	f := NewURLFactory()
	a := NewAPI(cfg, f)
	v := NewValidator()
	c := NewAPIConverter(v)
	g := NewGenerator()

	js, err := a.GetExchangeInfo()
	if err != nil {
		log.WithError(err).Panic()
	}
	job, err := a.GetOrderBook()
	if err != nil {
		log.WithError(err).Panic()
	}
	// TODO: Fetch trading fees

	s, err := c.ToSymbols(js, job)
	if err != nil {
		log.WithError(err).Panic()
	}

	t, syms, err := g.Generate(s, bases)
	if err != nil {
		log.WithError(err).Panic()
	}

	ja, err := a.GetUserAssets()
	if err != nil {
		log.WithError(err).Panic()
	}

	w, err := c.ToWallet(ja)
	if err != nil {
		log.WithError(err).Panic()
	}

	obw, err := NewOrderBookWebsocket(f)
	if err != nil {
		log.WithError(err).Panic()
	}

	k, err := a.GetListenKey()
	if err != nil {
		log.WithError(err).Panic()
	}

	ww, err := NewWalletWebsocket(f, k)
	if err != nil {
		log.WithError(err).Panic()
	}

	// for i := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20} {
	// 	tt := time.Now()
	// 	a.NewTestOrder()
	// 	a.NewTestOrder()
	// 	a.NewTestOrder()
	// 	fmt.Println(i, ": ", time.Since(tt))
	// }

	return &Engine{&sync.RWMutex{}, cfg, a, obw, ww, t, syms, w}
}

func (e *Engine) Run() {
	d := make(chan struct{})
	i := make(chan os.Signal, 1)
	signal.Notify(i, os.Interrupt)

	c := NewWebsocketConverter()

	go func() {
		defer close(d)
		defer e.orderBookWebsocket.Close()

		for {
			o, err := e.orderBookWebsocket.Read()
			if err != nil {
				log.WithError(err).Panic()
			}

			b, a, err := c.ToPrices(o)
			if err != nil {
				log.WithError(err).Panic()
			}
			e.Lock()
			if s, ok := e.symbols[o.Symbol]; ok {
				e.symbols[o.Symbol] = Symbol{
					Symbol:         s.Symbol,
					Base:           s.Base,
					BasePrecision:  s.BasePrecision,
					Quote:          s.Quote,
					QuotePrecision: s.QuotePrecision,
					Bid:            b,
					Ask:            a,
				}
			}
			e.Unlock()
		}
	}()

	go func() {
		defer close(d)
		defer e.walletWebsocket.Close()

		for {
			b, err := e.walletWebsocket.Read()
			if err != nil {
				log.WithError(err).Panic()
			}

			for _, bal := range b {
				a, err := utils.Stf(bal.Amount)
				if err != nil {
					log.WithError(err).Panic()
				}
				e.Lock()
				e.wallet[bal.Asset] = a
				e.Unlock()
			}
		}
	}()

	for _, t := range e.triangles {
		t := t
		go func() {
			defer close(d)
			for {
				e.Lock()
				if p := e.profitability(t); p > 0 {
					e.makeTrade(t, p)
				}
				if p := e.reverseProfitability(t); p > 0 {
					e.makeReverseTrade(t, p)
				}
				e.Unlock()
			}
		}()
	}

	for {
		select {
		case <-d:
			return
		case <-i:
			select {
			case <-d:
			case <-time.After(time.Microsecond):
			}
			return
		}
	}
}

func (e *Engine) profitability(triangle Triangle) float64 {
	return 1 / e.symbols[triangle.FirstPair()].Ask * 0.999 * 1 / e.symbols[triangle.SecondPair()].Ask * 0.999 * e.symbols[triangle.ThirdPair()].Bid * 0.999
}

func (e *Engine) reverseProfitability(triangle Triangle) float64 {
	return 1 / e.symbols[triangle.ThirdPair()].Ask * 0.999 * e.symbols[triangle.SecondPair()].Bid * 0.999 * e.symbols[triangle.FirstPair()].Bid * 0.999
}

func (e *Engine) makeTrade(triangle Triangle, profitability float64) {
	// Buy - Buy - Sell
	first := triangle.Intermediate + triangle.Base
	second := triangle.Ticker + triangle.Intermediate
	third := triangle.Ticker + triangle.Base

	tt := time.Now()
	// e.api.NewOrder(first, "BUY", e.wallet[e.symbols[first].Base], e.symbols[first].BasePrecision)
	// e.api.NewOrder(second, "BUY", e.wallet[e.symbols[second].Base], e.symbols[second].BasePrecision)
	// e.api.NewOrder(first, "SELL", e.wallet[e.symbols[third].Quote], e.symbols[second].QuotePrecision)

	e.api.NewTestOrder()
	e.api.NewTestOrder()
	e.api.NewTestOrder()

	// TODO: Remove later
	p := e.profitability(triangle)

	fmt.Println("BUY", first, " ->  BUY", second, " ->  SELL", third, " = ", profitability, " | API:", time.Since(tt), " | ", p)
}

func (e *Engine) makeReverseTrade(triangle Triangle, profitability float64) {
	// Buy - Sell - Sell
	tt := time.Now()
	// 	e.api.NewOrder(third, "BUY", e.wallet[e.symbols[third].Base], e.symbols[first].BasePrecision)
	// 	e.api.NewOrder(second, "SELL", e.wallet[e.symbols[second].Quote], e.symbols[second].QuotePrecision)
	// 	e.api.NewOrder(first, "SELL", e.wallet[e.symbols[first].Quote], e.symbols[second].QuotePrecision)

	e.api.NewTestOrder()
	e.api.NewTestOrder()
	e.api.NewTestOrder()

	// TODO: Remove later
	p := e.reverseProfitability(triangle)

	fmt.Println("BUY", triangle.ThirdPair(), " ->  SELL", triangle.SecondPair(), " ->  SELL", triangle.FirstPair(), " = ", profitability, " | API:", time.Since(tt), " | ", p)
}
