package initiator

import (
    "fmt"

    fix42er "github.com/quickfixgo/quickfix/fix42/executionreport"
    fix42osr "github.com/quickfixgo/quickfix/fix42/orderstatusrequest"
    fix42nos "github.com/quickfixgo/quickfix/fix42/newordersingle"

    "time"
    "github.com/shopspring/decimal"
    "github.com/quickfixgo/quickfix"
    "github.com/quickfixgo/quickfix/enum"
    "github.com/quickfixgo/quickfix/field"
)

func (e *Initiator) OnFIX42ExecutionReport(msg fix42er.ExecutionReport, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
    orderId, _ := msg.GetClOrdID()

    e.lock.Lock()
    e.Callbacks[orderId] <- msg
    e.lock.Unlock()
    return
}

func (e Initiator) QueryOrderSingleRequest(
    orderId string,
    symbol string,
    quantity int,
    limit float64,
    side enum.Side) fix42er.ExecutionReport {
    request := fix42nos.New(
        field.NewClOrdID(orderId),
        field.NewHandlInst(enum.HandlInst_MANUAL_ORDER_BEST_EXECUTION),
        field.NewSymbol(symbol),
        field.NewSide(side),
        field.NewTransactTime(time.Now()),
        field.NewOrdType(enum.OrdType_LIMIT))

    request.SetOrderQty(decimal.New(int64(quantity), 0), 5)
    request.SetPrice(decimal.NewFromFloat(limit), 0)

    queryHeader(request.Header)

    e.lock.Lock()
    e.Callbacks[orderId] = make(chan interface{})
    e.lock.Unlock()
    defer close(e.Callbacks[orderId])
    defer delete(e.Callbacks, orderId)

    go quickfix.Send(request)

    fmt.Printf("\tQueryOrderSingleRequest Waiting response for request %+v", request)
    res := (<- e.Callbacks[orderId]).(fix42er.ExecutionReport)
    fmt.Printf("\tQueryOrderSingleRequest Response recieved: %+v %+v", orderId, res)

    return res
}

func (e Initiator) QueryOrderStatusRequest(orderId string, symbol string, side enum.Side) fix42er.ExecutionReport {
    request := fix42osr.New(
        field.NewClOrdID(orderId),
        field.NewSymbol(symbol),
        field.NewSide(side))

    queryHeader(request.Header)

    e.lock.Lock()
    e.Callbacks[orderId] = make(chan interface{})
    e.lock.Unlock()
    defer close(e.Callbacks[orderId])
    defer delete(e.Callbacks, orderId)

    go quickfix.Send(request)

    fmt.Printf("\tQueryOrderStatusRequest Waiting response for request %+v", request)
    res := (<- e.Callbacks[orderId]).(fix42er.ExecutionReport)
    fmt.Printf("\tQueryOrderStatusRequest Response recieved: %+v %+v", orderId, res)

    return res
}