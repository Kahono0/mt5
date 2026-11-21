package mt5

import (
	"fmt"

	"github.com/kahono0/mt5/internal/ec"
)

type TradeRequest struct {
	ActionID       uint32
	TradeAction    TradeAction
	TradeSymbol    string
	TradeVolume    uint64
	Digits         uint32
	TradeOrder     int64
	TradeType      uint32
	TypeFilling    uint32
	TypeTime       uint32
	TypeFlags      uint32
	TypeReason     int
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
	TimeExpiration int
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

func (c *AppClient) SendOrder(req *TradeRequest) error {
	fields := req.ToFields()

	body, err := ec.Serialize(fields)
	if err != nil {
		return fmt.Errorf("failed to serialize trade request: %w", err)
	}

	if err := c.core.SendCommand(CommandSubmitTrade, body); err != nil {
		return fmt.Errorf("failed to send trade request: %w", err)
	}

	return nil
}
