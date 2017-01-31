package finance

import (
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// toInt converts a string to an int.
func toInt(value string) int {

	i, _ := strconv.Atoi(value)
	// if err != nil {
	// 	fmt.Println("error converting ", value, " to int: ", err)
	// }

	return i
}

// toDecimal converts a string to a decimal value.
func toDecimal(value string) decimal.Decimal {

	value = strings.Replace(value, "%", "", -1)
	dec, err := decimal.NewFromString(value)
	if err != nil {
		// fmt.Println("error converting ", value, " to decimal value: ", err)
		return decimal.NewFromFloat(0.0)
	}

	return dec
}

// parseDate converts a string to a proper date.
func parseDate(dString string) time.Time {

	d, err := time.Parse("1/2/2006", dString)
	if err != nil {
		// fmt.Println("error converting ", dString, " to date: ", err)
		return time.Time{}
	}
	return d
}

// parseDateAndTime converts a string to a proper date with a time.
func parseDateAndTime(dString string, tString string) time.Time {

	d, err := time.Parse("1/2/2006", dString)
	if err != nil {
		// fmt.Println("error converting ", dString, " to date: ", err)
		return time.Time{}
	}
	t, err := time.Parse("3:04pm", tString)
	if err != nil {
		// fmt.Println("error converting ", tString, " to time: ", err)
		return time.Time{}
	}
	loc, _ := time.LoadLocation("America/New_York")

	return time.Date(d.Year(), d.Month(), d.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc)
}

// parseDashedDate converts a string to a proper date and sets time to market close.
func parseDashedDate(dString string) time.Time {

	date, err := time.Parse("2006-01-02", dString)
	if err != nil {
		// fmt.Println("error converting ", dString, " to date: ", err)
		return time.Time{}
	}
	return date.Add(time.Hour * 16)
}

func parseMalformedDate(s string) string {

	chars := strings.Split(s, "")
	chars = chars[1:]
	chars = insert(chars, 4, "-")
	chars = insert(chars, 7, "-")
	return strings.Join(chars[:], "")
}

func insert(s []string, i int, x string) []string {

	s = append(s, "")
	copy(s[i+1:], s[i:])
	s[i] = x

	return s
}
