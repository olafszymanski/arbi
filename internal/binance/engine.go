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
	jsonSyms, jsonOB := apiInfo(a)

	v := NewValidator()
	c := NewConverter(v)
	s, err := c.Convert(jsonSyms, jsonOB)
	if err != nil {
		log.WithError(err).Panic()
	}

	g := NewGenerator(c)
	t, syms, err := g.Generate(s, bases)
	if err != nil {
		log.WithError(err).Panic()
	}

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

func apiInfo(api *API) ([]jsonSymbol, []jsonOrderBook) {
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

	go func() {
		defer close(done)
		for {
			e.Profitability()
		}
	}()

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

func (e *Engine) Profitability() {
	for _, t := range e.triangles {
		tt := time.Now()
		// Buy - Buy - Sell
		e.Lock()
		val := 1 / e.symbols[t.Intermediate+t.Base].Ask * 0.999 * 1 / e.symbols[t.Ticker+t.Intermediate].Ask * 0.999 * e.symbols[t.Ticker+t.Base].Bid * 0.999
		e.Unlock()
		if val > 1 {
			e.api.NewTestOrder()
			e.Lock()
			val1 := 1 / e.symbols[t.Intermediate+t.Base].Ask * 0.999 * 1 / e.symbols[t.Ticker+t.Intermediate].Ask * 0.999 * e.symbols[t.Ticker+t.Base].Bid * 0.999
			e.Unlock()
			e.api.NewTestOrder()
			e.Lock()
			val2 := 1 / e.symbols[t.Intermediate+t.Base].Ask * 0.999 * 1 / e.symbols[t.Ticker+t.Intermediate].Ask * 0.999 * e.symbols[t.Ticker+t.Base].Bid * 0.999
			e.Unlock()
			e.api.NewTestOrder()
			e.Lock()
			val3 := 1 / e.symbols[t.Intermediate+t.Base].Ask * 0.999 * 1 / e.symbols[t.Ticker+t.Intermediate].Ask * 0.999 * e.symbols[t.Ticker+t.Base].Bid * 0.999
			e.Unlock()
			fmt.Println(t.Ticker+t.Base, " -> ", t.Ticker+t.Intermediate, " -> ", t.Intermediate+t.Base, " = ", val, " | ", time.Since(tt), val1, val2, val3)
		}

		// Buy - Sell - Sell
		e.Lock()
		val = 1 / e.symbols[t.Ticker+t.Base].Ask * 0.999 * e.symbols[t.Ticker+t.Intermediate].Bid * 0.999 * e.symbols[t.Intermediate+t.Base].Bid * 0.999
		e.Unlock()
		if val > 1 {
			e.api.NewTestOrder()
			e.Lock()
			val1 := 1 / e.symbols[t.Ticker+t.Base].Ask * 0.999 * e.symbols[t.Ticker+t.Intermediate].Bid * 0.999 * e.symbols[t.Intermediate+t.Base].Bid * 0.999
			e.Unlock()
			e.api.NewTestOrder()
			e.Lock()
			val2 := 1 / e.symbols[t.Ticker+t.Base].Ask * 0.999 * e.symbols[t.Ticker+t.Intermediate].Bid * 0.999 * e.symbols[t.Intermediate+t.Base].Bid * 0.999
			e.Unlock()
			e.api.NewTestOrder()
			e.Lock()
			val3 := 1 / e.symbols[t.Ticker+t.Base].Ask * 0.999 * e.symbols[t.Ticker+t.Intermediate].Bid * 0.999 * e.symbols[t.Intermediate+t.Base].Bid * 0.999
			e.Unlock()
			fmt.Println(t.Ticker+t.Base, " -> ", t.Ticker+t.Intermediate, " -> ", t.Intermediate+t.Base, " = ", val, " | ", time.Since(tt), val1, val2, val3)
		}
	}
}
