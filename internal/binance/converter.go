package binance

import (
	"github.com/olafszymanski/arbi/pkg/utils"
)

type APIConverter struct {
	validator *Validator
}

func NewAPIConverter(validator *Validator) *APIConverter {
	return &APIConverter{validator}
}

func (c *APIConverter) ToSymbols(symbols []jsonSymbol, orderBooks []jsonOrderBook, tradeFees []jsonTradeFee) ([]Symbol, error) {
	syms := make([]Symbol, 0)
	for _, s := range symbols {
		for _, b := range orderBooks {
			if i, ok := c.validator.Validate(s, b); ok {
				bid, err := utils.Stf(b.Bid)
				if err != nil {
					return nil, err
				}
				ask, err := utils.Stf(b.Ask)
				if err != nil {
					return nil, err
				}
				sp, err := utils.Stf(s.Filters[i].Precision)
				if err != nil {
					return nil, err
				}
				p := utils.GetPrecision(sp)
				for _, t := range tradeFees {
					if s.Symbol == t.Symbol {
						m, err := utils.Stf(t.MakerFee)
						if err != nil {
							return nil, err
						}
						t, err := utils.Stf(t.TakerFee)
						if err != nil {
							return nil, err
						}
						syms = append(syms, Symbol{
							Symbol:    s.Symbol,
							Base:      s.Base,
							Quote:     s.Quote,
							Bid:       bid,
							Ask:       ask,
							Precision: p,
							MakerFee:  m,
							TakerFee:  t,
						})
					}
				}
			}
		}
	}
	return syms, nil
}

func (c *APIConverter) ToAssets(assets []jsonAsset) ([]Asset, error) {
	a := make([]Asset, 0)
	for _, as := range assets {
		f, err := utils.Stf(as.Free)
		if err != nil {
			return nil, err
		}
		a = append(a, Asset{
			Symbol: as.Asset,
			Amount: f,
		})
	}
	return a, nil
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
