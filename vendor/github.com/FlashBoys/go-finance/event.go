package finance

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	// Dividend constant.
	Dividend = "DIVIDEND"
	// Split constant.
	Split = "SPLIT"
)

// Event contains one historical event (either a split or a dividend).
type Event struct {
	Symbol      string
	EventType   string
	Date        time.Time
	DividendAmt decimal.Decimal
	SplitRatio  string
}

// newEvent creates a new instance of an event struct.
func newEvent(symbol string, row []string) *Event {

	e := &Event{
		Symbol:    symbol,
		EventType: row[0],
		Date:      parseDashedDate(parseMalformedDate(row[1])),
	}

	if e.EventType == Dividend {
		e.DividendAmt = toDecimal(row[2])
	} else {
		e.SplitRatio = row[2]
	}

	return e
}
