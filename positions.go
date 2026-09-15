package mt5

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/kahono0/mt5/internal/ec"
)

type Position struct {
	PositionID       int64
	TradeOrder       int64
	TimeCreate       uint64
	TimeUpdate       uint64
	TradeSymbol      string
	TradeAction      uint32
	PriceOpen        float64
	PriceClose       float64
	SL               float64
	TP               float64
	TradeVolume      uint64
	Profit           float64
	RateProfit       float64
	RateMargin       float64
	Commission       float64
	Storage          float64
	Expert           int64
	ExpertPositionID int64
	Comment          string
	ContractSize     float64
	Digits           uint32
	DigitsCurrency   uint32
	TradeReason      uint32
	ExternalID       string
	TimeSubSecondsC  int32 // i[24] in JS
	TimeSubSecondsU  int32 // i[25] in JS
}

type PositionsHandler func(positions []Position, msg Message) error

func (c *AppClient) RegisterPositionsHandler(handler PositionsHandler) {
	if handler == nil {
		return
	}

	c.OnT(Positions, func(positions []Position, msg Message) error {
		return handler(positions, msg)
	})
}

func DecodePositions(msg Message) ([]Position, error) {
	if msg.Command != CommandRequestPositions {
		return nil, fmt.Errorf("unexpected command %d for positions", msg.Command)
	}

	positions, err := ParsePositions(msg.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode positions: %w", err)
	}

	return positions, nil
}

// ParsePositions deserializes an ArrayBuffer response into positions.
func ParsePositions(resp []byte) ([]Position, error) {
	// fields []ec.Field, entrySize int
	fields := TradePositionFields
	entrySize, err := ec.GetSeriesSize(fields)
	if err != nil {
		return nil, err
	}
	if len(resp) < 4 {
		return nil, errors.New("response too short")
	}

	// First 4 bytes = number of positions (little endian)
	count := binary.LittleEndian.Uint32(resp[:4])
	offset := 4
	positions := make([]Position, 0, count)

	for i := uint32(0); i < count; i++ {
		if offset+entrySize > len(resp) {
			return nil, errors.New("buffer too short for all positions")
		}

		parsed, err := ec.Parse(resp[offset:offset+entrySize], fields, 0)
		if err != nil {
			return nil, err
		}

		// Map parsed slice to Position struct
		pos := Position{
			PositionID:       parsed[0].(int64),
			TradeOrder:       parsed[1].(int64),
			TimeCreate:       uint64(parsed[2].(uint32))*1000 + uint64(parsed[24].(int32)),
			TimeUpdate:       uint64(parsed[3].(uint32))*1000 + uint64(parsed[25].(int32)),
			TradeSymbol:      parsed[4].(string),
			TradeAction:      parsed[5].(uint32),
			PriceOpen:        parsed[6].(float64),
			PriceClose:       parsed[7].(float64),
			SL:               parsed[8].(float64),
			TP:               parsed[9].(float64),
			TradeVolume:      parsed[10].(uint64),
			Profit:           parsed[11].(float64),
			RateProfit:       parsed[12].(float64),
			RateMargin:       parsed[13].(float64),
			Commission:       parsed[14].(float64),
			Storage:          parsed[15].(float64),
			Expert:           parsed[16].(int64),
			ExpertPositionID: parsed[17].(int64),
			Comment:          parsed[18].(string),
			ContractSize:     parsed[19].(float64),
			Digits:           parsed[20].(uint32),
			DigitsCurrency:   parsed[21].(uint32),
			TradeReason:      parsed[22].(uint32),
			ExternalID:       parsed[23].(string),
		}

		positions = append(positions, pos)
		offset += entrySize
	}

	return positions, nil
}

var TradePositionFields = []ec.Field{
	{PropType: 17},                 // 0
	{PropType: 17},                 // 1
	{PropType: 6},                  // 2
	{PropType: 6},                  // 3
	{PropType: 11, PropLength: 64}, // 4
	{PropType: 6},                  // 5
	{PropType: 8},                  // 6
	{PropType: 8},                  // 7
	{PropType: 8},                  // 8
	{PropType: 8},                  // 9
	{PropType: 18},                 // 10
	{PropType: 8},                  // 11
	{PropType: 8},                  // 12
	{PropType: 8},                  // 13
	{PropType: 8},                  // 14
	{PropType: 8},                  // 15
	{PropType: 17},                 // 16
	{PropType: 17},                 // 17
	{PropType: 11, PropLength: 64}, // 18
	{PropType: 8},                  // 19
	{PropType: 6},                  // 20
	{PropType: 6},                  // 21
	{PropType: 6},                  // 22
	{PropType: 11, PropLength: 64}, // 23
	{PropType: 3},                  // 24
	{PropType: 3},                  // 25
}
