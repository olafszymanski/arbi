package utils

import "math"

func Round(value float64, precision int) float64 {
	r := math.Pow10(precision)
	return math.Round(value*r) / r
}
