package finance

import (
	"encoding/csv"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

var (
	firstRegex  = regexp.MustCompile(`(\w+:)(\d+\.?\d*)`)
	secondRegex = regexp.MustCompile(`(\w+):`)
)

// requestCSV fetches a csv from a supplied URL.
func requestCSV(url string) ([][]string, error) {

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	reader.FieldsPerRecord = -1
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return data, nil
}

// buildURL takes a base URL and parameters returns the full URL.
func buildURL(base string, params map[string]string) string {

	url, _ := url.ParseRequestURI(base)
	q := url.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	url.RawQuery = q.Encode()

	return url.String()
}

// request fetches a file from a supplied URL.
func request(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	result := string(contents)

	return transformResult(result), err
}

// transformResult parses malformed json responses.
func transformResult(input string) []byte {

	json := firstRegex.ReplaceAllString(input, "$1\"$2\"")
	return []byte(secondRegex.ReplaceAllString(json, "\"$1\":"))
}
