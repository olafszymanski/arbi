package binance

type Triangle struct {
	Base         string
	Intermediate string
	Ticker       string
}

type Generator struct {
}

func NewGenerator() *Generator {
	return &Generator{}
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
								t = append(t, Triangle{s1.Quote, s1.Base, s2.Base})
								if _, ok := s[s1.Symbol]; !ok {
									s[s1.Symbol] = s1
								}
								if _, ok := s[s2.Symbol]; !ok {
									s[s2.Symbol] = s2
								}
								if _, ok := s[s3.Symbol]; !ok {
									s[s3.Symbol] = s3
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
