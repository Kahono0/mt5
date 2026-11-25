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

var OrderPropsSize, _ = ec.GetSeriesSize(OrderProps)

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

func (o *Order) ParseFromBytes(buf []byte) error {
	r := recordReader{buf: buf}
	o.TradeOrder = r.readUint64()
	o.OrderID = r.readString(historyTextBytes)
	o.TradeSymbol = r.readString(historyTextBytes)
	timeSetup := r.readUint32()
	timeExpiration := r.readUint32()
	timeDone := r.readUint32()
	o.OrderType = int32(r.readUint32())
	o.TypeFilling = int32(r.readUint32())
	o.TypeTime = int32(r.readUint32())
	o.TypeReason = int32(r.readUint32())
	o.PriceOrder = r.readFloat64()
	o.PriceTrigger = r.readFloat64()
	o.PriceCurrent = r.readFloat64()
	o.PriceSL = r.readFloat64()
	o.PriceTP = r.readFloat64()
	o.VolumeInitial = r.readUint64()
	o.VolumeCurrent = r.readUint64()
	o.OrderState = int32(r.readUint32())
	o.Expert = r.readUint64()
	o.PositionID = r.readUint64()
	o.Comment = r.readString(historyTextBytes)
	o.ContractSize = r.readFloat64()
	o.Digits = int32(r.readUint32())
	o.DigitsCurrency = int32(r.readUint32())
	o.CommissionDaily = r.readFloat64()
	o.CommissionMonthly = r.readFloat64()
	o.MarginRate = r.readFloat64()
	o.ActivationMode = int32(r.readUint32())
	timeSetupSubseconds := r.readInt32()
	timeDoneSubseconds := r.readInt32()

	if err := r.finish(); err != nil {
		return err
	}

	o.TimeSetup = int64(timeSetup)*1000 + int64(timeSetupSubseconds)
	o.TimeExpiration = int64(timeExpiration) * 1000
	o.TimeDone = int64(timeDone)*1000 + int64(timeDoneSubseconds)

	return nil
}
