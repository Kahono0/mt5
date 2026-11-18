package mt5

import (
	"fmt"

	"github.com/kahono0/mt5/internal/ec"
)

type TradeEvent struct {
	TradeVolume    uint64
	PriceOrder     float64
	PriceTrigger   float64
	PriceSL        float64
	PriceTP        float64
	PriceDeviation uint32
	PriceTop       float64
	PriceBottom    float64
	ActionID       uint32
	TradeAction    uint32
	TradeSymbol    string
	Digits         uint32
	TradeOrder     int64
	TradeType      uint32
	TypeFilling    uint32
	TypeFlags      uint32
	Comment        string
	Position       uint64
	SendTime       int
	Login          uint64
	Complete       bool
}

type yp struct {
	Retcode    int
	Deal       uint64
	TradeOrder uint64
	Volume     uint64
	Price      float64
	ResTime    uint64
	Bid        float64
	Ask        float64
	Last       float64
	Comment    string
}

func ParseTradeEvent(t []byte) ([]*TradeEvent, []*yp, error) {
	parsed, err := ec.Parse(t, P, 0)
	if err != nil {
		return nil, nil, err
	}

	// parsed is []interface{someint, *Sp, *yp}
	spList := []*TradeEvent{}
	ypList := []*yp{}
	for _, item := range parsed {
		switch v := item.(type) {
		case *TradeEvent:
			spList = append(spList, v)
		case *yp:
			ypList = append(ypList, v)
		}
	}
	return spList, ypList, nil

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
		Parser:     ParseSp,
	},
	{
		PropType:   12,
		PropLength: PpLength,
		Parser:     ParseYP,
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

func ParseSp(t []byte, e int) (interface{}, error) {
	parsed, err := ec.Parse(t, Ap, e)
	if err != nil {
		return nil, err
	}

	if len(parsed) < 21 {
		return nil, fmt.Errorf("expected at least 21 values, got %d", len(parsed))
	}

	sp := &TradeEvent{
		ActionID:    parsed[0].(uint32),
		TradeAction: parsed[1].(uint32),
		TradeSymbol: parsed[2].(string),
		TradeVolume: parsed[3].(uint64),
		Digits:      parsed[4].(uint32),
		TradeOrder:  parsed[5].(int64),
		TradeType:   parsed[6].(uint32),
		TypeFilling: parsed[7].(uint32),
		TypeFlags:   parsed[9].(uint32),
		Comment:     parsed[18].(string),
		Position:    parsed[19].(uint64),
	}

	sp.PriceOrder = parsed[11].(float64)
	sp.PriceTrigger = parsed[12].(float64)
	sp.PriceSL = parsed[13].(float64)
	sp.PriceTP = parsed[14].(float64)
	sp.PriceDeviation = parsed[15].(uint32)
	sp.PriceTop = parsed[16].(float64)
	sp.PriceBottom = parsed[17].(float64)

	return sp, nil
}

// ParseYP parses fields using Pp and casts directly
func ParseYP(t []byte, e int) (interface{}, error) {
	parsed, err := ec.Parse(t, Pp, e) // Parse returns []interface{}
	if err != nil {
		return nil, err
	}

	s := &yp{}

	s.Retcode = int(parsed[0].(uint32))
	s.Deal = uint64(parsed[1].(int64))
	s.TradeOrder = uint64(parsed[2].(int64))
	s.Volume = uint64(parsed[3].(int64))
	s.Price = parsed[4].(float64)
	s.ResTime = uint64(parsed[5].(uint32))
	s.Bid = parsed[6].(float64)
	s.Ask = parsed[7].(float64)
	s.Last = parsed[8].(float64)
	s.Comment = parsed[9].(string)

	return s, nil
}

type TradeResult = yp

type TradeEventHandler func(event []*TradeEvent, yp []*TradeResult)

func (c *AppClient) RegisterTradeEventHandler(handler TradeEventHandler, opts ...HandlerOption) func() {
	if handler == nil {
		return func() {}
	}

	return c.RegisterHandler(CommandTradeEvent, func(msg Message) error {
		spList, ypList, err := ParseTradeEvent(msg.Body)
		if err != nil {
			return err
		}
		handler(spList, ypList)

		return nil
	}, opts...)
}
