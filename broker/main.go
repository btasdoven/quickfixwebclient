package main

import (
    "net/http"
    "time"
    "fmt"

    mux "github.com/gorilla/mux"

    "github.com/quickfixgo/quickfix/enum"
    "html/template"
    init2 "github.com/btasdoven/quickfixwebclient/broker/initiator"
)

var initiator init2.Initiator

func restStockHandler(w http.ResponseWriter, r *http.Request) {
    symbolReq := r.URL.Query().Get("symbol")
    fmt.Printf("sym: %v", symbolReq)
    reqId := time.Now().String()

    msg := initiator.QueryMarketDataRequest42(reqId, symbolReq)

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
    t, _ := template.New("").ParseFiles("home.tpl")
    err := t.ExecuteTemplate(w, "home.tpl", nil)

    if err != nil {
        panic(err)
    }
}

func main() {
	initiator = init2.NewInitiator()
    defer initiator.Stop()

    r := mux.NewRouter()
    r.HandleFunc("/", handler)
    r.HandleFunc("/marketData", restStockHandler)

    srv := &http.Server{
        Handler:      r,
        Addr:         ":9898",
        // Good practice: enforce timeouts for servers you create!
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

    srv.ListenAndServe()
}
