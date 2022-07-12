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
	Bid       float64
	Ask       float64
	Precision int
	MakerFee  float64
	TakerFee  float64
}

type Asset struct {
	Symbol string
	Amount float64
}

type Engine struct {
	*sync.RWMutex
	cfg                *config.Config
	api                *API
	listenKey          string
	orderBookWebsocket *OrderBookWebsocket
	walletWebsocket    *WalletWebsocket
	triangles          []Triangle
	data               *DataManager
	orders             uint8
	dailyOrders        uint32
	blocked            bool
}

func NewEngine(cfg *config.Config, bases []string) *Engine {
	log.Info("Initializing the engine...")

	f := NewURLFactory()
	a := NewAPI(cfg, f)
	v := NewValidator()
	c := NewAPIConverter(v)
	g := NewGenerator()

	js, job, jt, ja := getData(a)
	s, as := convert(c, js, job, jt, ja)
	t, d := generate(g, s, as, bases)
	k, obw, ww := connectWebsockets(f, a)

	// testLatency(a)

	log.Info("Successfully initialized the engine.")

	return &Engine{&sync.RWMutex{}, cfg, a, k, obw, ww, t, NewDataManager(d), 0, 0, false}
}

func (e *Engine) Run() {
	log.Info("Starting the main engine function...")

	d := make(chan struct{})
	i := make(chan os.Signal, 1)
	signal.Notify(i, os.Interrupt)

	c := NewWebsocketConverter()

	go func() {
		defer close(d)
		defer e.orderBookWebsocket.Close()

		log.Info("Starting order book websocket...")

		for {
			o, err := e.orderBookWebsocket.Read()
			if err != nil {
				log.WithError(err).Panic()
			}

			b, a, err := c.ToPrices(o)
			if err != nil {
				log.WithError(err).Panic()
			}
			if s, ok := e.data.SymbolExists(o.Symbol); ok {
				e.data.StoreSymbol(o.Symbol, Symbol{
					Symbol:    s.Symbol,
					Base:      s.Base,
					Quote:     s.Quote,
					Bid:       b,
					Ask:       a,
					Precision: s.Precision,
					MakerFee:  s.MakerFee,
					TakerFee:  s.TakerFee,
				})
				for _, t := range e.triangles {
					if p := e.profitability(t); p > 1.001 {
						e.makeTrade(t, p)
						return
					}
					// if p := e.reverseProfitability(t); p > 1.001 {
					// 	e.makeReverseTrade(t, p)
					// 	return
					// }
				}
			}
		}
	}()

	go func() {
		defer close(d)
		defer e.walletWebsocket.Close()

		log.Info("Starting wallet websocket...")

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
				e.data.StoreFloat(bal.Asset, a)
			}
		}
	}()

	go func() {
		defer close(d)
		for {
			time.Sleep(time.Minute * 30)
			e.api.KeepAliveListenKey(e.listenKey)
			log.Info("Successfully sent keep alive listen key request...")
		}
	}()

	// log.Info("Starting triangle goroutines...")
	// for _, t := range e.triangles {
	// 	t := t
	// 	go func() {
	// 		defer close(d)
	// 		for {
	// 			e.Lock()
	// 			p := e.profitability(t)
	// 			rp := e.reverseProfitability(t)
	// 			e.Unlock()

	// 			if p > 0.999 || rp > 0.999 {
	// 				fmt.Println(t, p, rp)
	// 			}

	// 			if p > 1.001 {
	// 				e.makeTrade(t, p)
	// 			}
	// 			if rp > 1.001 {
	// 				e.makeReverseTrade(t, p)
	// 			}
	// 		}
	// 	}()
	// }

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
	f := e.data.LoadSymbol(triangle.FirstPair())
	s := e.data.LoadSymbol(triangle.SecondPair())
	t := e.data.LoadSymbol(triangle.ThirdPair())
	return 1 / f.Ask * (1 - f.TakerFee) * 1 / s.Ask * (1 - s.TakerFee) * t.Bid * (1 - t.TakerFee)
}

// func (e *Engine) reverseProfitability(triangle Triangle) float64 {
// 	return 1 / e.symbols[triangle.ThirdPair()].Ask * 0.999 * e.symbols[triangle.SecondPair()].Bid * 0.999 * e.symbols[triangle.FirstPair()].Bid * 0.999
// }

func (e *Engine) makeTrade(triangle Triangle, profitability float64) {
	// Buy - Buy - Sell
	ti := time.Now()

	f := e.data.LoadSymbol(triangle.FirstPair())
	fq := e.data.LoadFloat(f.Quote)
	var frq float64
	if fq > 1 {
		frq = utils.Round(fq, f.Precision)
	} else {
		frq = utils.Round(fq, 8)
	}
	fo, err := e.api.NewOrder(triangle.FirstPair(), "BUY", frq)
	fmt.Println(f, fq, frq)
	fmt.Println(fo)
	if err != nil {
		log.Error(err)
		return
	}
	q, err := utils.Stf(fo.Quantity)
	if err != nil {
		log.Error(err)
		return
	}
	e.data.StoreFloat(f.Base, q*(1-f.TakerFee))

	s := e.data.LoadSymbol(triangle.SecondPair())
	sq := e.data.LoadFloat(s.Quote)
	var srq float64
	if sq > 1 {
		srq = utils.Round(sq, s.Precision)
	} else {
		srq = utils.Round(sq, 8)
	}
	so, err := e.api.NewOrder(triangle.SecondPair(), "BUY", srq)
	fmt.Println(s, sq, srq)
	fmt.Println(so)
	if err != nil {
		log.Error(err)
		return
	}
	q, err = utils.Stf(so.Quantity)
	if err != nil {
		log.Error(err)
		return
	}
	e.data.StoreFloat(s.Base, q*(1-s.TakerFee))

	t := e.data.LoadSymbol(triangle.ThirdPair())
	tq := e.data.LoadFloat(t.Base)
	var trq float64
	if tq > 1 {
		trq = utils.Round(tq, t.Precision)
	} else {
		trq = utils.Round(tq, 8)
	}
	to, err := e.api.NewOrder(triangle.ThirdPair(), "SELL", trq)
	fmt.Println(t, tq, trq)
	fmt.Println(to)
	if err != nil {
		log.Error(err)
		return
	}
	q, err = utils.Stf(to.Quantity)
	if err != nil {
		log.Error(err)
		return
	}
	e.data.StoreFloat(t.Quote, q*(1-t.TakerFee))

	si := time.Since(ti)

	fmt.Println("BUY", triangle.FirstPair(), " ->  BUY", triangle.SecondPair(), " ->  SELL", triangle.ThirdPair(), " = ", profitability, " | API:", si)
}

// func (e *Engine) makeReverseTrade(triangle Triangle, profitability float64) {
// 	// Buy - Sell - Sell
// 	t := time.Now()
// 	if err := e.api.NewOrder(triangle.ThirdPair(), "BUY", e.wallet[e.symbols[triangle.ThirdPair()].Quote]/e.symbols[triangle.ThirdPair()].Ask, e.symbols[triangle.ThirdPair()].Precision); err != nil {
// 		log.WithError(err).Error("Error while placing new order")
// 		return
// 	}
// 	if err := e.api.NewOrder(triangle.SecondPair(), "SELL", e.wallet[e.symbols[triangle.SecondPair()].Base]/e.symbols[triangle.SecondPair()].Bid, e.symbols[triangle.SecondPair()].Precision); err != nil {
// 		log.WithError(err).Error("Error while placing new order")
// 		return
// 	}
// 	if err := e.api.NewOrder(triangle.FirstPair(), "SELL", e.wallet[e.symbols[triangle.FirstPair()].Base]/e.symbols[triangle.FirstPair()].Bid, e.symbols[triangle.FirstPair()].Precision); err != nil {
// 		log.WithError(err).Error("Error while placing new order")
// 		return
// 	}
// 	// e.api.NewTestOrder()
// 	// e.api.NewTestOrder()
// 	// e.api.NewTestOrder()
// 	s := time.Since(t)

// 	// // TODO: Remove later
// 	// e.Lock()
// 	// p := e.reverseProfitability(triangle)
// 	// e.Unlock()

// 	fmt.Println("BUY", triangle.ThirdPair(), " ->  SELL", triangle.SecondPair(), " ->  SELL", triangle.FirstPair(), " = ", profitability, " | API:", s)
// }

func getData(api *API) ([]jsonSymbol, []jsonOrderBook, []jsonTradeFee, []jsonAsset) {
	s, err := api.GetExchangeInfo()
	if err != nil {
		log.WithError(err).Panic()
	}
	log.Info("Successfully retrieved exchange info.")
	o, err := api.GetOrderBook()
	if err != nil {
		log.WithError(err).Panic()
	}
	log.Info("Successfully retrieved order books.")
	t, err := api.GetTradeFees()
	if err != nil {
		log.WithError(err).Panic()
	}
	log.Info("Successfully retrieved trade fees.")
	a, err := api.GetUserAssets()
	if err != nil {
		log.WithError(err).Panic()
	}
	log.Info("Successfully retrieved user assets.")
	return s, o, t, a
}

func convert(converter *APIConverter, symbols []jsonSymbol, orderBooks []jsonOrderBook, tradeFees []jsonTradeFee, assets []jsonAsset) ([]Symbol, []Asset) {
	s, err := converter.ToSymbols(symbols, orderBooks, tradeFees)
	if err != nil {
		log.WithError(err).Panic()
	}
	log.Info("Successfully converted JSON symbols data to symbols.")
	a, err := converter.ToAssets(assets)
	if err != nil {
		log.WithError(err).Panic()
	}
	log.Info("Successfully converted JSON assets data to assets.")
	return s, a
}

func generate(generator *Generator, symbols []Symbol, assets []Asset, bases []string) ([]Triangle, sync.Map) {
	t, d, err := generator.Generate(symbols, assets, bases)
	if err != nil {
		log.WithError(err).Panic()
	}
	if len(t) > 1000 {
		t = t[:1000]
	}
	log.Info("Successfully generated triangles and symbol map.")
	return t, d
}

func connectWebsockets(factory *URLFactory, api *API) (string, *OrderBookWebsocket, *WalletWebsocket) {
	k, err := api.GetListenKey()
	if err != nil {
		log.WithError(err).Panic()
	}
	log.Info("Successfully retrieved the listen key.")

	obw, err := NewOrderBookWebsocket(factory)
	if err != nil {
		log.WithError(err).Panic()
	}
	log.Info("Successfully initialized order book websocket.")
	ww, err := NewWalletWebsocket(factory, k)
	if err != nil {
		log.WithError(err).Panic()
	}
	log.Info("Successfully initialized wallet websocket.")
	return k, obw, ww
}

func testLatency(api *API) {
	log.Info("Testing latency to api.binance.com...")
	l := 0
	for range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15} {
		tt := time.Now()
		api.NewTestOrder()
		api.NewTestOrder()
		api.NewTestOrder()
		l += int(time.Since(tt).Milliseconds())
	}
	log.Info(fmt.Sprintf("Average latency: %vms", l/15))
}
