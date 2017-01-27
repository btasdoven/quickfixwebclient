package finance

import (
	"strconv"
	"time"
)

// Expiration is a date for options expiry.
type Expiration struct {
	Month string    `json:"m"`
	Day   string    `json:"d"`
	Year  string    `json:"y"`
	Date  time.Time `json:",omitempty"`
}

// NewExpiration builds an expiration object from month-day-year strings.
func NewExpiration(month, day, year string) *Expiration {
	e := &Expiration{
		Day:   day,
		Month: month,
		Year:  year,
	}
	e.setDate()
	return e

}

// ExpirationFromDate builds an expiration object from a proper date.
func ExpirationFromDate(date time.Time) *Expiration {

	return &Expiration{
		Day:   strconv.Itoa(date.Day()),
		Month: strconv.Itoa(int(date.Month())),
		Year:  strconv.Itoa(date.Year()),
		Date:  date,
	}

}

func (exp *Expiration) setDate() {
	dString := exp.Month + "/" + exp.Day + "/" + exp.Year
	exp.Date = parseDate(dString)
}
