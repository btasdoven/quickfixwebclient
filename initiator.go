package main

import (
    "flag"
    "fmt"
    "os"
    "path"

    "github.com/quickfixgo/quickfix/field"
    "github.com/quickfixgo/quickfix/enum"
    "github.com/quickfixgo/quickfix"
    
    fix42md "github.com/quickfixgo/quickfix/fix42/marketdatasnapshotfullrefresh"
    fix42mdr "github.com/quickfixgo/quickfix/fix42/marketdatarequest"
)

//Initiator implements the quickfix.Application interface
type Initiator struct {
    *quickfix.MessageRouter
    Initiator *quickfix.Initiator
    Callbacks map[string]chan interface{}
}

func NewInitiator() (app Initiator) {
    flag.Parse()

    cfgFileName := path.Join("config", "Initiator.cfg")
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

    app = Initiator{MessageRouter: quickfix.NewMessageRouter(), Callbacks: make(map[string]chan interface{})}

    app.AddRoute(fix42md.Route(app.OnFIX42MarketData))

    fileLogFactory, err := quickfix.NewFileLogFactory(appSettings)

    if err != nil {
        fmt.Println("Error creating file log factory,", err)
        return
    }

    app.Initiator, err = quickfix.NewInitiator(app, quickfix.NewMemoryStoreFactory(), appSettings, fileLogFactory)
    if err != nil {
        fmt.Printf("Unable to create Initiator: %s\n", err)
        return
    }

    app.Initiator.Start()

    return app
}

func (e Initiator) Stop() {
    e.Initiator.Stop()
}

//OnCreate implemented as part of Application interface
func (e Initiator) OnCreate(sessionID quickfix.SessionID) {
    return
}

//OnLogon implemented as part of Application interface
func (e Initiator) OnLogon(sessionID quickfix.SessionID) {
    return
}

//OnLogout implemented as part of Application interface
func (e Initiator) OnLogout(sessionID quickfix.SessionID) {
    return
}

//FromAdmin implemented as part of Application interface
func (e Initiator) FromAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
    return
}

//ToAdmin implemented as part of Application interface
func (e Initiator) ToAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) {
    return
}

//ToApp implemented as part of Application interface
func (e Initiator) ToApp(msg *quickfix.Message, sessionID quickfix.SessionID) (err error) {
    fmt.Printf("\tSending %v\n", msg)
    return
}

//FromApp implemented as part of Application interface. This is the callback for all Application level messages from the counter party.
func (e Initiator) FromApp(msg *quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
    fmt.Printf("\tReceived %v\n", msg)
    e.Route(msg, sessionID)
    return
}

func querySenderCompID() field.SenderCompIDField {
    return field.NewSenderCompID("WEBUI")
}

func queryTargetCompID() field.TargetCompIDField {
    return field.NewTargetCompID("FIXIMULATOR")
}

type header interface {
    Set(f quickfix.FieldWriter) *quickfix.FieldMap
}

func queryHeader(h header) {
    h.Set(querySenderCompID())
    h.Set(queryTargetCompID())
    //h.Set(queryTargetSubID())
}

func (e *Initiator) OnFIX42MarketData(msg fix42md.MarketDataSnapshotFullRefresh, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
    reqId, _ := msg.GetMDReqID()
    e.Callbacks[reqId] <- msg
    return
}

func (e Initiator) queryMarketDataRequest42(requestId string, symbol string) fix42md.MarketDataSnapshotFullRefresh {
    request := fix42mdr.New(field.NewMDReqID(requestId),
        field.NewSubscriptionRequestType(enum.SubscriptionRequestType_SNAPSHOT),
        field.NewMarketDepth(0),
    )

    request.SetMDUpdateType(enum.MDUpdateType_FULL_REFRESH)

    entryTypes := fix42mdr.NewNoMDEntryTypesRepeatingGroup()
    entryTypes.Add().SetMDEntryType(enum.MDEntryType_BID)
    entryTypes.Add().SetMDEntryType(enum.MDEntryType_OFFER)
    entryTypes.Add().SetMDEntryType(enum.MDEntryType_TRADE)
    request.SetNoMDEntryTypes(entryTypes)

    relatedSym := fix42mdr.NewNoRelatedSymRepeatingGroup()
    relatedSym.Add().SetSymbol(symbol)
    request.SetNoRelatedSym(relatedSym)

    queryHeader(request.Header)

    e.Callbacks[requestId] = make(chan interface{})
    defer delete(e.Callbacks, requestId)

    go quickfix.Send(request)

    fmt.Printf("\tWaiting response for request %+v", request)
    res := (<- e.Callbacks[requestId]).(fix42md.MarketDataSnapshotFullRefresh)
    fmt.Printf("\tResponse recieved: %+v", requestId, res)
    return res
}