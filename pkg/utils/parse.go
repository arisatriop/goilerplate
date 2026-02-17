package utils

import "github.com/shopspring/decimal"

func ParseDecimal(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}
