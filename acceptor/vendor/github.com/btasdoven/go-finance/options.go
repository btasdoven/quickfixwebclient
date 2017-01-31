package finance

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

const (
	optionsURL = "http://www.google.com/finance/option_chain?"
)

type fetchResult struct {
	Expiry      json.RawMessage     `json:"expiry"`
	Expirations []*Expiration       `json:"expirations"`
	Underlying  json.RawMessage     `json:"underlying_id"`
	Price       string              `json:"underlying_price"`
	Calls       []map[string]string `json:"calls,omitempty"`
	Puts        []map[string]string `json:"puts,omitempty"`
}

// NewOptionsChain creates a new OptionChain instance.
func NewOptionsChain(symbol string) (oc *OptionChain, err error) {

	oc = &OptionChain{Symbol: symbol}
	oc.Expirations, oc.UnderlyingPrice, err = chainExpirations(symbol)
	if err != nil {
		return nil, err
	}

	return oc, err
}

// chainExpirations fetches the expiration dates for an options chain.
func chainExpirations(symbol string) (expirations []*Expiration, price decimal.Decimal, err error) {

	e := ExpirationFromDate(time.Now())
	params := urlParams(symbol, e)

	url := buildURL(optionsURL, params)
	result, err := getOptionsData(url)
	if err != nil {
		return expirations, price, err
	}
	for _, e := range result.Expirations {
		e.setDate()
	}

	return result.Expirations, toDecimal(result.Price), err
}

// urlParams builds url parameters from an expiration.
func urlParams(symbol string, e *Expiration) map[string]string {
	return map[string]string{
		"q":      symbol,
		"expd":   e.Day,
		"expm":   e.Month,
		"expy":   e.Year,
		"output": "json",
	}
}

// getOptionsData fetches data from the endpoint and returns an intermediate result.
func getOptionsData(url string) (fr *fetchResult, err error) {

	b, err := request(url)
	if err != nil {
		return nil, fmt.Errorf("options fetch error:  (error was: %s)\n", err.Error())
	}

	err = json.Unmarshal(b, &fr)
	if err != nil {
		return nil, fmt.Errorf("options format error:  (error was: %s)\n", err.Error())
	}

	return fr, nil
}

// GetExpirations returns the specified option's expirations
func (chain *OptionChain) GetExpirations() []*Expiration {
	return chain.Expirations
}

// GetStrikes returns the specified option's strikes.
func (chain *OptionChain) GetStrikes() (strikes []decimal.Decimal, err error) {
	calls, err := chain.GetCallsForExpiration(chain.Expirations[0])
	if err != nil {
		return
	}

	for _, o := range calls {
		strikes = append(strikes, o.Strike)
	}

	return
}

// GetOptionsExpiringNext fetches calls and puts with the shortest expiration date.
func (chain *OptionChain) GetOptionsExpiringNext() (calls []*Option, puts []*Option, err error) {

	return chain.GetOptionsForExpiration(chain.Expirations[0])
}

// GetOptionsForExpiration fetches calls and puts for the given expiration date.
func (chain *OptionChain) GetOptionsForExpiration(e *Expiration) (calls []*Option, puts []*Option, err error) {

	if !chain.expirationExists(e) {
		err = fmt.Errorf("Expiration does not exist.")
		return
	}

	url := buildURL(optionsURL, urlParams(chain.Symbol, e))

	result, err := getOptionsData(url)
	if err != nil {
		return
	}

	return newContractSlice(result.Calls), newContractSlice(result.Puts), nil
}

// GetCallsForExpiration fetches calls for the given expiration date.
func (chain *OptionChain) GetCallsForExpiration(e *Expiration) (calls []*Option, err error) {
	calls, _, err = chain.GetOptionsForExpiration(e)
	return
}

// GetPutsForExpiration fetches puts for the given expiration date.
func (chain *OptionChain) GetPutsForExpiration(e *Expiration) (puts []*Option, err error) {
	_, puts, err = chain.GetOptionsForExpiration(e)
	return
}
