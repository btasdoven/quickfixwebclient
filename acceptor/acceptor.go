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

    "github.com/FlashBoys/go-finance"

    fix42nos "github.com/quickfixgo/quickfix/fix42/newordersingle"
    fix42er "github.com/quickfixgo/quickfix/fix42/executionreport"
    fix42mdr "github.com/quickfixgo/quickfix/fix42/marketdatarequest"
    fix42md  "github.com/quickfixgo/quickfix/fix42/marketdatasnapshotfullrefresh"
)

type executor struct {
    orderID int
    execID  int
    *quickfix.MessageRouter
    quotes map[string]*Quote
}

type BidAsk struct {
    price   decimal.Decimal
    size    decimal.Decimal
}

type Quote struct {
    symbol      string
    trade       BidAsk
    bids        []BidAsk
    asks        []BidAsk
}

func (e *executor) getQuote(symbol string) *Quote {
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

    return
}

func (e *executor) OnFIX42NewOrderSingle(msg fix42nos.NewOrderSingle, sessionID quickfix.SessionID) (err quickfix.MessageRejectError) {
    ordType, err := msg.GetOrdType()
    if err != nil {
        return err
    }

    if ordType != enum.OrdType_LIMIT {
        return quickfix.ValueIsIncorrect(tag.OrdType)
    }

    symbol, err := msg.GetSymbol()
    if err != nil {
        return
    }

    side, err := msg.GetSide()
    if err != nil {
        return
    }

    orderQty, err := msg.GetOrderQty()
    if err != nil {
        return
    }

    price, err := msg.GetPrice()
    if err != nil {
        return
    }

    clOrdID, err := msg.GetClOrdID()
    if err != nil {
        return
    }

    stock := e.getQuote(symbol)
    leavesQty := orderQty
    orderStatus := enum.OrdStatus_NEW
    totalPrice := decimal.Zero
    lastPrice := decimal.Zero
    lastShares := decimal.Zero

    switch side {
    case enum.Side_BUY:
        for leavesQty.IntPart() > 0 && len(stock.asks) > 0 && stock.asks[0].price.Cmp(price) <= 0 {
            lastShares = decimal.Min(stock.asks[0].size, leavesQty)
            lastPrice = stock.asks[0].price

            leavesQty = leavesQty.Sub(lastShares)
            totalPrice = totalPrice.Add(lastPrice)

            stock.trade.price = lastPrice
            stock.trade.size = lastShares
            stock.asks[0].size = stock.asks[0].size.Sub(lastShares)
            if stock.asks[0].size.Cmp(decimal.Zero) == 0 {
                stock.asks = stock.asks[1:]
            }
        }
    }

    execQty := orderQty.Sub(leavesQty)
    avgPrice := decimal.Zero
    if execQty.Cmp(decimal.Zero) != 0 {
        avgPrice = totalPrice.Div(execQty)
    }

    if execQty.Equals(decimal.Zero) {
        orderStatus = enum.OrdStatus_CALCULATED
    } else if execQty.Equals(orderQty) {
        orderStatus = enum.OrdStatus_FILLED
    } else {
        orderStatus = enum.OrdStatus_PARTIALLY_FILLED
    }

    execReport := fix42er.New(
        e.genOrderID(),
        e.genExecID(),
        field.NewExecTransType(enum.ExecTransType_NEW),
        field.NewExecType(enum.ExecType_FILL),
        field.NewOrdStatus(orderStatus),
        field.NewSymbol(symbol),
        field.NewSide(side),
        field.NewLeavesQty(leavesQty, 2),
        field.NewCumQty(execQty, 2),
        field.NewAvgPx(avgPrice, 2),
    )

    execReport.SetClOrdID(clOrdID)
    execReport.SetOrderQty(orderQty, 2)
    execReport.SetLastShares(lastShares, 2)
    execReport.SetLastPx(lastPrice, 2)

    if msg.HasAccount() {
        acct, err := msg.GetAccount()
        if err != nil {
            return err
        }
        execReport.SetAccount(acct)
    }

    quickfix.SendToTarget(execReport, sessionID)

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
