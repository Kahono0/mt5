package mt5

import (
	"fmt"

	"github.com/kahono0/mt5/internal/ec"
)

type ReturnCode uint32

const (
	TRADE_RETCODE_REQUOTE        ReturnCode = 10004 // Requote
	TRADE_RETCODE_REJECT         ReturnCode = 10006 // Request rejected
	TRADE_RETCODE_CANCEL         ReturnCode = 10007 // Request canceled by trader
	TRADE_RETCODE_PLACED         ReturnCode = 10008 // Order placed
	TRADE_RETCODE_DONE           ReturnCode = 10009 // Request completed
	TRADE_RETCODE_DONE_PARTIAL   ReturnCode = 10010 // Only part of the request was completed
	TRADE_RETCODE_ERROR          ReturnCode = 10011 // Request processing error
	TRADE_RETCODE_TIMEOUT        ReturnCode = 10012 // Request canceled by timeout
	TRADE_RETCODE_INVALID        ReturnCode = 10013 // Invalid request
	TRADE_RETCODE_INVALID_VOLUME ReturnCode = 10014 // Invalid volume in the request
	TRADE_RETCODE_INVALID_PRICE  ReturnCode = 10015 // Invalid price in the request
	TRADE_RETCODE_INVALID_STOPS  ReturnCode = 10016 // Invalid stops in the request
)

type TradeAction uint32

const (
	TRADE_ACTION_DEAL     TradeAction = 3
	TRADE_ACTION_PENDING  TradeAction = 4
	TRADE_ACTION_SLTP     TradeAction = 5
	TRADE_ACTION_MODIFY   TradeAction = 6
	TRADE_ACTION_REMOVE   TradeAction = 7
	TRADE_ACTION_CLOSE_BY TradeAction = 8
)

type TradeEventRequest struct {
	TradeVolume    uint64
	PriceOrder     float64
	PriceTrigger   float64
	PriceSL        float64
	PriceTP        float64
	PriceDeviation uint32
	PriceTop       float64
	PriceBottom    float64
	ActionID       uint32
	TradeAction    TradeAction
	TradeSymbol    string
	Digits         uint32
	TradeOrder     int64
	TradeType      uint32
	TypeFilling    uint32
	TypeFlags      uint32
	Comment        string
	PositionBy     uint64
	SendTime       int
	Login          uint64
	Complete       bool
}

type TradeResult struct {
	Retcode    ReturnCode
	Deal       int64
	TradeOrder int64
	Volume     int64
	Price      float64
	ResTime    uint32
	Bid        float64
	Ask        float64
	Last       float64
	Comment    string
}

type TradeEvent struct {
	EventID      uint32
	TradeRequest *TradeEventRequest
	TradeResult  *TradeResult
}

func ParseTradeEvent(t []byte) (*TradeEvent, error) {
	parsed, err := ec.Parse(t, P, 0)
	if err != nil {
		return nil, err
	}

	return &TradeEvent{
		EventID:      parsed[0].(uint32),
		TradeRequest: parsed[1].(*TradeEventRequest),
		TradeResult:  parsed[2].(*TradeResult),
	}, nil
}

var Pp = []ec.Field{
	{PropType: 6},
	{PropType: 17},
	{PropType: 17},
	{PropType: 17},
	{PropType: 8},
	{PropType: 6},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 11, PropLength: 64},
}

var PpLength, _ = ec.GetSeriesSize(Pp)

var P = []ec.Field{
	{
		PropType: 6,
	},
	{
		PropType:   12,
		PropLength: Ep,
		Parser:     parseTradeEventRequest,
	},
	{
		PropType:   12,
		PropLength: PpLength,
		Parser:     parseTradeResult,
	},
}

var Ap = []ec.Field{
	{PropType: 6},
	{PropType: 6},
	{PropType: 11, PropLength: 64},
	{PropType: 18},
	{PropType: 6},
	{PropType: 17},
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 6},
	{PropType: 8},
	{PropType: 8},
	{PropType: 11, PropLength: 64},
	{PropType: 18},
	{PropType: 18},
	{PropType: 12, PropLength: 4},
}

var Ep, _ = ec.GetSeriesSize(Ap) // same as Pc.getSeriesSize(Ap)

func parseTradeEventRequest(t []byte, e int) (interface{}, error) {
	parsed, err := ec.Parse(t, Ap, e)
	if err != nil {
		return nil, err
	}

	if len(parsed) < 21 {
		return nil, fmt.Errorf("expected at least 21 values, got %d", len(parsed))
	}

	request := &TradeEventRequest{
		ActionID:    parsed[0].(uint32),
		TradeAction: TradeAction(parsed[1].(uint32)),
		TradeSymbol: parsed[2].(string),
		TradeVolume: parsed[3].(uint64),
		Digits:      parsed[4].(uint32),
		TradeOrder:  parsed[5].(int64),
		TradeType:   parsed[6].(uint32),
		TypeFilling: parsed[7].(uint32),
		TypeFlags:   parsed[9].(uint32),
		Comment:     parsed[18].(string),
		PositionBy:  parsed[19].(uint64),
	}

	request.PriceOrder = parsed[11].(float64)
	request.PriceTrigger = parsed[12].(float64)
	request.PriceSL = parsed[13].(float64)
	request.PriceTP = parsed[14].(float64)
	request.PriceDeviation = parsed[15].(uint32)
	request.PriceTop = parsed[16].(float64)
	request.PriceBottom = parsed[17].(float64)

	return request, nil
}

// parseTradeResult parses fields using Pp and casts directly
func parseTradeResult(t []byte, e int) (interface{}, error) {
	parsed, err := ec.Parse(t, Pp, e) // Parse returns []interface{}
	if err != nil {
		return nil, err
	}

	s := &TradeResult{}

	s.Retcode = ReturnCode(parsed[0].(uint32))
	s.Deal = parsed[1].(int64)
	s.TradeOrder = parsed[2].(int64)
	s.Volume = parsed[3].(int64)
	s.Price = parsed[4].(float64)
	s.ResTime = parsed[5].(uint32)
	s.Bid = parsed[6].(float64)
	s.Ask = parsed[7].(float64)
	s.Last = parsed[8].(float64)
	s.Comment = parsed[9].(string)

	return s, nil
}

type TradeEventHandler func(event *TradeEvent)

func (c *AppClient) RegisterTradeEventHandler(handler TradeEventHandler, opts ...HandlerOption) func() {
	if handler == nil {
		return func() {}
	}

	return c.RegisterHandler(CommandTradeEvent, func(msg Message) error {
		event, err := ParseTradeEvent(msg.Body)
		if err != nil {
			return err
		}
		handler(event)

		return nil
	}, opts...)
}
