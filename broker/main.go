package main

import (
    "html/template"
    "net/http"
    "time"
    "fmt"
    "strconv"
    "os"

    init2 "github.com/btasdoven/quickfixwebclient/broker/initiator"
    mux "github.com/gorilla/mux"
    "github.com/quickfixgo/quickfix/enum"
    "github.com/gorilla/sessions"
)

var initiator init2.Initiator
var store = sessions.NewCookieStore([]byte("something-very-secret"))

var orders []*Order

type Order struct {
    clOrdID string
    symbol  string
    side    enum.Side
}

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

func restOrderSingle(w http.ResponseWriter, r *http.Request) {
    symbolReq := r.URL.Query().Get("symbol")
    quantityReq, _ := strconv.Atoi(r.URL.Query().Get("quantity"))
    limitReq, _ := strconv.ParseFloat(r.URL.Query().Get("limit"), 64)
    sideReq := enum.Side(r.URL.Query().Get("side"))

    fmt.Printf("sym: %v ,q: %v, limit: %v", symbolReq, quantityReq, limitReq)

    orderId := time.Now().String()

    msg := initiator.QueryOrderSingleRequest(orderId, symbolReq, quantityReq, limitReq, sideReq)

    cumQty, _ := msg.GetCumQty()
    leavesQty, _ := msg.GetLeavesQty()
    lastPrice, _ := msg.GetLastPx()
    lastShares, _ := msg.GetLastShares()
    status, _ := msg.GetOrdStatus()
    side, _ := msg.GetSide()

    order := Order{
        clOrdID:orderId,
        symbol: symbolReq,
        side: side}

    orders = append(orders, &order)

    fmt.Fprintf(w, "Status: %v, Executed: %v, Remaining: %v, Last Price: %v, Last Shares: %v\n",
        status,
        cumQty,
        leavesQty,
        lastPrice,
        lastShares)
}

func restOrders(w http.ResponseWriter, r *http.Request) {
    if len(orders) > 0 {
        for _, order := range orders {
            fmt.Printf("Retrieveing order %v\n", order.clOrdID)
            msg := initiator.QueryOrderStatusRequest(order.clOrdID, order.symbol, order.side)

            cumQty, _ := msg.GetCumQty()
            leavesQty, _ := msg.GetLeavesQty()
            lastPrice, _ := msg.GetLastPx()
            lastShares, _ := msg.GetLastShares()
            status, _ := msg.GetOrdStatus()

            fmt.Fprintf(w, "Symbol: %v, Status: %v, Executed: %v, Remaining: %v, Last Price: %v, Last Shares: %v\n",
                order.symbol,
                status,
                cumQty,
                leavesQty,
                lastPrice,
                lastShares)
        }
    } else {
        fmt.Fprintf(w, "No order found")
        return
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
    r.HandleFunc("/marketData", restStockHandler).Methods("GET")
    r.HandleFunc("/orderSingle", restOrderSingle).Methods("GET")
    r.HandleFunc("/orders", restOrders).Methods("GET")

    srv := &http.Server{
        Handler:      r,
        Addr:         ":9898",
        // Good practice: enforce timeouts for servers you create!
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

    certFile := "/etc/letsencrypt/live/btasdoven.com/cert.pem"
    keyFile := "/etc/letsencrypt/live/btasdoven.com/privkey.pem"

    if _, err := os.Stat(certFile); os.IsNotExist(err) {
        certFile = ""
    }

    if _, err := os.Stat(keyFile); os.IsNotExist(err) {
        keyFile = ""
    }

    var err error
    if certFile != "" && keyFile != "" {
        fmt.Printf("Starting HTTPS server\n")
        err = srv.ListenAndServeTLS(certFile, keyFile)
    } else {
        fmt.Printf("Starting HTTP server\n")
        err = srv.ListenAndServe()
    }

    if err != nil {
        panic(err)
    }
}
