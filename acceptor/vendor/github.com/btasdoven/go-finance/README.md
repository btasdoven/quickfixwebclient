# go-finance

[![GoDoc](https://godoc.org/github.com/FlashBoys/go-finance?status.svg)](https://godoc.org/github.com/FlashBoys/go-finance)
[![Build Status](https://travis-ci.org/FlashBoys/go-finance.svg?branch=master)](https://travis-ci.org/FlashBoys/go-finance) [![codecov.io](https://codecov.io/github/FlashBoys/go-finance/coverage.svg?branch=master)](https://codecov.io/github/FlashBoys/go-finance?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/Flashboys/go-finance)](https://goreportcard.com/report/github.com/Flashboys/go-finance)
[![License MIT](https://img.shields.io/npm/l/express.svg)](http://opensource.org/licenses/MIT)

![codecov.io](https://codecov.io/github/FlashBoys/go-finance/branch.svg?branch=master)

`go-finance` is a Golang library for retrieving financial data for quantitative analysis.

To install go-finance, use the following command:

```
go get github.com/FlashBoys/go-finance
```


## Features

### Single security quotes

```go
package main

import (
	"fmt"

	"github.com/FlashBoys/go-finance"
)

func main() {
	// 15-min delayed full quote for Apple.
	q, err := finance.GetQuote("AAPL")
	if err == nil {
		fmt.Println(q)
	}
}
```

### Multiple securities quotes

```go
package main

import (
	"fmt"

	"github.com/FlashBoys/go-finance"
)

func main() {
	// 15-min delayed full quotes for Apple, Twitter, and Facebook.
	symbols := []string{"AAPL", "TWTR", "FB"}
	quotes, err := finance.GetQuotes(symbols)
	if err == nil {
		fmt.Println(quotes)
	}
}
```

### Quote history

```go
package main

import (
	"fmt"
	"time"

	"github.com/FlashBoys/go-finance"
)

func main() {
	// Set time bounds to 1 month starting Jan. 1.
	start, _ := time.Parse(time.RFC3339, "2016-01-01T16:00:00+00:00")
	end := start.AddDate(0, 1, 0)
	
	// Request daily history for TWTR.
	// IntervalDaily OR IntervalWeekly OR IntervalMonthly are supported.
	bars, err := finance.GetQuoteHistory("TWTR", start, end, finance.IntervalDaily)
	if err == nil {
		fmt.Println(bars)
	}
}
```

### Dividend/Split history

```go
package main

import (
	"fmt"
	"time"

	"github.com/FlashBoys/go-finance"
)

func main() {
	// Set time range from Jan 2010 up to the current date.
	// This example will return a slice of both dividends and splits.
	start, _ := time.Parse(time.RFC3339, "2010-01-01T16:00:00+00:00")
	end := time.Now()
	
	// Request event history for AAPL.
	events, err := finance.GetDividendSplitHistory("AAPL", start, end)
	if err == nil {
		fmt.Println(events)
	}
}
```

### Symbols download

```go
package main

import (
	"fmt"

	"github.com/FlashBoys/go-finance"
)

func main() {
	// Request all BATS symbols.
	symbols, err := finance.GetUSEquitySymbols()
	if err == nil {
		fmt.Println(symbols)
	}
}

```

### Options chains

```go
package main

import (
	"fmt"

	"github.com/FlashBoys/go-finance"
)

func main() {
	// Fetches the available expiration dates.
	chain, err := finance.NewOptionsChain("AAPL")
	if err != nil {
		panic(err)
	}

	// Some examples - see docs for full explanation.

	// Fetches puts and calls for the closest expiration date.
	calls, puts, err := chain.GetOptionsExpiringNext()
	if err == nil {
		panic(err)
  	}
	fmt.Println(calls)
	fmt.Println(puts)

	// Fetches puts and calls for the specified expiration date.
	calls, puts, err := chain.GetOptionsForExpiration(chain.Expirations[1])
	if err == nil {
		panic(err)
	}
	fmt.Println(calls)
	fmt.Println(puts)

	// Fetches calls for the specified expiration date.
	calls, err := chain.GetCallsForExpiration(chain.Expirations[1])
	if err == nil {
		panic(err)
	}
	fmt.Println(calls)
}

```

### Currency pairs quotes

```go
package main

import (
	"fmt"

	"github.com/FlashBoys/go-finance"
)

func main() {
	// Fetches the quote for USD/GBP pair.
	pq, err := finance.GetCurrencyPairQuote(finance.USDGBP)
	if err == nil {
		fmt.Println(pq)
	}
}

```

## Intentions

`go-finance` aims to provide a clean, flexible, way to retrieve financial data for your own projects. This is accomplished through an implementation of the full suite of Yahoo! Finance APIs and other sources. After much consideration and exploration of the labyrinthian endpoint system in the Yahoo Finance ecosystem, I've done my best to select and document proper consumption of the seemingly most stable endpoints.
The primary technical tenants of this project are:

  * Make financial data easy and fun to work with in Go-lang.
  * Abstract the burden of non-sexy model serialization away from the end-user.
  * Provide a mature framework where the end-user needs only be concerned with analysis instead of data sourcing.

There are several applications for this library. It's intentions are to be conducive to the following activities:

  * Quantitative financial analysis in Golang.
  * Academic study/comparison in a clean, easy language.
  * Algorithmic/Statistical-based strategy implementation.

## To-do

- [ ] Add greeks calculations to options data
- [ ] International securities quotes
- [ ] Sector/Industry components
- [ ] Indicies components
- [ ] Key stats (full profile) for securities

## Limitations (currently)

Given Yahoo! Finance's own perpetuation of a rabbit warren -like system of financial data YQL tables in varying degrees of deprecation, conflation/realtime, exchange availability, protocol access, and overall lack of consistent usage guidelines/documentation, I advise users of this library to be aware that you should not depend on it returning data to you 100% of the time. Build fail-safes and back-up plans into your own systems tasked with handling these cases as they arise. You should also probably complain to Yahoo to build better financial engineering tools since so many of us depend on them.

While dataframes (tabular data structures used for analytical operations atypical of what you see in the beaten track of web programming) are popular in the financial development community for use in prototyping models, those concepts are not the current focus of this project.

Yahoo also does not currently support a way to download a master list of symbols available. I compromised and am using the BATS list for now.

## Contributing

If you find this repo helpful, please give it a star. If you wish to discuss changes to it, please open an issue! This project is not as mature as it could be, and financial projects in Golang are in drastic need of some basic helpful dependencies as this repo aims to be.


## Similar Projects

I've taken features from the following projects, chosen for their stability, wide-spread usage, completeness and accuracy of the financial data we know to be publicly available, as well as how much I like using them in other projects given their own concise syntax inherent in their own design:

  * [pandas datareader](https://github.com/pydata/pandas-datareader) (Python) wide-spread use in academia.
  * [yahoofinance-api](https://github.com/sstrickx/yahoofinance-api) (Java) most popular java library for this purpose.
  * [quantmod](http://www.quantmod.com/) (R) a package for development/testing/deployment of quant strategy.
