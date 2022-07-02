package broker

import (
	"strings"
)

type Pair struct {
	Crypto string
	Stable string
	Price  float64
}

type Pairs map[string]Pair

type Price struct {
	Symbol string
	Price  float64
}

type Balance struct {
	Asset  string
	Amount float64
}

type Account map[string]Balance

func (p Pairs) HighestLowest(crypto string) (Pair, Pair) {
	hCrp, hStb, lCrp, lStb := "", "", "", ""
	var hPrc, lPrc float64
	for _, pr := range p {
		if crypto == pr.Crypto {
			if hPrc < pr.Price && strings.ToLower(pr.Stable) == "usdt" {
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
	return Pair{hCrp, hStb, hPrc}, Pair{lCrp, lStb, lPrc}
}

type IBroker interface {
	Subscribe()
}

func Profitability(high, low *Pair, fee, conversion float64) float64 {
	toStb := high.Price - high.Price*fee
	stbToStb := (toStb - toStb*fee) * conversion
	stbToC := stbToStb - stbToStb*fee
	return stbToC / low.Price
}
