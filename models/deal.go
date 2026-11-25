package models

import (
	"github.com/kahono0/mt5/internal/ec"
)

// Order represents a parsed order (like JS vd)
type Deal struct {
	Deal           int64
	DealID         string
	TradeOrder     int64
	TimeCreate     uint64
	TimeUpdate     uint64
	TradeSymbol    string
	TradeAction    uint32
	Entry          uint32
	PriceOpen      float64
	PriceClose     float64
	SL             float64
	TP             float64
	TradeVolume    uint64
	Profit         float64
	RateProfit     float64
	RateMargin     float64
	Commission     float64
	Storage        float64
	Expert         int64
	PositionID     int64
	Comment        string
	ContractSize   float64
	Digits         uint32
	DigitsCurrency uint32
	TradeReason    uint32
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
		TimeCreate:     uint64(values[3].(uint32))*1000 + uint64(values[25].(int32)),
		TimeUpdate:     uint64(values[4].(uint32))*1000 + uint64(values[26].(int32)),
		TradeSymbol:    values[5].(string),
		TradeAction:    values[6].(uint32),
		Entry:          values[7].(uint32),
		PriceOpen:      values[8].(float64),
		PriceClose:     values[9].(float64),
		SL:             values[10].(float64),
		TP:             values[11].(float64),
		TradeVolume:    values[12].(uint64),
		Profit:         values[13].(float64),
		RateProfit:     values[14].(float64),
		RateMargin:     values[15].(float64),
		Commission:     values[16].(float64),
		Storage:        values[17].(float64),
		Expert:         values[18].(int64),
		PositionID:     values[19].(int64),
		Comment:        values[20].(string),
		ContractSize:   values[21].(float64),
		Digits:         values[22].(uint32),
		DigitsCurrency: values[23].(uint32),
		TradeReason:    values[24].(uint32),
		CommissionFee:  values[27].(float64),
	}
	return deal
}

var DealPropsSize, _ = ec.GetSeriesSize(DealProps)

var DealProps = []ec.Field{
	{PropType: 17},
	{PropType: 11, PropLength: historyTextBytes},
	{PropType: 17},
	{PropType: 6},
	{PropType: 6},
	{PropType: 11, PropLength: historyTextBytes},
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
	{PropType: 11, PropLength: historyTextBytes},
	{PropType: 8},
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 3},
	{PropType: 3},
	{PropType: 8},
}

func (d *Deal) ParseFromBytes(buf []byte) error {
	r := recordReader{buf: buf}
	d.Deal = int64(r.readUint64())
	d.DealID = r.readString(historyTextBytes)
	d.TradeOrder = r.readInt64()
	timeCreate := r.readUint32()
	timeUpdate := r.readUint32()
	d.TradeSymbol = r.readString(historyTextBytes)
	d.TradeAction = r.readUint32()
	d.Entry = r.readUint32()
	d.PriceOpen = r.readFloat64()
	d.PriceClose = r.readFloat64()
	d.SL = r.readFloat64()
	d.TP = r.readFloat64()
	d.TradeVolume = r.readUint64()
	d.Profit = r.readFloat64()
	d.RateProfit = r.readFloat64()
	d.RateMargin = r.readFloat64()
	d.Commission = r.readFloat64()
	d.Storage = r.readFloat64()
	d.Expert = r.readInt64()
	d.PositionID = r.readInt64()
	d.Comment = r.readString(historyTextBytes)
	d.ContractSize = r.readFloat64()
	d.Digits = r.readUint32()
	d.DigitsCurrency = r.readUint32()
	d.TradeReason = r.readUint32()
	timeCreateSubseconds := r.readInt32()
	timeUpdateSubseconds := r.readInt32()
	d.CommissionFee = r.readFloat64()

	if err := r.finish(); err != nil {
		return err
	}

	//d.TimeCreate = 1000*timeCreate + uint32(timeCreateSubseconds)
	//d.TimeUpdate = 1000*timeUpdate + uint32(timeUpdateSubseconds)
	d.TimeCreate = uint64(timeCreate)*1000 + uint64(timeCreateSubseconds)
	d.TimeUpdate = uint64(timeUpdate)*1000 + uint64(timeUpdateSubseconds)

	return nil
}
