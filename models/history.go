package models

import (
	"encoding/binary"
	"fmt"
	"math"

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
		deals[i], err = decodeDeal(body[offset : offset+DealPropsSize])
		if err != nil {
			return fmt.Errorf("decode deal %d: %w", i, err)
		}
		offset += DealPropsSize
	}

	orderCount, offset, err := readHistoryCount(body, offset, OrderPropsSize, "orders")
	if err != nil {
		return err
	}
	orders := make([]*Order, orderCount)
	for i := range orders {
		orders[i], err = decodeOrder(body[offset : offset+OrderPropsSize])
		if err != nil {
			return fmt.Errorf("decode order %d: %w", i, err)
		}
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

func decodeDeal(buf []byte) (*Deal, error) {
	r := recordReader{buf: buf}
	deal := r.readUint64()
	dealID := r.readString(historyTextBytes)
	tradeOrder := r.readUint64()
	timeCreate := r.readUint32()
	timeUpdate := r.readUint32()
	tradeSymbol := r.readString(historyTextBytes)
	tradeAction := r.readUint32()
	entry := r.readUint32()
	priceOpen := r.readFloat64()
	priceClose := r.readFloat64()
	sl := r.readFloat64()
	tp := r.readFloat64()
	tradeVolume := r.readUint64()
	profit := r.readFloat64()
	rateProfit := r.readFloat64()
	rateMargin := r.readFloat64()
	commission := r.readFloat64()
	storage := r.readFloat64()
	expert := r.readUint64()
	positionID := r.readUint64()
	comment := r.readString(historyTextBytes)
	contractSize := r.readFloat64()
	digits := r.readUint32()
	digitsCurrency := r.readUint32()
	tradeReason := r.readUint32()
	timeCreateSubseconds := r.readInt32()
	timeUpdateSubseconds := r.readInt32()
	commissionFee := r.readFloat64()
	if err := r.finish(); err != nil {
		return nil, err
	}

	return &Deal{
		Deal:           int64(deal),
		DealID:         dealID,
		TradeOrder:     int64(tradeOrder),
		TimeCreate:     1000*timeCreate + uint32(timeCreateSubseconds),
		TimeUpdate:     1000*timeUpdate + uint32(timeUpdateSubseconds),
		TradeSymbol:    tradeSymbol,
		TradeAction:    tradeAction,
		Entry:          entry,
		PriceOpen:      priceOpen,
		PriceClose:     priceClose,
		SL:             sl,
		TP:             tp,
		TradeVolume:    int64(tradeVolume),
		Profit:         profit,
		RateProfit:     rateProfit,
		RateMargin:     rateMargin,
		Commission:     commission,
		Storage:        storage,
		Expert:         int64(expert),
		PositionID:     int64(positionID),
		Comment:        comment,
		ContractSize:   int64(contractSize),
		Digits:         int64(digits),
		DigitsCurrency: int64(digitsCurrency),
		TradeReason:    int64(tradeReason),
		CommissionFee:  commissionFee,
	}, nil
}

func decodeOrder(buf []byte) (*Order, error) {
	r := recordReader{buf: buf}
	tradeOrder := r.readUint64()
	orderID := r.readString(historyTextBytes)
	tradeSymbol := r.readString(historyTextBytes)
	timeSetup := r.readUint32()
	timeExpiration := r.readUint32()
	timeDone := r.readUint32()
	orderType := r.readUint32()
	typeFilling := r.readUint32()
	typeTime := r.readUint32()
	typeReason := r.readUint32()
	priceOrder := r.readFloat64()
	priceTrigger := r.readFloat64()
	priceCurrent := r.readFloat64()
	priceSL := r.readFloat64()
	priceTP := r.readFloat64()
	volumeInitial := r.readUint64()
	volumeCurrent := r.readUint64()
	orderState := r.readUint32()
	expert := r.readUint64()
	positionID := r.readUint64()
	comment := r.readString(historyTextBytes)
	contractSize := r.readFloat64()
	digits := r.readUint32()
	digitsCurrency := r.readUint32()
	commissionDaily := r.readFloat64()
	commissionMonthly := r.readFloat64()
	marginRate := r.readFloat64()
	activationMode := r.readUint32()
	timeSetupSubseconds := r.readInt32()
	timeDoneSubseconds := r.readInt32()
	if err := r.finish(); err != nil {
		return nil, err
	}

	return &Order{
		TradeOrder:        tradeOrder,
		OrderID:           orderID,
		TradeSymbol:       tradeSymbol,
		TimeSetup:         int64(timeSetup)*1000 + int64(timeSetupSubseconds),
		TimeExpiration:    int64(timeExpiration) * 1000,
		TimeDone:          int64(timeDone)*1000 + int64(timeDoneSubseconds),
		OrderType:         int32(orderType),
		TypeFilling:       int32(typeFilling),
		TypeTime:          int32(typeTime),
		TypeReason:        int32(typeReason),
		PriceOrder:        priceOrder,
		PriceTrigger:      priceTrigger,
		PriceCurrent:      priceCurrent,
		PriceSL:           priceSL,
		PriceTP:           priceTP,
		VolumeInitial:     volumeInitial,
		VolumeCurrent:     volumeCurrent,
		OrderState:        int32(orderState),
		Expert:            expert,
		PositionID:        positionID,
		Comment:           comment,
		ContractSize:      contractSize,
		Digits:            int32(digits),
		DigitsCurrency:    int32(digitsCurrency),
		CommissionDaily:   commissionDaily,
		CommissionMonthly: commissionMonthly,
		MarginRate:        marginRate,
		ActivationMode:    int32(activationMode),
	}, nil
}

type recordReader struct {
	buf    []byte
	offset int
	err    error
}

func (r *recordReader) take(size int) []byte {
	if r.err != nil {
		return nil
	}
	if size < 0 || size > len(r.buf)-r.offset {
		r.err = fmt.Errorf("record too short at byte %d: need %d bytes", r.offset, size)
		return nil
	}
	value := r.buf[r.offset : r.offset+size]
	r.offset += size
	return value
}

func (r *recordReader) readUint32() uint32 {
	buf := r.take(4)
	if buf == nil {
		return 0
	}
	return binary.LittleEndian.Uint32(buf)
}

func (r *recordReader) readInt32() int32 { return int32(r.readUint32()) }

func (r *recordReader) readUint64() uint64 {
	buf := r.take(8)
	if buf == nil {
		return 0
	}
	return binary.LittleEndian.Uint64(buf)
}

func (r *recordReader) readFloat64() float64 {
	return math.Float64frombits(r.readUint64())
}

func (r *recordReader) readString(size int) string {
	return ec.DecodeUTF16(r.take(size))
}

func (r *recordReader) finish() error {
	if r.err != nil {
		return r.err
	}
	if r.offset != len(r.buf) {
		return fmt.Errorf("record decoder consumed %d of %d bytes", r.offset, len(r.buf))
	}
	return nil
}

var TradeHistoryFields = []ec.Field{
	{PropType: 13, PropLength: DealPropsSize, PropLengthType: 6, InnerFields: DealProps, FromValues: Deal{}.FromValues},
	{PropType: 13, PropLength: OrderPropsSize, PropLengthType: 6, InnerFields: OrderProps, FromValues: Order{}.FromValuesOrder},
}
