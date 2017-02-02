package main

import (
    "flag"
    "fmt"
    "path"
    "os"
    "os/signal"
    "strconv"

    "github.com/quickfixgo/quickfix"
    "github.com/quickfixgo/quickfix/enum"
    "github.com/quickfixgo/quickfix/field"
    "github.com/quickfixgo/quickfix/tag"
    "github.com/shopspring/decimal"

    "github.com/btasdoven/go-finance"

    fix42nos "github.com/quickfixgo/quickfix/fix42/newordersingle"
    fix42osr "github.com/quickfixgo/quickfix/fix42/orderstatusrequest"
    fix42er "github.com/quickfixgo/quickfix/fix42/executionreport"
    fix42mdr "github.com/quickfixgo/quickfix/fix42/marketdatarequest"
    fix42md  "github.com/quickfixgo/quickfix/fix42/marketdatasnapshotfullrefresh"
    "sort"
)

type executor struct {
    orderID int
    execID  int
    *quickfix.MessageRouter
    quotes map[string]*Quote
    orders []*Order
}

type Order struct {
    ClOrdID       string
    ExecID        string
    ExecType      enum.ExecType
    ExecTransType enum.ExecTransType
    OrderStatus   enum.OrdStatus
    OrdType       enum.OrdType
    Side          enum.Side
    Symbol        string

    Price       decimal.Decimal
    OrderQty    decimal.Decimal

    LeavesQty   decimal.Decimal
    CumQty      decimal.Decimal
    AvgPx       decimal.Decimal
    TotalPrice  decimal.Decimal

    LastPrice   decimal.Decimal
    LastShares  decimal.Decimal
}

func (o *Order) Process(price decimal.Decimal, quantity decimal.Decimal) {
    if (o.Side == enum.Side_BUY && o.Price.Cmp(price) < 0) ||
        (o.Side == enum.Side_SELL && o.Price.Cmp(price) > 0) {
        return
    }

    qtyToProcess := decimal.Min(o.LeavesQty, quantity)

    if qtyToProcess.Cmp(decimal.Zero) <= 0 {
        return
    }

    o.LastPrice = price
    o.LastShares = qtyToProcess
    o.CumQty = o.CumQty.Add(qtyToProcess)
    o.LeavesQty = o.LeavesQty.Sub(qtyToProcess)
    o.TotalPrice.Add(price.Mul(qtyToProcess))
    o.AvgPx = o.TotalPrice.Div(o.CumQty)

    if o.CumQty.Equals(decimal.Zero) {
        o.OrderStatus = enum.OrdStatus_NEW
    } else if o.CumQty.Equals(o.OrderQty) {
        o.OrderStatus = enum.OrdStatus_FILLED
    } else {
        o.OrderStatus = enum.OrdStatus_PARTIALLY_FILLED
    }
}

type BidAsk struct {
    price   decimal.Decimal
    size    decimal.Decimal
    order   *Order
}

type AskList []BidAsk
type BidList []BidAsk

func (a AskList) Len() int           { return len(a) }
func (a AskList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a AskList) Less(i, j int) bool { return a[i].price.Cmp(a[j].price) < 0 }

func (a BidList) Len() int           { return len(a) }
func (a BidList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BidList) Less(i, j int) bool { return a[i].price.Cmp(a[j].price) > 0 }


func (e *executor) DumpOrders() {
    fmt.Printf("\n---------------------------------------\n")
    for _, q := range e.quotes {
        fmt.Printf("\t%v\n", q.symbol)
        for i := len(q.bids) - 1; i >= 0; i-- {
            fmt.Printf("\t\t(%v %v) %+v\n", q.bids[i].price, q.bids[i].size, q.bids[i].order)
        }
        fmt.Printf("\t\t--> %v %v %v\n", q.trade.price, q.trade.size, q.trade.order)

        for _, a := range q.asks {
            fmt.Printf("\t\t(%v %v) %+v\n", a.price, a.size, a.order)
        }

        fmt.Printf("\n")
    }

    for _, o := range e.orders {
        fmt.Printf("order: %+v\n", o)
    }

    fmt.Printf("\n\n")
}

type Quote struct {
    symbol      string
    trade       BidAsk
    bids        []BidAsk
    asks        []BidAsk
}

func (e *executor) getQuote(symbol string) *Quote {
    fmt.Printf("---Symbol--- %v --- %v\n", symbol, len(e.quotes))
    if _, ok := e.quotes[symbol]; !ok {
        stock, _ := finance.GetQuote(symbol)
        e.quotes[symbol] = &Quote{
            symbol: symbol,
            trade: BidAsk{
                price: stock.LastTradePrice,
                size: decimal.New(int64(stock.LastTradeSize), 0)},
            asks: make([]BidAsk, 2),
            bids: make([]BidAsk, 1),
        }

        e.quotes[symbol].asks[0].price = stock.Ask
        e.quotes[symbol].bids[0].price = stock.Bid
        e.quotes[symbol].asks[0].size = decimal.New(int64(stock.AskSize), 0)
        e.quotes[symbol].bids[0].size = decimal.New(int64(stock.BidSize), 0)

        e.quotes[symbol].asks[1].price = stock.Ask.Add(decimal.New(4, 0))
        e.quotes[symbol].asks[1].size = decimal.New(12, 0)
    }

    return e.quotes[symbol]
}

func newExecutor() *executor {
    e := &executor{MessageRouter: quickfix.NewMessageRouter()}
    e.AddRoute(fix42nos.Route(e.OnFIX42NewOrderSingle))
    e.AddRoute(fix42mdr.Route(e.OnFIX42MarketDataRequest))
    e.AddRoute(fix42osr.Route(e.OnFIX42OrderStatusRequest))

    e.quotes = make(map[string]*Quote)
    return e
}

func (e *executor) genOrderID() field.OrderIDField {
    e.orderID++
    return field.NewOrderID(strconv.Itoa(e.orderID))
}

func (e *executor) genExecID() field.ExecIDField {
    e.execID++
    return field.NewExecID(strconv.Itoa(e.execID))
}

//quickfix.Application interface
func (e executor) OnCreate(sessionID quickfix.SessionID)                           { return }
func (e executor) OnLogon(sessionID quickfix.SessionID)                            { return }
func (e executor) OnLogout(sessionID quickfix.SessionID)                           { return }
func (e executor) ToAdmin(msg *quickfix.Message, sessionID quickfix.SessionID)     { return }
func (e executor) ToApp(msg *quickfix.Message, sessionID quickfix.SessionID) error { return nil }
func (e executor) FromAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) quickfix.MessageRejectError {
    return nil
}

//Use Message Cracker on Incoming Application Messages
func (e *executor) FromApp(msg *quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
    fmt.Printf("Received %v\n", msg)
    return e.Route(msg, sessionID)
}

func (e *executor) OnFIX42MarketDataRequest(msg fix42mdr.MarketDataRequest, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
    fmt.Printf("[SERVER] - MDR: %+v\n", msg.Message)

    mdReqID, _ := msg.GetMDReqID()

    noRelatedSym, _ := msg.GetNoRelatedSym()

    length := noRelatedSym.Len()

    if length <= 0 {
        return
    }

    symbol, _ := noRelatedSym.Get(0).GetSymbol()
    fmt.Printf("\tSymbol: %+v\n", symbol)

    stock := e.getQuote(symbol)
    fmt.Printf("\tStock: %+v\n", stock)

    noMDEntryTypes, _ := msg.GetNoMDEntryTypes()
    noMDEntries := fix42md.NewNoMDEntriesRepeatingGroup()

    for i := 0; i < noMDEntryTypes.Len(); i++ {
        noMDEntryType, _ := noMDEntryTypes.Get(i).GetMDEntryType()
        noMDEntries.Add().SetMDEntryType(noMDEntryType)

        switch noMDEntryType {
        case enum.MDEntryType_BID:
            bid := BidAsk{}
            if len(stock.bids) > 0 {
                bid = stock.bids[0]
            }
            noMDEntries.Get(i).SetMDEntryPx(bid.price, 5)
            noMDEntries.Get(i).SetMDEntrySize(bid.size, 5)
        case enum.MDEntryType_OFFER:
            ask := BidAsk{}
            if len(stock.asks) > 0 {
                ask = stock.asks[0]
            }
            noMDEntries.Get(i).SetMDEntryPx(ask.price, 5)
            noMDEntries.Get(i).SetMDEntrySize(ask.size, 5)
        case enum.MDEntryType_TRADE:
            noMDEntries.Get(i).SetMDEntryPx(stock.trade.price, 5)
            noMDEntries.Get(i).SetMDEntrySize(stock.trade.size, 5)
        }
    }

    md := fix42md.New(field.NewSymbol(symbol))
    md.SetMDReqID(mdReqID)    
    md.SetNoMDEntries(noMDEntries)    

    fmt.Printf("\tSending %+v", md)
    quickfix.SendToTarget(md, sessionID)

    e.DumpOrders()

    return
}

func (e *executor) OnFIX42NewOrderSingle(msg fix42nos.NewOrderSingle, sessionID quickfix.SessionID) (err quickfix.MessageRejectError) {

    var order Order

    order.OrdType, err = msg.GetOrdType()
    if err != nil {
        return err
    }

    if order.OrdType != enum.OrdType_LIMIT {
        return quickfix.ValueIsIncorrect(tag.OrdType)
    }

    order.Symbol, err = msg.GetSymbol()
    if err != nil {
        return
    }

    order.Side, err = msg.GetSide()
    if err != nil {
        return
    }

    order.OrderQty, err = msg.GetOrderQty()
    if err != nil {
        return
    }

    order.Price, err = msg.GetPrice()
    if err != nil {
        return
    }

    order.ClOrdID, err = msg.GetClOrdID()
    if err != nil {
        return
    }

    stock := e.getQuote(order.Symbol)
    order.LeavesQty = order.OrderQty
    order.OrderStatus = enum.OrdStatus_NEW
    order.ExecTransType = enum.ExecTransType_NEW
    order.ExecType = enum.ExecType_FILL
    order.TotalPrice = decimal.Zero
    order.LastPrice = decimal.Zero
    order.LastShares = decimal.Zero
    order.ExecID = e.genExecID().Value()

    switch order.Side {
    case enum.Side_BUY:
        for order.LeavesQty.IntPart() > 0 && len(stock.asks) > 0 && stock.asks[0].price.Cmp(order.Price) <= 0 {
            order.Process(stock.asks[0].price, stock.asks[0].size)

            stock.trade.price = order.LastPrice
            stock.trade.size = order.LastShares
            stock.trade.order = &order

            stock.asks[0].size = stock.asks[0].size.Sub(order.LastShares)
            if stock.asks[0].size.Cmp(decimal.Zero) == 0 {
                if stock.asks[0].order != nil {
                    stock.asks[0].order.Process(order.LastPrice, order.LastShares)
                }
                stock.asks = stock.asks[1:]
            }
        }
    case enum.Side_SELL:
        for order.LeavesQty.IntPart() > 0 && len(stock.bids) > 0 && stock.bids[0].price.Cmp(order.Price) >= 0 {
            order.Process(stock.bids[0].price, stock.bids[0].size)
            stock.bids[0].order.Process(order.LastPrice, order.LastShares)

            stock.trade.price = order.LastPrice
            stock.trade.size = order.LastShares
            stock.trade.order = &order

            stock.bids[0].size = stock.bids[0].size.Sub(order.LastShares)
            if stock.bids[0].size.Cmp(decimal.Zero) == 0 {
                if stock.asks[0].order != nil {
                    stock.bids[0].order.Process(order.LastPrice, order.LastShares)
                }
                stock.bids = stock.bids[1:]
            }
        }
    }

    if order.OrderStatus != enum.OrdStatus_FILLED {
        bidAsk := BidAsk{order: &order, price: order.Price, size: order.LeavesQty}

        switch order.Side {
        case enum.Side_BUY:
            e.quotes[order.Symbol].bids = append(e.quotes[order.Symbol].bids, bidAsk)
            sort.Sort(BidList(e.quotes[order.Symbol].bids))
        case enum.Side_SELL:
            e.quotes[order.Symbol].asks = append(e.quotes[order.Symbol].asks, bidAsk)
            sort.Sort(AskList(e.quotes[order.Symbol].asks))
        }
    }

    execReport := fix42er.New(
        field.NewOrderID(order.ClOrdID),
        field.NewExecID(order.ExecID),
        field.NewExecTransType(order.ExecTransType),
        field.NewExecType(order.ExecType),
        field.NewOrdStatus(order.OrderStatus),
        field.NewSymbol(order.Symbol),
        field.NewSide(order.Side),
        field.NewLeavesQty(order.LeavesQty, 2),
        field.NewCumQty(order.CumQty, 2),
        field.NewAvgPx(order.AvgPx, 2),
    )

    execReport.SetClOrdID(order.ClOrdID)
    execReport.SetOrderQty(order.OrderQty, 2)
    execReport.SetLastShares(order.LastShares, 2)
    execReport.SetLastPx(order.LastPrice, 2)

    e.orders = append(e.orders, &order)

    if msg.HasAccount() {
        acct, err := msg.GetAccount()
        if err != nil {
            return err
        }
        execReport.SetAccount(acct)
    }

    quickfix.SendToTarget(execReport, sessionID)

    e.DumpOrders()
    return
}

func (e *executor) OnFIX42OrderStatusRequest(msg fix42osr.OrderStatusRequest, sessionID quickfix.SessionID) (err quickfix.MessageRejectError) {

    clOrdID, _ := msg.GetClOrdID()
    //symbol, _ := msg.GetSymbol()
    //side, _ := msg.GetSide()

    fmt.Printf("[SERVER]: OrderStatusRequest %v\n", clOrdID)

    var order *Order
    for _, o := range e.orders {
        if o.ClOrdID == clOrdID {
            order = o
            break
        }
    }

    fmt.Printf("Order: %v\n", order)

    execReport := fix42er.New(
        field.NewOrderID(order.ClOrdID),
        field.NewExecID(order.ExecID),
        field.NewExecTransType(order.ExecTransType),
        field.NewExecType(order.ExecType),
        field.NewOrdStatus(order.OrderStatus),
        field.NewSymbol(order.Symbol),
        field.NewSide(order.Side),
        field.NewLeavesQty(order.LeavesQty, 2),
        field.NewCumQty(order.CumQty, 2),
        field.NewAvgPx(order.AvgPx, 2),
    )

    execReport.SetClOrdID(order.ClOrdID)
    execReport.SetOrderQty(order.OrderQty, 2)
    execReport.SetLastShares(order.LastShares, 2)
    execReport.SetLastPx(order.LastPrice, 2)
    execReport.SetPrice(order.Price, 2)

    if msg.HasAccount() {
        acct, err := msg.GetAccount()
        if err != nil {
            return err
        }
        execReport.SetAccount(acct)
    }

    quickfix.SendToTarget(execReport, sessionID)

    e.DumpOrders()
    return
}

func main() {
    flag.Parse()

    cfgFileName := path.Join("config", "acceptor.cfg")
    if flag.NArg() > 0 {
        cfgFileName = flag.Arg(0)
    }

    cfg, err := os.Open(cfgFileName)
    if err != nil {
        fmt.Printf("Error opening %v, %v\n", cfgFileName, err)
        return
    }

    appSettings, err := quickfix.ParseSettings(cfg)
    if err != nil {
        fmt.Println("Error reading cfg,", err)
        return
    }

    logFactory := quickfix.NewScreenLogFactory()
    app := newExecutor()

    acceptor, err := quickfix.NewAcceptor(app, quickfix.NewMemoryStoreFactory(), appSettings, logFactory)
    if err != nil {
        fmt.Printf("Unable to create Acceptor: %s\n", err)
        return
    }

    err = acceptor.Start()
    if err != nil {
        fmt.Printf("Unable to start Acceptor: %s\n", err)
        return
    }

    interrupt := make(chan os.Signal)
    signal.Notify(interrupt, os.Interrupt, os.Kill)
    <-interrupt

    acceptor.Stop()
}
