package binance

import "github.com/olafszymanski/arbi/pkg/utils"

type Converter struct {
	validator *Validator
}

func NewConverter(validator *Validator) *Converter {
	return &Converter{validator}
}

func (c *Converter) Convert(symbols []jsonSymbol, orderBooks []jsonOrderBook) ([]Symbol, error) {
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
