package main

import (
    "net/http"
    "time"
    "fmt"

    mux "github.com/gorilla/mux"

    "github.com/quickfixgo/quickfix/enum"
    "html/template"
)

var initiator Initiator

func restStockHandler(w http.ResponseWriter, r *http.Request) {
    symbolReq := r.URL.Query().Get("symbol")
    fmt.Printf("sym: %v", symbolReq)
    reqId := time.Now().String()

    msg := initiator.queryMarketDataRequest42(reqId, symbolReq)

    symbol, _ := msg.GetSymbol()
    noMDEntries, _ := msg.GetNoMDEntries()

    fmt.Fprintf(w, "%v\n", symbol)

    for i := 0; i < noMDEntries.Len(); i++ {
        entry := noMDEntries.Get(i)
        entryType, _ := entry.GetMDEntryType();
        entryPx, _ := entry.GetMDEntryPx()
        entrySize, _ := entry.GetMDEntrySize()

        var typeStr string

        switch entryType {
        case enum.MDEntryType_BID:
            typeStr = "Bid"
        case enum.MDEntryType_OFFER:
            typeStr = "Offer"
        case enum.MDEntryType_TRADE:
            typeStr = "Trade"
        }

        fmt.Fprintf(w, "%v: %v x %v\n", typeStr, entryPx, entrySize)
    }
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("req")
    t, _ := template.New("").ParseFiles("home.tpl")
    fmt.Printf("req 2")
    err := t.ExecuteTemplate(w, "home.tpl", nil)

    if err != nil {
        panic(err)
    }

    fmt.Printf("req 3")
}

func main() {
	initiator = NewInitiator()
    defer initiator.Stop()

    r := mux.NewRouter()
    r.HandleFunc("/", handler)
    r.HandleFunc("/marketData", restStockHandler)

    srv := &http.Server{
        Handler:      r,
        Addr:         "localhost:8080",
        // Good practice: enforce timeouts for servers you create!
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

    srv.ListenAndServe()
}
