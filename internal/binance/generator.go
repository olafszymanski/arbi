package binance

import "strings"

type Triangle struct {
	Base         string
	Intermediate string
	Ticker       string
}

type Generator struct {
	converter *Converter
}

func NewGenerator(converter *Converter) *Generator {
	return &Generator{converter}
}

// Generates triangular combinations along with unique crypto pairs - needed later for websocket data flow.
func (v *Generator) Generate(symbols []Symbol, bases []string) ([]Triangle, map[string]Symbol, error) {
	t := make([]Triangle, 0)
	s := make(map[string]Symbol)
	for _, b := range bases {
		for _, s1 := range symbols {
			if s1.Quote == b {
				for _, s2 := range symbols {
					if s1.Base == s2.Quote {
						for _, s3 := range symbols {
							if s2.Base == s3.Base && s3.Quote == s1.Quote {
								t = append(t, Triangle{strings.ToLower(s1.Quote), strings.ToLower(s1.Base), strings.ToLower(s2.Base)})
								if _, ok := s[strings.ToLower(s1.Symbol)]; !ok {
									s[strings.ToLower(s1.Symbol)] = s1
								}
								if _, ok := s[strings.ToLower(s2.Symbol)]; !ok {
									s[strings.ToLower(s2.Symbol)] = s2
								}
								if _, ok := s[strings.ToLower(s3.Symbol)]; !ok {
									s[strings.ToLower(s3.Symbol)] = s3
								}
							}
						}
					}
				}
			}
		}
	}
	return t, s, nil
}
