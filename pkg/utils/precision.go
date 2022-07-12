package utils

import (
	"strconv"
	"strings"
)

func GetPrecision(precision float64) int {
	s := strconv.FormatFloat(precision, 'f', -1, 64)
	i := strings.IndexByte(s, '.')
	if i > -1 {
		return len(s) - i - 1
	}
	return 0
}
