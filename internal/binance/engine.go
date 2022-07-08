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
	Symbol    string
	Base      string
	Quote     string
	Precision uint8
	Bid       float64
	Ask       float64
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

func NewEngine(cfg *config.Config, bases []string) (*Engine, error) {
	f := NewURLFactory()
	a := NewAPI(cfg, f)
	v := NewValidator()
	c := NewAPIConverter(v)
	g := NewGenerator()

	js, err := a.GetExchangeInfo()
	if err != nil {
		return nil, err
	}
	job, err := a.GetOrderBook()
	if err != nil {
		return nil, err
	}

	s, err := c.ToSymbols(js, job)
	if err != nil {
		return nil, err
	}

	t, syms, err := g.Generate(s, bases)
	if err != nil {
		return nil, err
	}

	ja, err := a.GetUserAssets()
	if err != nil {
		return nil, err
	}

	w, err := c.ToWallet(ja)
	if err != nil {
		return nil, err
	}

	obw, err := NewOrderBookWebsocket(f)
	if err != nil {
		return nil, err
	}

	// ww, err := NewWalletWebsocket(f)
	// if err != nil {
	// 	return nil, err
	// }

	for i := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20} {
		tt := time.Now()
		a.NewTestOrder()
		a.NewTestOrder()
		a.NewTestOrder()
		fmt.Println(i, ": ", time.Since(tt))
	}

	return &Engine{&sync.RWMutex{}, cfg, a, obw, nil, t, syms, w}, nil
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
			j, err := e.orderBookWebsocket.Read()
			if err != nil {
				log.WithError(err).Panic()
			}

			b, a, err := c.ToPrices(j)
			if err != nil {
				log.WithError(err).Panic()
			}
			e.Lock()
			if s, ok := e.symbols[j.Symbol]; ok {
				e.symbols[j.Symbol] = Symbol{
					Symbol:    s.Symbol,
					Base:      s.Base,
					Quote:     s.Quote,
					Precision: s.Precision,
					Bid:       b,
					Ask:       a,
				}
			}
			e.Unlock()
		}
	}()

	for _, t := range e.triangles {
		t := t
		go func() {
			defer close(d)
			for {
				e.makeTrade(t)
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

func (e *Engine) makeTrade(triangle Triangle) {
	// Buy - Buy - Sell
	e.Lock()
	val := 1 / e.symbols[triangle.Intermediate+triangle.Base].Ask * 0.999 * 1 / e.symbols[triangle.Ticker+triangle.Intermediate].Ask * 0.999 * e.symbols[triangle.Ticker+triangle.Base].Bid * 0.999
	e.Unlock()
	if val > 1 {
		tt := time.Now()
		e.api.NewTestOrder()
		e.api.NewTestOrder()
		e.api.NewTestOrder()
		val1 := 1 / e.symbols[triangle.Intermediate+triangle.Base].Ask * 0.999 * 1 / e.symbols[triangle.Ticker+triangle.Intermediate].Ask * 0.999 * e.symbols[triangle.Ticker+triangle.Base].Bid * 0.999
		fmt.Println(triangle.Intermediate+triangle.Base, " -> ", triangle.Ticker+triangle.Intermediate, " -> ", triangle.Ticker+triangle.Base, " = ", val, " | API:", time.Since(tt), " | ", val1)
	}

	// Buy - Sell - Sell
	e.Lock()
	val = 1 / e.symbols[triangle.Ticker+triangle.Base].Ask * 0.999 * e.symbols[triangle.Ticker+triangle.Intermediate].Bid * 0.999 * e.symbols[triangle.Intermediate+triangle.Base].Bid * 0.999
	e.Unlock()
	if val > 1 {
		tt := time.Now()
		e.api.NewTestOrder()
		e.api.NewTestOrder()
		e.api.NewTestOrder()
		val1 := 1 / e.symbols[triangle.Ticker+triangle.Base].Ask * 0.999 * e.symbols[triangle.Ticker+triangle.Intermediate].Bid * 0.999 * e.symbols[triangle.Intermediate+triangle.Base].Bid * 0.999
		fmt.Println(triangle.Ticker+triangle.Base, " -> ", triangle.Ticker+triangle.Intermediate, " -> ", triangle.Intermediate+triangle.Base, " = ", val, " | API:", time.Since(tt), " | ", val1)
	}
}
