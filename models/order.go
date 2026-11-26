package models

import (
	"github.com/kahono0/mt5/enums"
	"github.com/kahono0/mt5/internal/ec"
)

type Order struct {
	TradeOrder        int64
	OrderID           string
	TradeSymbol       string
	TimeSetup         uint64
	TimeExpiration    uint64
	TimeDone          uint64
	OrderType         enums.OrderType
	TypeFilling       uint32
	TypeTime          uint32
	TypeReason        uint32
	PriceOrder        float64
	PriceTrigger      float64
	PriceCurrent      float64
	PriceSL           float64
	PriceTP           float64
	VolumeInitial     int64
	VolumeCurrent     int64
	OrderState        uint32
	Expert            int64
	PositionID        int64
	Comment           string
	ContractSize      float64
	Digits            uint32
	DigitsCurrency    uint32
	CommissionDaily   float64
	CommissionMonthly float64
	MarginRate        float64
	ActivationMode    uint32

	//TimeSubSecondsSetup int32
	//TimeSubSecondsDone  int32
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

func (o Order) Fields() []ec.Field {
	return OrderProps
}

func (o Order) FromValuesOrder(parsed []interface{}) any {
	parsed[3] = uint64(parsed[3].(uint32))*1000 + uint64(parsed[28].(int32))
	parsed[4] = uint64(parsed[4].(uint32)) * 1000
	parsed[5] = uint64(parsed[5].(uint32))*1000 + uint64(parsed[29].(int32))

	order := &Order{
		TradeOrder:        parsed[0].(int64),
		OrderID:           parsed[1].(string),
		TradeSymbol:       parsed[2].(string),
		TimeSetup:         parsed[3].(uint64),
		TimeExpiration:    parsed[4].(uint64),
		TimeDone:          parsed[5].(uint64),
		OrderType:         enums.OrderType(parsed[6].(uint32)),
		TypeFilling:       parsed[7].(uint32),
		TypeTime:          parsed[8].(uint32),
		TypeReason:        parsed[9].(uint32),
		PriceOrder:        parsed[10].(float64),
		PriceTrigger:      parsed[11].(float64),
		PriceCurrent:      parsed[12].(float64),
		PriceSL:           parsed[13].(float64),
		PriceTP:           parsed[14].(float64),
		VolumeInitial:     parsed[15].(int64),
		VolumeCurrent:     parsed[16].(int64),
		OrderState:        parsed[17].(uint32),
		Expert:            parsed[18].(int64),
		PositionID:        parsed[19].(int64),
		Comment:           parsed[20].(string),
		ContractSize:      parsed[21].(float64),
		Digits:            parsed[22].(uint32),
		DigitsCurrency:    parsed[23].(uint32),
		CommissionDaily:   parsed[24].(float64),
		CommissionMonthly: parsed[25].(float64),
		MarginRate:        parsed[26].(float64),
		ActivationMode:    parsed[27].(uint32),
	}

	return order
}

var OrderPropsSize, _ = ec.GetSeriesSize(OrderProps)

func (o *Order) ParseFromBytes(buf []byte) error {
	r := recordReader{buf: buf}
	o.TradeOrder = r.readInt64()
	o.OrderID = r.readString(historyTextBytes)
	o.TradeSymbol = r.readString(historyTextBytes)
	timeSetup := r.readUint32()
	timeExpiration := r.readUint32()
	timeDone := r.readUint32()
	o.OrderType = enums.OrderType(r.readUint32())
	o.TypeFilling = r.readUint32()
	o.TypeTime = r.readUint32()
	o.TypeReason = r.readUint32()
	o.PriceOrder = r.readFloat64()
	o.PriceTrigger = r.readFloat64()
	o.PriceCurrent = r.readFloat64()
	o.PriceSL = r.readFloat64()
	o.PriceTP = r.readFloat64()
	o.VolumeInitial = r.readInt64()
	o.VolumeCurrent = r.readInt64()
	o.OrderState = r.readUint32()
	o.Expert = r.readInt64()
	o.PositionID = r.readInt64()
	o.Comment = r.readString(historyTextBytes)
	o.ContractSize = r.readFloat64()
	o.Digits = r.readUint32()
	o.DigitsCurrency = r.readUint32()
	o.CommissionDaily = r.readFloat64()
	o.CommissionMonthly = r.readFloat64()
	o.MarginRate = r.readFloat64()
	o.ActivationMode = r.readUint32()
	timeSetupSubseconds := r.readInt32()
	timeDoneSubseconds := r.readInt32()

	if err := r.finish(); err != nil {
		return err
	}

	o.TimeSetup = uint64(timeSetup)*1000 + uint64(timeSetupSubseconds)
	o.TimeExpiration = uint64(timeExpiration) * 1000
	o.TimeDone = uint64(timeDone)*1000 + uint64(timeDoneSubseconds)

	return nil
}
