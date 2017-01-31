package initiator

import (
    "fmt"

    fix42md "github.com/quickfixgo/quickfix/fix42/marketdatasnapshotfullrefresh"
    fix42mdr "github.com/quickfixgo/quickfix/fix42/marketdatarequest"
    "github.com/quickfixgo/quickfix"
    "github.com/quickfixgo/quickfix/field"
    "github.com/quickfixgo/quickfix/enum"
)

func (e *Initiator) OnFIX42MarketData(msg fix42md.MarketDataSnapshotFullRefresh, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
    reqId, _ := msg.GetMDReqID()
    e.lock.Lock()
    e.Callbacks[reqId] <- msg
    e.lock.Unlock()
    return
}

func (e Initiator) QueryMarketDataRequest42(requestId string, symbol string) fix42md.MarketDataSnapshotFullRefresh {
    request := fix42mdr.New(
        field.NewMDReqID(requestId),
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

    e.lock.Lock()
    e.Callbacks[requestId] = make(chan interface{})
    e.lock.Unlock()
    defer close(e.Callbacks[requestId])
    defer delete(e.Callbacks, requestId)

    go quickfix.Send(request)

    fmt.Printf("\tWaiting response for request %+v", request)
    res := (<- e.Callbacks[requestId]).(fix42md.MarketDataSnapshotFullRefresh)
    fmt.Printf("\tResponse recieved: %+v %+v", requestId, res)

    return res
}