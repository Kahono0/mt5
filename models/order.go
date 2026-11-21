package models

import "github.com/kahono0/mt5/internal/ec"

type Order struct {
	TradeOrder        uint64
	OwnerID           int64
	OrderID           string
	TradeSymbol       string
	TimeSetup         int64
	TimeExpiration    int64
	TimeDone          int64
	OrderType         int32
	TypeFilling       int32
	TypeTime          int32
	TypeReason        int32
	PriceOrder        float64
	PriceTrigger      float64
	PriceCurrent      float64
	PriceSL           float64
	PriceTP           float64
	VolumeInitial     uint64
	VolumeCurrent     uint64
	OrderState        int32
	Expert            uint64
	PositionID        uint64
	Comment           string
	ContractSize      float64
	Digits            int32
	DigitsCurrency    int32
	CommissionDaily   float64
	CommissionMonthly float64
	MarginRate        float64
	ActivationMode    int32
}

func (o Order) Fields() []ec.Field {
	return OrderProps
}

func (o Order) FromValuesOrder(parsed []interface{}) any {
	// adjust times like in JS version
	parsed[3] = int64(parsed[3].(uint32))*1000 + int64(parsed[28].(int32))
	parsed[4] = int64(parsed[4].(uint32)) * 1000
	parsed[5] = int64(parsed[5].(uint32))*1000 + int64(parsed[29].(int32))

	order := &Order{
		TradeOrder:        uint64(parsed[0].(int64)),
		OrderID:           parsed[1].(string),
		TradeSymbol:       parsed[2].(string),
		TimeSetup:         parsed[3].(int64),
		TimeExpiration:    parsed[4].(int64),
		TimeDone:          parsed[5].(int64),
		OrderType:         int32(parsed[6].(uint32)),
		TypeFilling:       int32(parsed[7].(uint32)),
		TypeTime:          int32(parsed[8].(uint32)),
		TypeReason:        int32(parsed[9].(uint32)),
		PriceOrder:        parsed[10].(float64),
		PriceTrigger:      parsed[11].(float64),
		PriceCurrent:      parsed[12].(float64),
		PriceSL:           parsed[13].(float64),
		PriceTP:           parsed[14].(float64),
		VolumeInitial:     uint64(parsed[15].(int64)),
		VolumeCurrent:     uint64(parsed[16].(int64)),
		OrderState:        int32(parsed[17].(uint32)),
		Expert:            uint64(parsed[18].(int64)),
		PositionID:        uint64(parsed[19].(int64)),
		Comment:           parsed[20].(string),
		ContractSize:      parsed[21].(float64),
		Digits:            int32(parsed[22].(uint32)),
		DigitsCurrency:    int32(parsed[23].(uint32)),
		CommissionDaily:   parsed[24].(float64),
		CommissionMonthly: parsed[25].(float64),
		MarginRate:        parsed[26].(float64),
		ActivationMode:    int32(parsed[27].(uint32)),
	}

	return order
}

var OrderProps = []ec.Field{
	{PropType: 17},
	{PropType: 11, PropLength: historyTextBytes},
	{PropType: 11, PropLength: historyTextBytes},
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 17},
	{PropType: 17},
	{PropType: 6},
	{PropType: 17},
	{PropType: 17},
	{PropType: 11, PropLength: historyTextBytes},
	{PropType: 8},
	{PropType: 6},
	{PropType: 6},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 6},
	{PropType: 3},
	{PropType: 3},
}

var OrderPropsSize, _ = ec.GetSeriesSize(OrderProps)
