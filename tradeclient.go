package main

import (
	"flag"
	"fmt"
	"os"
	"path"
    "time"

	"github.com/btasdoven/quickfixwebclient/internal"
	"github.com/quickfixgo/quickfix"
    
    fix42md "github.com/quickfixgo/quickfix/fix42/marketdatasnapshotfullrefresh"
)

//TradeClient implements the quickfix.Application interface
type TradeClient struct {
    *quickfix.MessageRouter
}

//OnCreate implemented as part of Application interface
func (e TradeClient) OnCreate(sessionID quickfix.SessionID) {
	return
}

//OnLogon implemented as part of Application interface
func (e TradeClient) OnLogon(sessionID quickfix.SessionID) {
	return
}

//OnLogout implemented as part of Application interface
func (e TradeClient) OnLogout(sessionID quickfix.SessionID) {
	return
}

//FromAdmin implemented as part of Application interface
func (e TradeClient) FromAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	return
}

//ToAdmin implemented as part of Application interface
func (e TradeClient) ToAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) {
	return
}

//ToApp implemented as part of Application interface
func (e TradeClient) ToApp(msg *quickfix.Message, sessionID quickfix.SessionID) (err error) {
	//msg.build()
	fmt.Printf("Sending %v\n", msg)
	return
}

//FromApp implemented as part of Application interface. This is the callback for all Application level messages from the counter party.
func (e TradeClient) FromApp(msg *quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	fmt.Printf("Received %v\n", msg)
    e.Route(msg, sessionID)
	return
}

func (e *TradeClient) OnFIX42MarketData(msg fix42md.MarketDataSnapshotFullRefresh, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
    fmt.Println("Geldi\n")
    return
}

func main() {
	flag.Parse()

	cfgFileName := path.Join("config", "tradeclient.cfg")
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

	app := TradeClient{MessageRouter: quickfix.NewMessageRouter()}
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
	    err = internal.QueryMarketDataRequest()
    
		if err != nil {
			fmt.Printf("%v\n", err)
		}

        time.Sleep(1000 * time.Millisecond)
	}

	initiator.Stop()
}
