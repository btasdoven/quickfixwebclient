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
}

func newExecutor() *executor {
    e := &executor{MessageRouter: quickfix.NewMessageRouter()}
    e.AddRoute(fix42nos.Route(e.OnFIX42NewOrderSingle))
    e.AddRoute(fix42mdr.Route(e.OnFIX42MarketDataRequest))

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

    stock, _ := finance.GetQuote(symbol)

    fmt.Printf("\tStock: %+v\n", stock)

    noMDEntryTypes, _ := msg.GetNoMDEntryTypes()
    noMDEntries := fix42md.NewNoMDEntriesRepeatingGroup()

    for i := 0; i < noMDEntryTypes.Len(); i++ {
        noMDEntryType, _ := noMDEntryTypes.Get(i).GetMDEntryType()
        noMDEntries.Add().SetMDEntryType(noMDEntryType)

        switch noMDEntryType {
        case enum.MDEntryType_BID:
            noMDEntries.Get(i).SetMDEntryPx(decimal.Decimal(stock.Bid), 5)
            noMDEntries.Get(i).SetMDEntrySize(decimal.NewFromFloat(float64(stock.BidSize)), 5)
        case enum.MDEntryType_OFFER:
            noMDEntries.Get(i).SetMDEntryPx(decimal.Decimal(stock.Ask), 5)
            noMDEntries.Get(i).SetMDEntrySize(decimal.NewFromFloat(float64(stock.AskSize)), 5)
        case enum.MDEntryType_TRADE:
            noMDEntries.Get(i).SetMDEntryPx(decimal.Decimal(stock.LastTradePrice), 5)
            noMDEntries.Get(i).SetMDEntrySize(decimal.NewFromFloat(float64(stock.LastTradeSize)), 5)
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

    execReport := fix42er.New(
        e.genOrderID(),
        e.genExecID(),
        field.NewExecTransType(enum.ExecTransType_NEW),
        field.NewExecType(enum.ExecType_FILL),
        field.NewOrdStatus(enum.OrdStatus_FILLED),
        field.NewSymbol(symbol),
        field.NewSide(side),
        field.NewLeavesQty(decimal.Zero, 2),
        field.NewCumQty(orderQty, 2),
        field.NewAvgPx(price, 2),
    )

    execReport.SetClOrdID(clOrdID)
    execReport.SetOrderQty(orderQty, 2)
    execReport.SetLastShares(orderQty, 2)
    execReport.SetLastPx(price, 2)

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
