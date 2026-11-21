package models

import "github.com/kahono0/mt5/internal/ec"

type TradeHistory struct {
	Deals  []*Deal
	Orders []*Order
}

func (t TradeHistory) Fields() []ec.Field {
	return TradeHistoryFields
}

func (t TradeHistory) FromValues(values []interface{}) any {
	// convert []interface{} to []*Deal and []*Order
	deals := make([]*Deal, len(values[0].([]interface{})))
	for i, v := range values[0].([]interface{}) {
		deals[i] = v.(*Deal)
	}

	orders := make([]*Order, len(values[1].([]interface{})))
	for i, v := range values[1].([]interface{}) {
		orders[i] = v.(*Order)
	}

	return &TradeHistory{
		Deals:  deals,
		Orders: orders,
	}
}

var TradeHistoryFields = []ec.Field{
	{PropType: 13, PropLength: DealPropsSize, PropLengthType: 6, InnerFields: DealProps, FromValues: Deal{}.FromValues},
	{PropType: 13, PropLength: OrderPropsSize, PropLengthType: 6, InnerFields: OrderProps, FromValues: Order{}.FromValuesOrder},
}
