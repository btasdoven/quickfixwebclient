package initiator

import (
    "flag"
    "fmt"
    "os"
    "path"

    "github.com/quickfixgo/quickfix/field"
    "github.com/quickfixgo/quickfix"
    
    fix42md "github.com/quickfixgo/quickfix/fix42/marketdatasnapshotfullrefresh"
    fix42er "github.com/quickfixgo/quickfix/fix42/executionreport"
)

//Initiator implements the quickfix.Application interface
type Initiator struct {
    *quickfix.MessageRouter
    Initiator *quickfix.Initiator
    Callbacks map[string]chan interface{}
}

func NewInitiator() (app Initiator) {
    flag.Parse()

    cfgFileName := path.Join("config", "initiator.cfg")
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
    app.AddRoute(fix42er.Route(app.OnFIX42ExecutionReport))

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