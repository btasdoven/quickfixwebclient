package internal

import (
	"github.com/quickfixgo/quickfix"
	"github.com/quickfixgo/quickfix/enum"
	"github.com/quickfixgo/quickfix/field"
	fix42mdr "github.com/quickfixgo/quickfix/fix42/marketdatarequest"
)

func querySenderCompID() field.SenderCompIDField {
	return field.NewSenderCompID("WEBUI")
}

func queryTargetCompID() field.TargetCompIDField {
	return field.NewTargetCompID("FIXIMULATOR")
}

func queryTargetSubID() field.TargetSubIDField {
	return field.NewTargetSubID("TargetSubID")
}

type header interface {
	Set(f quickfix.FieldWriter) *quickfix.FieldMap
}

func queryHeader(h header) {
	h.Set(querySenderCompID())
	h.Set(queryTargetCompID())
	//h.Set(queryTargetSubID())
}

func queryMarketDataRequest42() fix42mdr.MarketDataRequest {
	request := fix42mdr.New(field.NewMDReqID("MARKETDATAID"),
		field.NewSubscriptionRequestType(enum.SubscriptionRequestType_SNAPSHOT),
		field.NewMarketDepth(0),
	)

	entryTypes := fix42mdr.NewNoMDEntryTypesRepeatingGroup()
	entryTypes.Add().SetMDEntryType(enum.MDEntryType_BID)
	request.SetNoMDEntryTypes(entryTypes)

	relatedSym := fix42mdr.NewNoRelatedSymRepeatingGroup()
	relatedSym.Add().SetSymbol("LNUX")
	request.SetNoRelatedSym(relatedSym)

	queryHeader(request.Header)
	return request
}

func QueryMarketDataRequest() error {

    req := queryMarketDataRequest42()

    return quickfix.Send(req)
}
