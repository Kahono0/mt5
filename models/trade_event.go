package models

import (
	"encoding/binary"
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
	ter.TradeAction = TradeAction(r.readUint32())
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
	tr.Retcode = ReturnCode(r.readUint32())
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
