package finance

import "fmt"

const symbolsURL = "http://www.batstrading.com/market_data/symbol_listing/csv/"

// GetUSEquitySymbols fetches the symbols available through BATS, ~8k symbols.
func GetUSEquitySymbols() ([]string, error) {

	table, err := getSymbolsFromURL(symbolsURL)
	if err != nil {
		return []string{}, fmt.Errorf("error fetching symbols:  (error was: %s)\n", err.Error())
	}

	return processSymbols(table), nil
}

// getSymbolsFromURL fetches the csv from the endpoint.
func getSymbolsFromURL(url string) (table [][]string, err error) {

	return requestCSV(symbolsURL)
}

// processSymbols turns the raw table data of quotes into a slice of symbols.
func processSymbols(table [][]string) (symbols []string) {

	for idx, row := range table {
		if idx != 0 {
			symbols = append(symbols, row[0])
		}
	}
	return symbols
}
