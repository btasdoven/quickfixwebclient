package initiator

import (
    "github.com/quickfixgo/quickfix"
    "github.com/quickfixgo/quickfix/field"
    "github.com/quickfixgo/quickfix/enum"
    "fmt"

    fix42er "github.com/quickfixgo/quickfix/fix42/executionreport"
    fix42nos "github.com/quickfixgo/quickfix/fix42/newordersingle"
    "time"
    "github.com/shopspring/decimal"
)

func (e *Initiator) OnFIX42ExecutionReport(msg fix42er.ExecutionReport, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
    orderId, _ := msg.GetClOrdID()
    e.Callbacks[orderId] <- msg
    return
}

func (e Initiator) QueryOrderSingleRequest(orderId string, symbol string, quantity int, limit int) fix42er.ExecutionReport {
    request := fix42nos.New(
        field.NewClOrdID(orderId),
        field.NewHandlInst(enum.HandlInst_MANUAL_ORDER_BEST_EXECUTION),
        field.NewSymbol(symbol),
        field.NewSide(enum.Side_BUY),
        field.NewTransactTime(time.Now()),
        field.NewOrdType(enum.OrdType_LIMIT))

    request.SetOrderQty(decimal.New(int64(quantity), 0), 5)
    request.SetPrice(decimal.New(int64(limit), 0), 0)

    queryHeader(request.Header)

    e.Callbacks[orderId] = make(chan interface{})
    defer delete(e.Callbacks, orderId)

    go quickfix.Send(request)

    fmt.Printf("\tWaiting response for request %+v", request)
    res := (<- e.Callbacks[orderId]).(fix42er.ExecutionReport)
    fmt.Printf("\tResponse recieved: %+v %+v", orderId, res)

    return res
}