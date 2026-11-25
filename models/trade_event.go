package models

import (
	"encoding/binary"
	"fmt"

	"github.com/kahono0/mt5/enums"
	"github.com/kahono0/mt5/internal/ec"
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
	TradeAction    enums.TradeAction
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

var TradeEventRequestProps = []ec.Field{
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

var TradeEventRequestSize, _ = ec.GetSeriesSize(TradeEventRequestProps)

func (ter *TradeEventRequest) ParseFromBytes(b []byte) error {
	r := recordReader{buf: b}
	ter.ActionID = r.readUint32()
	ter.TradeAction = enums.TradeAction(r.readUint32())
	ter.TradeSymbol = r.readString(64)
	ter.TradeVolume = r.readUint64()
	ter.Digits = r.readUint32()
	ter.TradeOrder = int64(r.readUint64())
	ter.TradeType = r.readUint32()
	ter.TypeFilling = r.readUint32()
	_ = r.readUint32() // TypeTime, not used
	ter.TypeFlags = r.readUint32()
	_ = r.readInt32() // TypeReason, not used
	ter.PriceOrder = r.readFloat64()
	ter.PriceTrigger = r.readFloat64()
	ter.PriceSL = r.readFloat64()
	ter.PriceTP = r.readFloat64()
	ter.PriceDeviation = r.readUint32()
	ter.PriceTop = r.readFloat64()
	ter.PriceBottom = r.readFloat64()
	ter.Comment = r.readString(64)
	ter.PositionBy = r.readUint64()
	_ = r.readUint64() // Position, not used
	_ = r.readInt32()  // TimeExpiration, not used

	return r.finish()
}

type TradeResult struct {
	Retcode    enums.TradeReturnCode
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

var TradeResultProps = []ec.Field{
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

var TradeResultSize, _ = ec.GetSeriesSize(TradeResultProps)

func (tr *TradeResult) ParseFromBytes(b []byte) error {
	r := recordReader{buf: b}
	tr.Retcode = enums.TradeReturnCode(r.readUint32())
	tr.Deal = int64(r.readUint64())
	tr.TradeOrder = int64(r.readUint64())
	tr.Volume = int64(r.readUint64())
	tr.Price = r.readFloat64()
	tr.ResTime = r.readUint32()
	tr.Bid = r.readFloat64()
	tr.Ask = r.readFloat64()
	tr.Last = r.readFloat64()
	tr.Comment = r.readString(64)

	return r.finish()
}

type TradeEvent struct {
	EventID      uint32
	TradeRequest *TradeEventRequest
	TradeResult  *TradeResult
}

func (te *TradeEvent) ParseFromBytes(b []byte) error {
	if len(b) < 4 {
		return fmt.Errorf("buffer too short to read EventID")
	}
	te.EventID = binary.LittleEndian.Uint32(b[:4])
	offset := 4

	te.TradeRequest = &TradeEventRequest{}
	if err := te.TradeRequest.ParseFromBytes(b[offset : offset+TradeEventRequestSize]); err != nil {
		return fmt.Errorf("failed to parse TradeEventRequest: %w", err)
	}

	offset += TradeEventRequestSize

	te.TradeResult = &TradeResult{}
	if err := te.TradeResult.ParseFromBytes(b[offset : offset+TradeResultSize]); err != nil {
		return fmt.Errorf("failed to parse TradeResult: %w", err)
	}

	return nil
}
