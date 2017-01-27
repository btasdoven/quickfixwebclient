package finance

import (
	"time"

	"github.com/shopspring/decimal"
)

// FXFields are the requested fx pair fields.
var fXFields = []string{
	"s",  // Symbol
	"n",  // Name
	"d1", // LastTradeDate
	"t1", // LastTradeTime
	"l1", // LastTradePriceOnly
	"c1", // Change
	"p2", // ChangeinPercent
	"g",  // DaysLow
	"h",  // DaysHigh
	"j",  // YearLow
	"k",  // YearHigh
}

// FXPairQuote represents the quote of a currency pair.
type FXPairQuote struct {
	Symbol           string
	PairName         string
	LastTime         time.Time
	LastRate         decimal.Decimal
	ChangeNominal    decimal.Decimal
	ChangePercent    decimal.Decimal
	DayLow           decimal.Decimal
	DayHigh          decimal.Decimal
	FiftyTwoWeekLow  decimal.Decimal
	FiftyTwoWeekHigh decimal.Decimal
}

// newFXPairQuote creates a new instance of quote of a currency pair.
func newFXPairQuote(row []string) *FXPairQuote {

	fields := make(map[string]string, 0)
	for idx, v := range fXFields {
		fields[v] = row[idx]
	}

	return &FXPairQuote{
		Symbol:           fields["s"],
		PairName:         fields["n"],
		LastTime:         parseDateAndTime(fields["d1"], fields["t1"]),
		LastRate:         toDecimal(fields["l1"]),
		ChangeNominal:    toDecimal(fields["c1"]),
		ChangePercent:    toDecimal(fields["p2"]),
		DayLow:           toDecimal(fields["g"]),
		DayHigh:          toDecimal(fields["h"]),
		FiftyTwoWeekLow:  toDecimal(fields["j"]),
		FiftyTwoWeekHigh: toDecimal(fields["k"]),
	}
}
