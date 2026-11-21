package models

import (
	"github.com/kahono0/mt5/internal/ec"
)

// Order represents a parsed order (like JS vd)
type Deal struct {
	Deal           int64
	DealID         string
	TradeOrder     int64
	TimeCreate     uint32
	TimeUpdate     uint32
	TradeSymbol    string
	TradeAction    uint32
	Entry          uint32
	PriceOpen      float64
	PriceClose     float64
	SL             float64
	TP             float64
	TradeVolume    int64
	Profit         float64
	RateProfit     float64
	RateMargin     float64
	Commission     float64
	Storage        float64
	Expert         int64
	PositionID     int64
	Comment        string
	ContractSize   int64
	Digits         int64
	DigitsCurrency int64
	TradeReason    int64
	CommissionFee  float64
}

func (d Deal) Fields() []ec.Field {
	return DealProps
}

func (d Deal) FromValues(values []interface{}) any {
	deal := &Deal{
		Deal:           values[0].(int64),
		DealID:         values[1].(string),
		TradeOrder:     values[2].(int64),
		TimeCreate:     1000*values[3].(uint32) + uint32(values[25].(int32)),
		TimeUpdate:     1000*values[4].(uint32) + uint32(values[26].(int32)),
		TradeSymbol:    values[5].(string),
		TradeAction:    values[6].(uint32),
		Entry:          values[7].(uint32),
		PriceOpen:      values[8].(float64),
		PriceClose:     values[9].(float64),
		SL:             values[10].(float64),
		TP:             values[11].(float64),
		TradeVolume:    int64(values[12].(uint64)),
		Profit:         values[13].(float64),
		RateProfit:     values[14].(float64),
		RateMargin:     values[15].(float64),
		Commission:     values[16].(float64),
		Storage:        values[17].(float64),
		Expert:         values[18].(int64),
		PositionID:     values[19].(int64),
		Comment:        values[20].(string),
		ContractSize:   int64(values[21].(float64)),
		Digits:         int64(values[22].(uint32)),
		DigitsCurrency: int64(values[23].(uint32)),
		TradeReason:    int64(values[24].(uint32)),
		CommissionFee:  values[27].(float64),
	}
	return deal
}

var DealProps = []ec.Field{
	{PropType: 17},
	{PropType: 11, PropLength: 64},
	{PropType: 17},
	{PropType: 6},
	{PropType: 6},
	{PropType: 11, PropLength: 64},
	{PropType: 6},
	{PropType: 6},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 18},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 17},
	{PropType: 17},
	{PropType: 11, PropLength: 64},
	{PropType: 8},
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 3},
	{PropType: 3},
	{PropType: 8},
}

var DealPropsSize, _ = ec.GetSeriesSize(DealProps)
