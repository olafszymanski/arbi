package binance

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(symbol jsonSymbol, book jsonOrderBook) bool {
	if symbol.Symbol == book.Symbol && (book.Bid != "0.00000000" && book.Ask != "0.00000000") {
		for _, p := range symbol.Permissions {
			if p == "SPOT" {
				return true
			}
		}
	}
	return false
}
