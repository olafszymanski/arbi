package binance

import "github.com/olafszymanski/arbi/pkg/utils"

type APIConverter struct {
	validator *Validator
}

func NewAPIConverter(validator *Validator) *APIConverter {
	return &APIConverter{validator}
}

func (c *APIConverter) ToSymbols(symbols []jsonSymbol, orderBooks []jsonOrderBook) ([]Symbol, error) {
	syms := make([]Symbol, 0)
	for _, s := range symbols {
		for _, b := range orderBooks {
			if c.validator.Validate(s, b) {
				bid, err := utils.Stf(b.Bid)
				if err != nil {
					return nil, err
				}
				ask, err := utils.Stf(b.Ask)
				if err != nil {
					return nil, err
				}
				syms = append(syms, Symbol{
					Symbol:    s.Symbol,
					Base:      s.Base,
					Quote:     s.Quote,
					Precision: s.Precision,
					Bid:       bid,
					Ask:       ask,
				})
			}
		}
	}
	return syms, nil
}

func (c *APIConverter) ToWallet(assets []jsonAsset) (Wallet, error) {
	w := make(Wallet)
	for _, a := range assets {
		f, err := utils.Stf(a.Free)
		if err != nil {
			return nil, err
		}
		w[a.Asset] = f
	}
	return w, nil
}

type WebsocketConverter struct {
}

func NewWebsocketConverter() *WebsocketConverter {
	return &WebsocketConverter{}
}

func (c *WebsocketConverter) ToPrices(ticker *jsonOrderBookTicker) (float64, float64, error) {
	b, err := utils.Stf(ticker.Bid)
	if err != nil {
		return 0, 0, err
	}
	a, err := utils.Stf(ticker.Ask)
	if err != nil {
		return 0, 0, err
	}
	return b, a, nil
}
