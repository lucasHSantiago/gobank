package util

import "slices"

const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
)

var currencies []string = []string{
	USD, EUR, CAD,
}

func IsSupportedCurrency(currency string) bool {
	return slices.Contains(currencies, currency)
}
