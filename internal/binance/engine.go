package binance

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/olafszymanski/arbi/config"
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

type Engine struct {
	*sync.RWMutex
	cfg       *config.Config
	api       *API
	websocket *Websocket
	triangles []Triangle
	symbols   map[string]Symbol
}

func NewEngine(cfg *config.Config, bases []string) *Engine {
	f := NewURLFactory()
	a := NewAPI(cfg, f)
	s := convertJSON(getJSON(a))
	t, syms := generate(s, bases)

	w, err := NewWebsocket(f)
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

	return &Engine{&sync.RWMutex{}, cfg, a, w, t, syms}
}

func getJSON(api *API) ([]jsonSymbol, []jsonOrderBook) {
	js, err := api.GetExchangeInfo()
	if err != nil {
		log.WithError(err).Panic()
	}
	job, err := api.GetOrderBook()
	if err != nil {
		log.WithError(err).Panic()
	}
	return js, job
}

func convertJSON(symbols []jsonSymbol, orderBooks []jsonOrderBook) []Symbol {
	v := NewValidator()
	c := NewConverter(v)
	s, err := c.Convert(symbols, orderBooks)
	if err != nil {
		log.WithError(err).Panic()
	}
	return s
}

func generate(symbols []Symbol, bases []string) ([]Triangle, map[string]Symbol) {
	g := NewGenerator()
	t, s, err := g.Generate(symbols, bases)
	if err != nil {
		log.WithError(err).Panic()
	}
	return t, s
}

func (e *Engine) Run() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})

	u := NewUpdater()

	go func() {
		defer close(done)
		defer e.websocket.Close()
		for {
			j, err := e.websocket.Read()
			if err != nil {
				log.WithError(err).Panic()
			}

			s, err := u.Update(e.symbols, *j)
			if err != nil {
				log.WithError(err).Panic()
			}
			if s != nil {
				e.Lock()
				e.symbols[j.Symbol] = *s
				e.Unlock()
			}
		}
	}()

	for _, t := range e.triangles {
		t := t
		go func() {
			defer close(done)
			for {
				e.makeTrade(t)
			}
		}()
	}

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			select {
			case <-done:
			case <-time.After(time.Microsecond):
			}
			return
		}
	}
}

func (e *Engine) makeTrade(triangle Triangle) {
	tt := time.Now()
	// Buy - Buy - Sell
	e.Lock()
	val := 1 / e.symbols[triangle.Intermediate+triangle.Base].Ask * 0.999 * 1 / e.symbols[triangle.Ticker+triangle.Intermediate].Ask * 0.999 * e.symbols[triangle.Ticker+triangle.Base].Bid * 0.999
	e.Unlock()
	if val > 1 {
		e.api.NewTestOrder()
		e.api.NewTestOrder()
		e.api.NewTestOrder()
		e.Lock()
		val1 := 1 / e.symbols[triangle.Intermediate+triangle.Base].Ask * 0.999 * 1 / e.symbols[triangle.Ticker+triangle.Intermediate].Ask * 0.999 * e.symbols[triangle.Ticker+triangle.Base].Bid * 0.999
		e.Unlock()
		fmt.Println(triangle.Ticker+triangle.Base, " -> ", triangle.Ticker+triangle.Intermediate, " -> ", triangle.Intermediate+triangle.Base, " = ", val, " | ", time.Since(tt), val1)
	}

	// Buy - Sell - Sell
	e.Lock()
	val = 1 / e.symbols[triangle.Ticker+triangle.Base].Ask * 0.999 * e.symbols[triangle.Ticker+triangle.Intermediate].Bid * 0.999 * e.symbols[triangle.Intermediate+triangle.Base].Bid * 0.999
	e.Unlock()
	if val > 1 {
		e.api.NewTestOrder()
		e.api.NewTestOrder()
		e.api.NewTestOrder()
		e.Lock()
		val1 := 1 / e.symbols[triangle.Ticker+triangle.Base].Ask * 0.999 * e.symbols[triangle.Ticker+triangle.Intermediate].Bid * 0.999 * e.symbols[triangle.Intermediate+triangle.Base].Bid * 0.999
		e.Unlock()
		fmt.Println(triangle.Ticker+triangle.Base, " -> ", triangle.Ticker+triangle.Intermediate, " -> ", triangle.Intermediate+triangle.Base, " = ", val, " | ", time.Since(tt), val1)
	}
}
