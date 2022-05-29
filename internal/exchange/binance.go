package exchange

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/olafszymanski/arbi/cmd/config"
)

type pair struct {
	Crypto string
	Stable string
	Price  float64
}
type pairs map[string]pair

func (p pairs) HighestLowest(crypto string) (pair, pair) {
	hCrp, hStb, lCrp, lStb := "", "", "", ""
	var hPrc, lPrc float64
	for _, pr := range p {
		if crypto == pr.Crypto {
			if hPrc < pr.Price {
				hCrp = pr.Crypto
				hStb = pr.Stable
				hPrc = pr.Price
			}
			if lPrc == 0 || lPrc > pr.Price {
				lCrp = pr.Crypto
				lStb = pr.Stable
				lPrc = pr.Price
			}
		}
	}
	return pair{hCrp, hStb, hPrc}, pair{lCrp, lStb, lPrc}
}

type Binance struct {
	cfg  *config.Config
	lock sync.RWMutex
	prs  pairs
}

func NewBinance(cfg *config.Config, symbols map[string][]string) *Binance {
	res, err := http.Get(makeApiUrl(cfg, symbols))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	type Result struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	var prcs []Result
	json.Unmarshal(body, &prcs)

	prs := make(pairs)
	for key, syms := range symbols {
		for _, sym := range syms {
			s := key + sym
			for _, pr := range prcs {
				if pr.Symbol == s {
					prc, err := strconv.ParseFloat(pr.Price, 64)
					if err != nil {
						panic(err)
					}
					prs[s] = pair{key, sym, prc}
				}
			}
		}
	}

	return &Binance{
		cfg: cfg,
		prs: prs,
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
					log.Printf("Goroutine: %s panicked, error: %v", sym, r)
				}
			}()

			conn, _, err := websocket.DefaultDialer.Dial(makeWebsocketUrl(b.cfg, sym), nil)
			log.Printf("Connecting to %s websocket...", sym)
			if err != nil {
				panic(err)
			}
			defer conn.Close()

			for {
				_, data, err := conn.ReadMessage()
				if err != nil {
					log.Println("read:", err)
					return
				}

				var res Result
				if err := json.Unmarshal(data, &res); err != nil {
					panic(err)
				}

				prc, err := strconv.ParseFloat(res.Price, 64)
				if err != nil {
					panic(err)
				}
				b.lock.Lock()
				b.prs[sym] = pair{pr.Crypto, pr.Stable, prc}
				high, low := b.prs.HighestLowest(pr.Crypto)
				b.lock.Unlock()

				if val := b.calcProfitability(&high, &low); val > b.cfg.Binance.MinProfit {
					fmt.Println(high, low, val)
				}
			}
		}()
	}
}

func (b *Binance) calcProfitability(high, low *pair) float64 {
	toStb := high.Price - high.Price*b.cfg.Binance.Fee
	stbToStb := toStb - toStb*b.cfg.Binance.Fee
	stbToC := stbToStb - stbToStb*b.cfg.Binance.Fee
	return stbToC / low.Price
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
