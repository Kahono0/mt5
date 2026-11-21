package models

import (
	"encoding/binary"
	"fmt"

	"github.com/kahono0/mt5/internal/ec"
)

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

const historyTextBytes = 64

var TradeHistoryFields = []ec.Field{
	{PropType: 13, PropLength: DealPropsSize, PropLengthType: 6, InnerFields: DealProps, FromValues: Deal{}.FromValues},
	{PropType: 13, PropLength: OrderPropsSize, PropLengthType: 6, InnerFields: OrderProps, FromValues: Order{}.FromValuesOrder},
}

// ParseFromBytes decodes the two fixed-width record arrays directly into
// typed models, avoiding the generic parser's per-field []interface{} values
// and FromValues conversion pass.
func (t *TradeHistory) ParseFromBytes(body []byte) error {
	dealCount, offset, err := readHistoryCount(body, 0, DealPropsSize, "deals")
	if err != nil {
		return err
	}
	deals := make([]*Deal, dealCount)
	for i := range deals {
		deal := &Deal{}
		if err := deal.ParseFromBytes(body[offset : offset+DealPropsSize]); err != nil {
			return fmt.Errorf("decode deal %d: %w", i, err)
		}
		deals[i] = deal
		offset += DealPropsSize
	}

	orderCount, offset, err := readHistoryCount(body, offset, OrderPropsSize, "orders")
	if err != nil {
		return err
	}
	orders := make([]*Order, orderCount)
	for i := range orders {
		order := &Order{}
		if err := order.ParseFromBytes(body[offset : offset+OrderPropsSize]); err != nil {
			return fmt.Errorf("decode order %d: %w", i, err)
		}
		orders[i] = order
		offset += OrderPropsSize
	}

	t.Deals = deals
	t.Orders = orders
	return nil
}

func readHistoryCount(body []byte, offset, recordSize int, name string) (int, int, error) {
	if recordSize <= 0 {
		return 0, offset, fmt.Errorf("invalid %s record size %d", name, recordSize)
	}
	if offset < 0 || offset > len(body) || len(body)-offset < 4 {
		return 0, offset, fmt.Errorf("buffer too short for %s count", name)
	}
	count := binary.LittleEndian.Uint32(body[offset : offset+4])
	offset += 4
	remainingRecords := (len(body) - offset) / recordSize
	if uint64(count) > uint64(remainingRecords) {
		return 0, offset, fmt.Errorf("buffer too short for %s: count %d exceeds available records %d", name, count, remainingRecords)
	}
	return int(count), offset, nil
}
