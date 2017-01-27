package finance

import (
	"time"

	"github.com/shopspring/decimal"
)

// Bar represents a single bar(candle) in time-series of quotes.
type Bar struct {
	Symbol   string
	Date     time.Time
	Open     decimal.Decimal
	High     decimal.Decimal
	Low      decimal.Decimal
	Close    decimal.Decimal
	Volume   int
	AdjClose decimal.Decimal
}

// newBar creates a new instance of a bar struct.
func newBar(symbol string, row []string) *Bar {
	return &Bar{
		Symbol:   symbol,
		Date:     parseDashedDate(row[0]),
		Open:     toDecimal(row[1]),
		High:     toDecimal(row[2]),
		Low:      toDecimal(row[3]),
		Close:    toDecimal(row[4]),
		Volume:   toInt(row[5]),
		AdjClose: toDecimal(row[6]),
	}
}
