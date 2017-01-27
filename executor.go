package main

import (
	"flag"
	"fmt"
	"path"

	"github.com/quickfixgo/quickfix"
	"github.com/quickfixgo/quickfix/enum"
	"github.com/quickfixgo/quickfix/field"
	"github.com/quickfixgo/quickfix/tag"
	"github.com/shopspring/decimal"

	fix42nos "github.com/quickfixgo/quickfix/fix42/newordersingle"
	fix42er "github.com/quickfixgo/quickfix/fix42/executionreport"
    fix42mdr "github.com/quickfixgo/quickfix/fix42/marketdatarequest"
    fix42md  "github.com/quickfixgo/quickfix/fix42/marketdatasnapshotfullrefresh"

	"os"
	"os/signal"
	"strconv"
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
	return e.Route(msg, sessionID)
}

func (e *executor) OnFIX42MarketDataRequest(msg fix42mdr.MarketDataRequest, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
    fmt.Printf("[SERVER] - Received %#v\n", msg)

    noRelatedSym, _ := msg.GetNoRelatedSym()
    fmt.Printf("%#v %v %+v\n", noRelatedSym.RepeatingGroup, noRelatedSym, noRelatedSym)

    symbol, _ := noRelatedSym.Get(0).GetSymbol()
    fmt.Printf("%+v\n", symbol)
    
    md := fix42md.New(field.NewSymbol(symbol))
    mdReqID, _ := msg.GetMDReqID()
    fmt.Println("%v", mdReqID)

    noMDEntryTypes, _ := msg.GetNoMDEntryTypes()
    fmt.Println("%v", noMDEntryTypes)

    noMDEntryType, _ := noMDEntryTypes.Get(0).GetMDEntryType()
    
    noMDEntries := fix42md.NewNoMDEntriesRepeatingGroup()
    noMDEntries.Add().SetMDEntryType(noMDEntryType)
    noMDEntries.Get(0).SetMDEntryPx(decimal.NewFromFloat(10), 5)

    md.SetMDReqID(mdReqID)    
    md.SetNoMDEntries(noMDEntries)    

    fmt.Println("[SERVER] - Sending %v", md)
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

	cfgFileName := path.Join("config", "executor.cfg")
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
