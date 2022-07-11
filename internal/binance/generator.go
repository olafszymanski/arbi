package binance

import "sync"

type Triangle struct {
	Base         string
	Intermediate string
	Ticker       string
}

func (t Triangle) FirstPair() string {
	return t.Intermediate + t.Base
}

func (t Triangle) SecondPair() string {
	return t.Ticker + t.Intermediate
}

func (t Triangle) ThirdPair() string {
	return t.Ticker + t.Base
}

type Generator struct {
}

func NewGenerator() *Generator {
	return &Generator{}
}

// Generates triangular combinations along with unique crypto pairs - needed later for websocket data flow.
func (v *Generator) Generate(symbols []Symbol, assets []Asset, bases []string) ([]Triangle, sync.Map, error) {
	t := make([]Triangle, 0)
	var d sync.Map
	for _, b := range bases {
		for _, s1 := range symbols {
			if s1.Quote == b {
				for _, s2 := range symbols {
					if s1.Base == s2.Quote {
						for _, s3 := range symbols {
							if s2.Base == s3.Base && s3.Quote == s1.Quote {
								t = append(t, Triangle{s1.Quote, s1.Base, s2.Base})
								if _, ok := d.Load(s1.Symbol); !ok {
									d.Store(s1.Symbol, s1)
								}
								if _, ok := d.Load(s2.Symbol); !ok {
									d.Store(s2.Symbol, s2)
								}
								if _, ok := d.Load(s3.Symbol); !ok {
									d.Store(s3.Symbol, s3)
								}
							}
						}
					}
				}
			}
		}
	}
	for _, a := range assets {
		d.Store(a.Symbol, a.Amount)
	}
	return t, d, nil
}
