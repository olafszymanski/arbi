package binance

import "github.com/olafszymanski/arbi/pkg/utils"

type Updater struct {
}

func NewUpdater() *Updater {
	return &Updater{}
}

func (u *Updater) Update(symbols map[string]Symbol, ticker jsonOrderBookTicker) (*Symbol, error) {
	if _, ok := symbols[ticker.Symbol]; ok {
		b, err := utils.Stf(ticker.Bid)
		if err != nil {
			return nil, err
		}
		a, err := utils.Stf(ticker.Ask)
		if err != nil {
			return nil, err
		}
		return &Symbol{
			Symbol:    symbols[ticker.Symbol].Symbol,
			Base:      symbols[ticker.Symbol].Base,
			Quote:     symbols[ticker.Symbol].Quote,
			Precision: symbols[ticker.Symbol].Precision,
			Bid:       b,
			Ask:       a,
		}, nil
	}
	return nil, nil
}
