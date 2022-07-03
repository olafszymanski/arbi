package utils

import (
	"strconv"
)

func Stf(value string) (float64, error) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0.0, err
	}
	return f, nil
}
