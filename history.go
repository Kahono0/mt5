package mt5

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/kahono0/mt5/internal/ec"
)

type TradeHistory struct {
	Deals  []*Deal
	Orders []*Order
}

var DealProps = []ec.Field{
	{PropType: 17},
	{PropType: 11, PropLength: 64},
	{PropType: 17},
	{PropType: 6},
	{PropType: 6},
	{PropType: 11, PropLength: 64},
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
	{PropType: 11, PropLength: 64},
	{PropType: 8},
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 3},
	{PropType: 3},
	{PropType: 8},
}

var DealPropsSize, _ = ec.GetSeriesSize(DealProps)

var OrderProps = Id
var OrderPropsSize, _ = ec.GetSeriesSize(OrderProps)

var TradeHistoryFields = []ec.Field{
	{PropType: 13, PropLength: DealPropsSize, PropLengthType: 6, Parser: parseDeal},
	{PropType: 13, PropLength: OrderPropsSize, PropLengthType: 6, Parser: parseHOrder},
}

func parseDeal(buf []byte, offset int) (interface{}, error) {
	values, err := ec.Parse(buf, DealProps, offset)
	if err != nil {
		return nil, fmt.Errorf("parseDeal: %w", err)
	}

	d := &Deal{
		Deal:           values[0].(int64),
		DealID:         values[1].(string),
		TradeOrder:     values[2].(int64),
		TimeCreate:     1000*values[3].(uint32) + uint32(values[25].(int32)),
		TimeUpdate:     1000*values[4].(uint32) + uint32(values[26].(int32)),
		TradeSymbol:    values[5].(string),
		TradeAction:    values[6].(uint32),
		Entry:          values[7].(uint32),
		PriceOpen:      values[8].(float64),
		PriceClose:     values[9].(float64),
		SL:             values[10].(float64),
		TP:             values[11].(float64),
		TradeVolume:    int64(values[12].(uint64)),
		Profit:         values[13].(float64),
		RateProfit:     values[14].(float64),
		RateMargin:     values[15].(float64),
		Commission:     values[16].(float64),
		Storage:        values[17].(float64),
		Expert:         values[18].(int64),
		PositionID:     values[19].(int64),
		Comment:        values[20].(string),
		ContractSize:   int64(values[21].(float64)),
		Digits:         int64(values[22].(uint32)),
		DigitsCurrency: int64(values[23].(uint32)),
		TradeReason:    int64(values[24].(uint32)),
		CommissionFee:  values[27].(float64),
	}
	return d, nil
}

func parseHOrder(buf []byte, offset int) (interface{}, error) {
	parsed, err := ec.Parse(buf, OrderProps, offset)
	if err != nil {
		return nil, err
	}

	// adjust times like in JS version
	parsed[3] = int64(parsed[3].(uint32))*1000 + int64(parsed[28].(int32))
	parsed[4] = int64(parsed[4].(uint32)) * 1000
	parsed[5] = int64(parsed[5].(uint32))*1000 + int64(parsed[29].(int32))

	s := &Order{
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

	return s, nil

}

func parseDeals(buf []byte, offset int) (interface{}, int, error) {
	log.Printf("Parsing deals at offset %d : len(buf)=%d", offset, len(buf))
	count := binary.LittleEndian.Uint32(buf[offset : offset+4])
	log.Printf("Parsing %d deals at offset %d", count, offset)
	offset += 4

	deals := make([]*Deal, count)
	for i := range count {
		log.Printf("Parsing deal %d at offset %d", i, offset)
		deal, err := parseDeal(buf, offset)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse deal at index %d: %w", i, err)
		}
		deals[i] = deal.(*Deal)
		offset += DealPropsSize
	}

	return deals, offset, nil
}

func parseOrders(buf []byte, offset int) (interface{}, int, error) {
	log.Printf("Parsing orders at offset %d", offset)

	count := binary.LittleEndian.Uint32(buf[offset : offset+4])
	offset += 4

	orders := make([]*Order, count)

	for i := range count {
		order, err := parseHOrder(buf, offset)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse order at index %d: %w", i, err)
		}
		orders[i] = order.(*Order)
		offset += OrderPropsSize
	}

	return orders, offset, nil
}

func parseTradeHistory(body []byte) (*TradeHistory, error) {
	values, err := ec.Parse(body, TradeHistoryFields, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trade history: %w", err)
	}

	// convert []interface{} to []*Deal and []*Order
	deals := make([]*Deal, len(values[0].([]interface{})))
	for i, v := range values[0].([]interface{}) {
		deals[i] = v.(*Deal)
	}

	orders := make([]*Order, len(values[1].([]interface{})))
	for i, v := range values[1].([]interface{}) {
		orders[i] = v.(*Order)
	}

	history := &TradeHistory{
		Deals:  deals,
		Orders: orders,
	}

	return history, nil
}

type TradeHistoryHandler func(history *TradeHistory)

func (c *AppClient) RegisterTradeHistoryHandler(handler TradeHistoryHandler, opts ...HandlerOption) func() {
	if handler == nil {
		return func() {}
	}

	return c.RegisterHandler(CommandLoadHistory, func(msg Message) error {
		history, err := parseTradeHistory(msg.Body)
		if err != nil {
			return fmt.Errorf("failed to parse trade history: %w", err)
		}

		handler(history)
		return nil
	}, opts...)
}

// RequestTradeHistory sends a request for trade history between the
// given timestamps. The client must be authenticated.
func (c *AppClient) RequestTradeHistory(from, to time.Time) error {
	if !c.IsAuthenticated() {
		return ErrNotAuthenticated
	}
	body := buildTradeHistoryPayload(from, to)
	if err := c.core.SendCommand(CommandLoadHistory, body); err != nil {
		return fmt.Errorf("failed to request trade history: %w", err)
	}
	return nil
}

func (c *AppClient) GetTradeHistory(ctx context.Context, from, to time.Time) (*TradeHistory, error) {
	if !c.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}

	type result struct {
		history *TradeHistory
		err     error
	}

	ch := make(chan result, 1)

	cancel := c.RegisterTradeHistoryHandler(func(history *TradeHistory) {
		ch <- result{history: history, err: nil}
	}, WithErrorHandler(func(err error) {
		ch <- result{history: nil, err: err}
	}))
	defer cancel()

	if err := c.RequestTradeHistory(from, to); err != nil {
		return nil, fmt.Errorf("failed to request trade history: %w", err)
	}

	select {
	case res := <-ch:
		return res.history, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// BuildTradeHistoryPayload encodes a from/to timestamp pair into the
// binary payload format expected by the trade history command.
func buildTradeHistoryPayload(from, to time.Time) []byte {
	body := make([]byte, 8)
	binary.LittleEndian.PutUint32(body[0:4], uint32(from.Unix()))
	binary.LittleEndian.PutUint32(body[4:8], uint32(to.Unix()))
	return body
}
