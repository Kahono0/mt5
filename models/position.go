package models

import (
	"encoding/binary"
	"fmt"

	"github.com/kahono0/mt5/internal/ec"
)

type Position struct {
	PositionID       int64
	TradeOrder       int64
	TimeCreate       uint64
	TimeUpdate       uint64
	TradeSymbol      string
	TradeAction      uint32
	PriceOpen        float64
	PriceClose       float64
	SL               float64
	TP               float64
	TradeVolume      uint64
	Profit           float64
	RateProfit       float64
	RateMargin       float64
	Commission       float64
	Storage          float64
	Expert           int64
	ExpertPositionID int64
	Comment          string
	ContractSize     float64
	Digits           uint32
	DigitsCurrency   uint32
	TradeReason      uint32
	ExternalID       string

	//TimeSubSecondsC  int32 // i[24] in JS
	//TimeSubSecondsU  int32 // i[25] in JS
}

var PositionFields = []ec.Field{
	{PropType: 17},                 // 0
	{PropType: 17},                 // 1
	{PropType: 6},                  // 2
	{PropType: 6},                  // 3
	{PropType: 11, PropLength: 64}, // 4
	{PropType: 6},                  // 5
	{PropType: 8},                  // 6
	{PropType: 8},                  // 7
	{PropType: 8},                  // 8
	{PropType: 8},                  // 9
	{PropType: 18},                 // 10
	{PropType: 8},                  // 11
	{PropType: 8},                  // 12
	{PropType: 8},                  // 13
	{PropType: 8},                  // 14
	{PropType: 8},                  // 15
	{PropType: 17},                 // 16
	{PropType: 17},                 // 17
	{PropType: 11, PropLength: 64}, // 18
	{PropType: 8},                  // 19
	{PropType: 6},                  // 20
	{PropType: 6},                  // 21
	{PropType: 6},                  // 22
	{PropType: 11, PropLength: 64}, // 23
	{PropType: 3},                  // 24
	{PropType: 3},                  // 25
}

var PositionPropsSize, _ = ec.GetSeriesSize(PositionFields)

func (p Position) Fields() []ec.Field {
	return PositionFields
}

func (p Position) FromValues(values []interface{}) interface{} {
	return &Position{
		PositionID:       values[0].(int64),
		TradeOrder:       values[1].(int64),
		TimeCreate:       uint64(values[2].(uint32))*1000 + uint64(values[24].(int32)),
		TimeUpdate:       uint64(values[3].(uint32))*1000 + uint64(values[25].(int32)),
		TradeSymbol:      values[4].(string),
		TradeAction:      values[5].(uint32),
		PriceOpen:        values[6].(float64),
		PriceClose:       values[7].(float64),
		SL:               values[8].(float64),
		TP:               values[9].(float64),
		TradeVolume:      values[10].(uint64),
		Profit:           values[11].(float64),
		RateProfit:       values[12].(float64),
		RateMargin:       values[13].(float64),
		Commission:       values[14].(float64),
		Storage:          values[15].(float64),
		Expert:           values[16].(int64),
		ExpertPositionID: values[17].(int64),
		Comment:          values[18].(string),
		ContractSize:     values[19].(float64),
		Digits:           values[20].(uint32),
		DigitsCurrency:   values[21].(uint32),
		TradeReason:      values[22].(uint32),
		ExternalID:       values[23].(string),
	}
}

func (p *Position) ParseFromBytes(buf []byte) error {
	r := recordReader{buf: buf}

	p.PositionID = r.readInt64()
	p.TradeOrder = r.readInt64()
	timeCreate := r.readUint32()
	timeUpdate := r.readUint32()
	p.TradeSymbol = r.readString(64)
	p.TradeAction = r.readUint32()
	p.PriceOpen = r.readFloat64()
	p.PriceClose = r.readFloat64()
	p.SL = r.readFloat64()
	p.TP = r.readFloat64()
	p.TradeVolume = r.readUint64()
	p.Profit = r.readFloat64()
	p.RateProfit = r.readFloat64()
	p.RateMargin = r.readFloat64()
	p.Commission = r.readFloat64()
	p.Storage = r.readFloat64()
	p.Expert = r.readInt64()
	p.ExpertPositionID = r.readInt64()
	p.Comment = r.readString(64)
	p.ContractSize = r.readFloat64()
	p.Digits = r.readUint32()
	p.DigitsCurrency = r.readUint32()
	p.TradeReason = r.readUint32()
	p.ExternalID = r.readString(64)
	timeSubSecondsC := r.readInt32()
	timeSubSecondsU := r.readInt32()

	if err := r.finish(); err != nil {
		return err
	}

	p.TimeCreate = uint64(timeCreate)*1000 + uint64(timeSubSecondsC)
	p.TimeUpdate = uint64(timeUpdate)*1000 + uint64(timeSubSecondsU)

	return nil
}

func ParsePositionsFromBytes(buf []byte) ([]*Position, error) {
	if len(buf) < 4 {
		return nil, fmt.Errorf("buffer too short to read positions count")
	}

	count := binary.LittleEndian.Uint32(buf[0:4])
	offset := 4

	positions := make([]*Position, count)

	for i := uint32(0); i < count; i++ {
		if offset+PositionPropsSize > len(buf) {
			return nil, fmt.Errorf("buffer too short to read position at index %d", i)
		}

		position := &Position{}
		if err := position.ParseFromBytes(buf[offset : offset+PositionPropsSize]); err != nil {
			return nil, fmt.Errorf("failed to parse position at index %d: %v", i, err)
		}

		positions[i] = position
		offset += PositionPropsSize
	}

	return positions, nil
}
