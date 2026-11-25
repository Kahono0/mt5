package models

import (
	"github.com/kahono0/mt5/enums"
	"github.com/kahono0/mt5/internal/ec"
)

type TradeRequest struct {
	ActionID       uint32
	TradeAction    enums.TradeAction
	TradeSymbol    string
	TradeVolume    uint64
	Digits         uint32
	TradeOrder     int64
	TradeType      uint32
	TypeFilling    uint32
	TypeTime       uint32
	TypeFlags      uint32
	TypeReason     uint32
	PriceOrder     float64
	PriceTrigger   float64
	PriceSL        float64
	PriceTP        float64
	PriceDeviation uint32
	PriceTop       float64
	PriceBottom    float64
	Comment        string
	TradePosition  uint64
	PositionBy     uint64
	TimeExpiration uint32
}

// ToFields converts a TradeRequest into a slice of Fields
func (t *TradeRequest) ToFields() []ec.Field {
	return []ec.Field{
		{PropType: 6, PropValue: t.ActionID},
		{PropType: 6, PropValue: uint32(t.TradeAction)},
		{PropType: 11, PropValue: t.TradeSymbol[:min(32, len(t.TradeSymbol))], PropLength: 64},
		{PropType: 18, PropValue: t.TradeVolume},
		{PropType: 6, PropValue: t.Digits},
		{PropType: 18, PropValue: t.TradeOrder},
		{PropType: 6, PropValue: t.TradeType},
		{PropType: 6, PropValue: t.TypeFilling},
		{PropType: 6, PropValue: t.TypeTime},
		{PropType: 6, PropValue: t.TypeFlags},
		{PropType: 6, PropValue: t.TypeReason},
		{PropType: 8, PropValue: t.PriceOrder},
		{PropType: 8, PropValue: t.PriceTrigger},
		{PropType: 8, PropValue: t.PriceSL},
		{PropType: 8, PropValue: t.PriceTP},
		{PropType: 6, PropValue: t.PriceDeviation},
		{PropType: 8, PropValue: t.PriceTop},
		{PropType: 8, PropValue: t.PriceBottom},
		{PropType: 11, PropValue: t.Comment[:min(32, len(t.Comment))], PropLength: 64},
		{PropType: 18, PropValue: t.TradePosition},
		{PropType: 18, PropValue: t.PositionBy},
		{PropType: 6, PropValue: t.TimeExpiration},
	}
}

var TradeRequestProps = []ec.Field{
	{PropType: 6},
	{PropType: 6},
	{PropType: 11, PropLength: 64},
	{PropType: 18},
	{PropType: 6},
	{PropType: 18},
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
	{PropType: 6},
}

var TradeRequestSize, _ = ec.GetSeriesSize(TradeRequestProps)

func (tr *TradeRequest) Serialize() ([]byte, error) {
	w := newRecordWriter(TradeRequestSize)
	w.writeUint32(tr.ActionID)
	w.writeUint32(uint32(tr.TradeAction))
	w.writeUTF16(tr.TradeSymbol, 64)
	w.writeUint64(tr.TradeVolume)
	w.writeUint32(tr.Digits)
	w.writeInt64(tr.TradeOrder)
	w.writeUint32(tr.TradeType)
	w.writeUint32(tr.TypeFilling)
	w.writeUint32(tr.TypeTime)
	w.writeUint32(tr.TypeFlags)
	w.writeInt32(int32(tr.TypeReason))
	w.writeFloat64(tr.PriceOrder)
	w.writeFloat64(tr.PriceTrigger)
	w.writeFloat64(tr.PriceSL)
	w.writeFloat64(tr.PriceTP)
	w.writeUint32(tr.PriceDeviation)
	w.writeFloat64(tr.PriceTop)
	w.writeFloat64(tr.PriceBottom)
	w.writeUTF16(tr.Comment, 64)
	w.writeUint64(tr.TradePosition)
	w.writeUint64(tr.PositionBy)
	w.writeInt32(int32(tr.TimeExpiration))

	return w.finish()
}
