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
	listenKey          string
	orderBookWebsocket *OrderBookWebsocket
	walletWebsocket    *WalletWebsocket
	triangles          []Triangle
	symbols            map[string]Symbol
	wallet             Wallet
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

	js, err := a.GetExchangeInfo()
	if err != nil {
		log.WithError(err).Panic()
	}
	log.Info("Successfully retrieved exchange info.")
	job, err := a.GetOrderBook()
	if err != nil {
		log.WithError(err).Panic()
	}
	log.Info("Successfully retrieved order books.")
	// TODO: Fetch trading fees

	s, err := c.ToSymbols(js, job)
	if err != nil {
		log.WithError(err).Panic()
	}
	log.Info("Successfully converted JSON symbols data to symbols.")

	t, syms, err := g.Generate(s, bases)
	if err != nil {
		log.WithError(err).Panic()
	}
	if len(t) > 1000 {
		t = t[:1000]
	}
	log.Info("Successfully generated triangles and symbol map.")

	ja, err := a.GetUserAssets()
	if err != nil {
		log.WithError(err).Panic()
	}
	log.Info("Successfully retrieved user assets.")

	w, err := c.ToWallet(ja)
	if err != nil {
		log.WithError(err).Panic()
	}
	log.Info("Successfully converted JSON user assets data to wallet.")

	obw, err := NewOrderBookWebsocket(f)
	if err != nil {
		log.WithError(err).Panic()
	}
	log.Info("Successfully initialized order book websocket.")

	k, err := a.GetListenKey()
	if err != nil {
		log.WithError(err).Panic()
	}
	log.Info("Successfully retrieved the listen key.")

	ww, err := NewWalletWebsocket(f, k)
	if err != nil {
		log.WithError(err).Panic()
	}
	log.Info("Successfully initialized wallet websocket.")

	// log.Info("Testing latency to api.binance.com...")
	// for i := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15} {
	// 	tt := time.Now()
	// 	a.NewTestOrder()
	// 	a.NewTestOrder()
	// 	a.NewTestOrder()
	// 	fmt.Println(fmt.Sprintf("T%v:", i), time.Since(tt))
	// }

	log.Info("Successfully initialized the engine.")

	return &Engine{&sync.RWMutex{}, cfg, a, k, obw, ww, t, syms, w, 0, 0, false}
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
			e.Lock()
			if s, ok := e.symbols[o.Symbol]; ok {
				e.symbols[o.Symbol] = Symbol{
					Symbol:    s.Symbol,
					Base:      s.Base,
					Quote:     s.Quote,
					Bid:       b,
					Ask:       a,
					Precision: s.Precision,
				}
				for _, t := range e.triangles {
					if p := e.profitability(t); p > 1.001 {
						e.makeTrade(t, p)
					}
					if p := e.reverseProfitability(t); p > 1.001 {
						e.makeReverseTrade(t, p)
					}
				}
			}
			e.Unlock()
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
				e.Lock()
				e.wallet[bal.Asset] = a
				e.Unlock()
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

	// go func() {
	// 	defer close(d)
	// 	for {
	// 		e.Lock()
	// 		o := e.orders
	// 		// do := e.dailyOrders
	// 		e.Unlock()

	// 		if o > 48 {
	// 			e.Lock()
	// 			e.blocked = true
	// 			e.orders = 0
	// 			e.blocked = false
	// 			e.Unlock()
	// 		}
	// 		// TODO: Implement daily orders
	// 		// if do > 159990 {
	// 		// 	e.Lock()
	// 		// }
	// 	}
	// }()

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
	return 1 / e.symbols[triangle.FirstPair()].Ask * 0.999 * 1 / e.symbols[triangle.SecondPair()].Ask * 0.999 * e.symbols[triangle.ThirdPair()].Bid * 0.999
}

func (e *Engine) reverseProfitability(triangle Triangle) float64 {
	return 1 / e.symbols[triangle.ThirdPair()].Ask * 0.999 * e.symbols[triangle.SecondPair()].Bid * 0.999 * e.symbols[triangle.FirstPair()].Bid * 0.999
}

func (e *Engine) makeTrade(triangle Triangle, profitability float64) {
	// Buy - Buy - Sell
	t := time.Now()
	if err := e.api.NewOrder(triangle.FirstPair(), "BUY", e.wallet[e.symbols[triangle.FirstPair()].Base], e.symbols[triangle.FirstPair()].Precision); err != nil {
		log.WithError(err).Error("Error while placing new order")
		return
	}
	if err := e.api.NewOrder(triangle.SecondPair(), "BUY", e.wallet[e.symbols[triangle.SecondPair()].Base], e.symbols[triangle.SecondPair()].Precision); err != nil {
		log.WithError(err).Error("Error while placing new order")
		return
	}
	if err := e.api.NewOrder(triangle.ThirdPair(), "SELL", e.wallet[e.symbols[triangle.ThirdPair()].Quote], e.symbols[triangle.SecondPair()].Precision); err != nil {
		log.WithError(err).Error("Error while placing new order")
		return
	}
	s := time.Since(t)

	// e.api.NewTestOrder()
	// e.api.NewTestOrder()
	// e.api.NewTestOrder()

	// // TODO: Remove later
	// e.Lock()
	// p := e.profitability(triangle)
	// e.Unlock()

	fmt.Println("BUY", triangle.FirstPair(), " ->  BUY", triangle.SecondPair(), " ->  SELL", triangle.ThirdPair(), " = ", profitability, " | API:", s)
}

func (e *Engine) makeReverseTrade(triangle Triangle, profitability float64) {
	// Buy - Sell - Sell
	t := time.Now()
	if err := e.api.NewOrder(triangle.ThirdPair(), "BUY", e.wallet[e.symbols[triangle.ThirdPair()].Base], e.symbols[triangle.ThirdPair()].Precision); err != nil {
		log.WithError(err).Error("Error while placing new order")
		return
	}
	if err := e.api.NewOrder(triangle.SecondPair(), "SELL", e.wallet[e.symbols[triangle.SecondPair()].Quote], e.symbols[triangle.SecondPair()].Precision); err != nil {
		log.WithError(err).Error("Error while placing new order")
		return
	}
	if err := e.api.NewOrder(triangle.FirstPair(), "SELL", e.wallet[e.symbols[triangle.FirstPair()].Quote], e.symbols[triangle.FirstPair()].Precision); err != nil {
		log.WithError(err).Error("Error while placing new order")
		return
	}
	// e.api.NewTestOrder()
	// e.api.NewTestOrder()
	// e.api.NewTestOrder()
	s := time.Since(t)

	// // TODO: Remove later
	// e.Lock()
	// p := e.reverseProfitability(triangle)
	// e.Unlock()

	fmt.Println("BUY", triangle.ThirdPair(), " ->  SELL", triangle.SecondPair(), " ->  SELL", triangle.FirstPair(), " = ", profitability, " | API:", s)
}
