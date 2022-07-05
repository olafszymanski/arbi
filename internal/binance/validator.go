package binance

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

// Validates returned JSON files, currently checks if API was able to correctly download bid and ask data along with permissions.
func (v *Validator) Validate(symbol jsonSymbol, orderBook jsonOrderBook) bool {
	if symbol.Symbol == orderBook.Symbol && (orderBook.Bid != "0.00000000" && orderBook.Ask != "0.00000000") {
		for _, p := range symbol.Permissions {
			if p == "SPOT" {
				return true
			}
		}
	}
	return false
}
