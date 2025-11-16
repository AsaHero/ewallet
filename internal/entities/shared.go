package entities

import "math"

type Language string

const (
	RU Language = "ru"
)

func (l Language) String() string {
	return string(l)
}

type Currency string

const (
	NULL Currency = ""
	UZS  Currency = "UZS"
	RUB  Currency = "RUB"
	USD  Currency = "USD"
	EUR  Currency = "EUR"
	JPY  Currency = "JPY"
	KWD  Currency = "KWD"
)

var currencyScale = map[Currency]int{
	RUB: 2,
	USD: 2,
	EUR: 2,
	JPY: 0,
	KWD: 3,
	UZS: 2,
}

func (c Currency) String() string {
	return string(c)
}

func (c Currency) Scale() int {
	if s, ok := currencyScale[c]; ok {
		return s
	}
	return 2
}

func MinorFromMajor(major float64, scale int) int64 {
	multiplier := math.Pow10(scale)
	return int64(math.Round(major * multiplier))
}

func MajorFromMinor(minor int64, scale int) float64 {
	multiplier := math.Pow10(scale)
	return float64(minor) / multiplier
}
