package finance

import "github.com/shopspring/decimal"

// OptionChain contains the option contracts for a given symbol and expiration.
type OptionChain struct {
	Symbol          string
	UnderlyingPrice decimal.Decimal
	Expirations     []*Expiration
}

// Option represents an instance of an option contract.
type Option struct {
	ContractID    string
	Security      string
	Strike        decimal.Decimal
	Price         decimal.Decimal
	ChangeNominal decimal.Decimal
	ChangePercent decimal.Decimal
	Bid           decimal.Decimal
	Ask           decimal.Decimal
	Volume        int
	OpenInterest  int
}

// newOptionContract creates a new instance of an option.
func newOptionContract(option map[string]string) *Option {

	c := &Option{
		ContractID:    option["cid"],
		Security:      option["s"],
		Strike:        toDecimal(option["strike"]),
		Price:         toDecimal(option["p"]),
		ChangeNominal: toDecimal(option["c"]),
		Bid:           toDecimal(option["b"]),
		Ask:           toDecimal(option["a"]),
		Volume:        toInt(option["vol"]),
		OpenInterest:  toInt(option["oi"]),
	}

	if c.Price.IntPart() != 0 {
		hundred, _ := decimal.NewFromString("100")
		c.ChangePercent = ((c.ChangeNominal).Div(c.Price.Sub(c.ChangeNominal))).Mul(hundred).Truncate(2)
	}

	return c
}

// newContractSlice creates a new slice of contracts.
func newContractSlice(options []map[string]string) (contracts []*Option) {

	for _, op := range options {
		contracts = append(contracts, newOptionContract(op))
	}

	return contracts
}

func (chain *OptionChain) expirationExists(e *Expiration) bool {

	for _, exp := range chain.Expirations {
		if exp.Day == e.Day && exp.Month == e.Month && exp.Year == e.Year {
			return true
		}
	}
	return false
}
