package mt5

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/kahono0/mt5/internal/ec"
)

// structs

type OrderPosition struct {
	FlagMask          uint32
	TransactionID     uint32
	TransactionType   uint32
	Deal              interface{} // Can be *Order or raw parsed data
	TradePosition     *Position   // Can be *Position or raw parsed data
	TradeOrder        interface{}
	Balance           float64
	Credit            float64
	CommissionDaily   float64
	CommissionMonthly float64
	AccProfit         float64
	Deals             []interface{}
	Positions         []interface{}
	IsOrder           bool
}

type Order struct {
	TradeOrder        uint64
	OwnerID           int64
	OrderID           string
	TradeSymbol       string
	TimeSetup         int64
	TimeExpiration    int64
	TimeDone          int64
	OrderType         int32
	TypeFilling       int32
	TypeTime          int32
	TypeReason        int32
	PriceOrder        float64
	PriceTrigger      float64
	PriceCurrent      float64
	PriceSL           float64
	PriceTP           float64
	VolumeInitial     uint64
	VolumeCurrent     uint64
	OrderState        int32
	Expert            uint64
	PositionID        uint64
	Comment           string
	ContractSize      float64
	Digits            int32
	DigitsCurrency    int32
	CommissionDaily   float64
	CommissionMonthly float64
	MarginRate        float64
	ActivationMode    int32
}

// Order represents a parsed order (like JS vd)
type Deal struct {
	Deal           int64
	DealID         string
	TradeOrder     int64
	TimeCreate     uint32
	TimeUpdate     uint32
	TradeSymbol    string
	TradeAction    uint32
	Entry          uint32
	PriceOpen      float64
	PriceClose     float64
	SL             float64
	TP             float64
	TradeVolume    int64
	Profit         float64
	RateProfit     float64
	RateMargin     float64
	Commission     float64
	Storage        float64
	Expert         int64
	PositionID     int64
	Comment        string
	ContractSize   int64
	Digits         int64
	DigitsCurrency int64
	TradeReason    int64
	CommissionFee  float64
}

type Ju struct {
	FlagMask        int
	TransactionID   int64
	TransactionType int
	TradeOrder      *Order
}

// fields

// OrderPositionProps corresponds to JS orderPositionProps
var OrderPositionProps = []ec.Field{
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{
		PropType:   12,
		PropLength: 0,
		Parser:     ParseNewOrder,
	}, // Pd length will be set after NewOrderProps
	{
		PropType:   12,
		PropLength: 0,
		Parser:     ParseNewPosition,
	}, // gu length will be set after NewPositionProps
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
}

// NewOrderProps corresponds to JS newOrderProps
var NewOrderProps = []ec.Field{
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

// NewPositionProps corresponds to JS newPositionProps
var NewPositionProps = []ec.Field{
	{PropType: 17},
	{PropType: 17},
	{PropType: 6},
	{PropType: 6},
	{PropType: 11, PropLength: 64},
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
	{PropType: 11, PropLength: 64},
	{PropType: 3},
	{PropType: 3},
}

var Zu = []ec.Field{
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{
		PropType:   12,
		PropLength: Od,
		Parser:     Xd,
	},
}

// Id defines the fields for parsing
var Id = []ec.Field{
	{PropType: 17},
	{PropType: 11, PropLength: 64},
	{PropType: 11, PropLength: 64},
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 17},
	{PropType: 17},
	{PropType: 6},
	{PropType: 17},
	{PropType: 17},
	{PropType: 11, PropLength: 64},
	{PropType: 8},
	{PropType: 6},
	{PropType: 6},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 6},
	{PropType: 3},
	{PropType: 3},
}

// vars

// Calculate sizes
var (
	Pd, _ = ec.GetSeriesSize(NewOrderProps)
	Gu, _ = ec.GetSeriesSize(OrderPositionProps)
	gu, _ = ec.GetSeriesSize(NewPositionProps)
	Qu, _ = ec.GetSeriesSize(Zu)
	Od, _ = ec.GetSeriesSize(Id)
)

var (
	ErrBufferTooShort = errors.New("buffer too short")
)

func init() {
	// Now that Pd and gu are known, set lengths for PropType 12 fields
	OrderPositionProps[3].PropLength = Pd
	OrderPositionProps[4].PropLength = gu
}

type OrderPositionHandler func(event []*OrderPosition)

func (c *AppClient) RegisterOrderPositionHandler(handler OrderPositionHandler, opts ...HandlerOption) func() {
	if handler == nil {
		return func() {}
	}
	return c.RegisterHandler(CommandOrderPosition, func(m Message) error {
		events, err := ParseOrderPositionEvents(m.Body)
		if err != nil {
			return fmt.Errorf("failed to parse order/position events: %w", err)
		}
		handler(events)
		return nil
	}, opts...)
}

func ParseOrderPositionEvents(buf []byte) ([]*OrderPosition, error) {
	var events []*OrderPosition

	i := 0
	bufLen := len(buf)

	// Read the initial uint32 type (should be 2)
	if bufLen < 4 {
		return nil, errors.New("buffer too short to read type")
	}
	msgType := binary.LittleEndian.Uint32(buf[i:])
	if msgType == 0 {
		//parsed, err := ParseTp(buf, 0)
		//if err != nil {
		//	return nil, fmt.Errorf("failed to parse heartbeat: %w", err)
		//}
		//fmt.Printf("Heartbeat: %s\n", utils.AsPrettyJson(parsed))

		for i < bufLen {
			parsed, err := ParseTp(buf, i)
			if err != nil {
				err := fmt.Errorf("failed to parse heartbeat at offset %d: %w", i, err)
				_ = err
				return events, nil
			}
			events = append(events, parsed)
			size := Qu
			i += size
		}

		return events, nil // empty array
	}
	if msgType != 2 {
		return nil, fmt.Errorf("unexpected message type: %d", msgType)
	}
	// Skip the initial "type" check (always 2 in your case)
	//i += 4 // equivalent to reading the uint32 type at the start

	for i < bufLen {
		// Parse the order/position at current offset
		op, err := Ku(buf, i)
		if err != nil {
			// fmt.Errorf("failed to parse order/position at offset %d: %w", i, err)
			return events, nil // return what we have so far
		}
		events = append(events, op)

		// Move offset forward by the total bytes of this order/position
		size, _ := ParseOrderPositions(buf, i) // returns total bytes consumed
		i += size
	}

	return events, nil
}

// Tp parses a byte slice using Zu fields and returns a Ju instance
func ParseTp(t []byte, e int) (*OrderPosition, error) {
	if len(t)-e < Qu {
		return nil, fmt.Errorf("buffer too small: got %d, need at least %d", len(t)-e, Qu)
	}
	parsedData, err := ec.Parse(t, Zu, e)
	if err != nil {
		return nil, err
	}

	//// Type assertions for the parsed fields
	//flagMask, _ := parsedData[0].(uint32)
	//transactionID, _ := parsedData[1].(uint32)
	//transactionType, _ := parsedData[2].(uint32)
	tradeOrder, _ := parsedData[3].(*Order)

	return &OrderPosition{
		FlagMask:        parsedData[0].(uint32),
		TransactionID:   parsedData[1].(uint32),
		TransactionType: parsedData[2].(uint32),
		TradeOrder:      tradeOrder,
	}, nil
}

// Xd parses []byte into an Ad struct using Id fields
func Xd(t []byte, e int) (interface{}, error) {
	parsed, err := ec.Parse(t, Id, e)
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

func Ku(buf []byte, offset int) (*OrderPosition, error) {
	i := offset

	// Parse the fixed header using ec.Parse
	parsed, err := ec.Parse(buf, OrderPositionProps, i)
	if err != nil {
		return nil, err
	}

	op := &OrderPosition{
		FlagMask:          parsed[0].(uint32),
		TransactionID:     parsed[1].(uint32),
		TransactionType:   parsed[2].(uint32),
		Deal:              parsed[3],             // placeholder, actual parsing happens below
		TradePosition:     parsed[4].(*Position), // placeholder
		Balance:           parsed[5].(float64),
		Credit:            parsed[6].(float64),
		CommissionDaily:   parsed[7].(float64),
		CommissionMonthly: parsed[8].(float64),
		AccProfit:         parsed[9].(float64),
	}

	// Move the offset past the header
	//i += Gu

	//// Parse Deals array
	//deals, _, err := Nd(buf, i)
	//if err != nil {
	//	return nil, err
	//}
	//op.Deals = deals

	//// Move offset past deals
	//ddSize, err := Dd(buf, i)
	//if err != nil {
	//	return nil, err
	//}
	//i += ddSize

	//// Parse Positions array
	//positions, _, err := bu(buf, i)
	//if err != nil {
	//	return nil, err
	//}
	//op.Positions = positions

	//// Move offset past positions
	//fuSize, err := fu(buf, i)
	//if err != nil {
	//	return nil, err
	//}
	//i += fuSize

	return op, nil
}

// Position represents a parsed position (like JS vd)
// ParseNewOrder parses an order from the buffer
func ParseNewOrder(buf []byte, offset int) (interface{}, error) {
	values, err := ec.Parse(buf, NewOrderProps, offset)
	if err != nil {
		return nil, err
	}
	if len(values) < 28 {
		return nil, errors.New("unexpected parse length for new order")
	}

	o := &Deal{
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
	return o, nil
}

// ParseNewPosition parses a position from the buffer
func ParseNewPosition(buf []byte, offset int) (interface{}, error) {
	values, err := ec.Parse(buf, NewPositionProps, offset)
	if err != nil {
		return nil, err
	}
	if len(values) < 25 {
		return nil, errors.New("unexpected parse length for new position")
	}

	p := &Position{
		PositionID:       values[0].(int64),
		TradeOrder:       values[1].(int64),
		TimeCreate:       uint64(1000*values[2].(uint32)) + uint64(values[24].(int32)),
		TimeUpdate:       uint64(1000*values[3].(uint32)) + uint64(values[25].(int32)),
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

	return p, nil
}

func ParseOrderPositions(buf []byte, off int) (int, error) {
	total := 0
	i := off

	// Add the size of the fixed order/position header
	total += Gu
	i += Gu

	// Add the size of the orders array
	ddSize, err := Dd(buf, i)
	if err != nil {
		return 0, err
	}
	total += ddSize
	i += ddSize

	// Add the size of the positions array
	fuSize, err := fu(buf, i)
	if err != nil {
		return 0, err
	}
	total += fuSize
	i += fuSize

	return total, nil
}

// Dd computes the byte size of N orders starting at offset
func Dd(buf []byte, off int) (int, error) {
	if len(buf)-off < 4 {
		return 0, ErrBufferTooShort
	}
	count := int(binary.LittleEndian.Uint32(buf[off:]))
	return 4 + count*Pd, nil
}

// fu computes the byte size of N positions starting at offset
func fu(buf []byte, off int) (int, error) {
	if len(buf)-off < 4 {
		return 0, ErrBufferTooShort
	}
	count := int(binary.LittleEndian.Uint32(buf[off:]))
	return 4 + count*gu, nil
}
