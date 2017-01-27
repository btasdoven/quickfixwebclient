package main

import (
	"flag"
	"fmt"
	"os"
	"path"
    "time"

	"github.com/quickfixgo/quickfix/field"
	"github.com/quickfixgo/quickfix/enum"
	"github.com/quickfixgo/quickfix"
    
    fix42md "github.com/quickfixgo/quickfix/fix42/marketdatasnapshotfullrefresh"
	fix42mdr "github.com/quickfixgo/quickfix/fix42/marketdatarequest"
)

//Initiator implements the quickfix.Application interface
type Initiator struct {
    *quickfix.MessageRouter
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
	fmt.Printf("Sending %v\n", msg)
	return
}

//FromApp implemented as part of Application interface. This is the callback for all Application level messages from the counter party.
func (e Initiator) FromApp(msg *quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	fmt.Printf("Received %v\n", msg)
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
	symbol, _ := msg.GetSymbol()

	noMDEntries, _ := msg.GetNoMDEntries()

	for i := 0; i < noMDEntries.Len(); i++ {
		entry := noMDEntries.Get(i)
		entryType, _ := entry.GetMDEntryType();
		entryPx, _ := entry.GetMDEntryPx()
		entrySize, _ := entry.GetMDEntrySize()

		fmt.Printf("%v: %v, %v %v\n", symbol, enum.MDEntryType(entryType), entryPx, entrySize)
	}

	return
}

func (e Initiator) QueryMarketDataRequest() error {
	req := queryMarketDataRequest42()
	return quickfix.Send(req)
}

func queryMarketDataRequest42() fix42mdr.MarketDataRequest {
	request := fix42mdr.New(field.NewMDReqID("MARKETDATAID"),
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
	relatedSym.Add().SetSymbol("MSFT")
	request.SetNoRelatedSym(relatedSym)

	queryHeader(request.Header)
	return request
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

	app = Initiator{MessageRouter: quickfix.NewMessageRouter()}
	app.AddRoute(fix42md.Route(app.OnFIX42MarketData))

	fileLogFactory, err := quickfix.NewFileLogFactory(appSettings)

	if err != nil {
		fmt.Println("Error creating file log factory,", err)
		return
	}

	initiator, err := quickfix.NewInitiator(app, quickfix.NewMemoryStoreFactory(), appSettings, fileLogFactory)
	if err != nil {
		fmt.Printf("Unable to create Initiator: %s\n", err)
		return
	}

	initiator.Start()

	for {
		err = app.QueryMarketDataRequest()

		if err != nil {
			fmt.Printf("%v\n", err)
		}

		time.Sleep(5000 * time.Millisecond)
	}

	initiator.Stop()
	return app
}