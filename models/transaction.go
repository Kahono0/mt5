package models

import (
	"encoding/binary"
	"fmt"

	"github.com/kahono0/mt5/internal/ec"
)

type OrderTransaction struct {
	FlagMask        uint32
	TransactionID   uint32
	TransactionType uint32
	TradeOrder      *Order
}

var OrderTransactionProps = []ec.Field{
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 13, PropLength: OrderPropsSize, PropLengthType: 6, InnerFields: OrderProps, FromValues: Order{}.FromValuesOrder},
}

var OrderTransactionSize, _ = ec.GetSeriesSize(OrderTransactionProps)

func (ot *OrderTransaction) ParseFromBytes(b []byte) error {
	ot.FlagMask = binary.LittleEndian.Uint32(b[0:4])
	ot.TransactionID = binary.LittleEndian.Uint32(b[4:8])
	ot.TransactionType = binary.LittleEndian.Uint32(b[8:12])

	offset := 12

	order := &Order{}
	if err := order.ParseFromBytes(b[offset : offset+OrderPropsSize]); err != nil {
		return err
	}
	ot.TradeOrder = order

	return nil
}

type DealTransaction struct {
	FlagMask          uint32
	TransactionID     uint32
	TransactionType   uint32
	Deal              *Deal
	TradePosition     *Position // Can be *Position or raw parsed data
	Balance           float64
	Credit            float64
	CommissionDaily   float64
	CommissionMonthly float64
	AccProfit         float64
}

var DealTransactionProps = []ec.Field{
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 13, PropLength: DealPropsSize, PropLengthType: 6, InnerFields: DealProps, FromValues: Deal{}.FromValues},
	{PropType: 13, PropLength: PositionPropsSize, PropLengthType: 6, InnerFields: PositionFields, FromValues: Position{}.FromValues},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
}

var DealTransactionSize, _ = ec.GetSeriesSize(DealTransactionProps)

func (dt *DealTransaction) ParseFromBytes(b []byte) error {
	//r := recordReader{buf: b}
	//dt.FlagMask = r.readUint32()
	//dt.TransactionID = r.readUint32()
	//dt.TransactionType = r.readUint32()
	dt.FlagMask = binary.LittleEndian.Uint32(b[0:4])
	dt.TransactionID = binary.LittleEndian.Uint32(b[4:8])
	dt.TransactionType = binary.LittleEndian.Uint32(b[8:12])

	offset := 12

	deal := &Deal{}
	if err := deal.ParseFromBytes(b[offset : offset+DealPropsSize]); err != nil {
		return err
	}
	offset += DealPropsSize
	dt.Deal = deal

	position := &Position{}
	if err := position.ParseFromBytes(b[offset : offset+PositionPropsSize]); err != nil {
		return err
	}
	offset += PositionPropsSize
	dt.TradePosition = position

	r := recordReader{buf: b[offset:]}

	dt.Balance = r.readFloat64()
	dt.Credit = r.readFloat64()
	dt.CommissionDaily = r.readFloat64()
	dt.CommissionMonthly = r.readFloat64()
	dt.AccProfit = r.readFloat64()

	return r.finish()
}

type EnvelopeTransaction struct {
	MsgType           uint32
	OrderTransactions []*OrderTransaction
	DealTransactions  []*DealTransaction
}

func (et *EnvelopeTransaction) ParseFromBytes(b []byte) error {
	if len(b) < 4 {
		return fmt.Errorf("buffer too short to read message type")
	}
	et.MsgType = binary.LittleEndian.Uint32(b[0:4])
	offset := 0

	switch et.MsgType {
	case 0:
		for offset < len(b) {
			if offset+OrderTransactionSize > len(b) {
				// return fmt.Errorf("buffer too short to read OrderTransaction at offset %d", offset)
				return nil // return what we have so far
			}
			orderTx := &OrderTransaction{}
			if err := orderTx.ParseFromBytes(b[offset : offset+OrderTransactionSize]); err != nil {
				return fmt.Errorf("failed to parse OrderTransaction at offset %d: %w", offset, err)
			}
			et.OrderTransactions = append(et.OrderTransactions, orderTx)
			offset += OrderTransactionSize
		}
	case 2:
		for offset < len(b) {
			if offset+DealTransactionSize > len(b) {
				// return fmt.Errorf("buffer too short to read DealTransaction at offset %d", offset)
				return nil // return what we have so far
			}
			dealTx := &DealTransaction{}
			if err := dealTx.ParseFromBytes(b[offset : offset+DealTransactionSize]); err != nil {
				return fmt.Errorf("failed to parse DealTransaction at offset %d: %w", offset, err)
			}
			et.DealTransactions = append(et.DealTransactions, dealTx)
			offset += DealTransactionSize
		}
	default:
		return fmt.Errorf("unexpected message type: %d", et.MsgType)
	}

	return nil
}
