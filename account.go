package mt5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/kahono0/mt5/internal/ec"
)

type AccountTradeSettings struct {
	SymbolPath             string
	SpreadDiff             int32
	SpreadDiffBalance      int32
	TradeMode              int32
	TradeStopsLevel        int32
	TradeFreezeLevel       int32
	TradeExemode           int32
	TradeFillFlags         int32
	TradeTimeFlags         int32
	TradeOrderFlags        int32
	TradeRequestFlags      int32
	TradeRequestTimeout    int32
	TradeInstantFlags      int32
	TradeInstantTimeout    int32
	TradeInstantSlipProfit int32
	TradeInstantSlipLosing int32
	TradeInstantVolume     uint64
	//TradeInstantCheckmode  int32
	//TradeMarketFlags       int32
	//TradeExchangeFlags     int32
	//VolumeMin              float64
	//VolumeMax              float64
	//VolumeStep             float64
	//VolumeLimit            float64
	//MarginFlags            int32
	//MarginInitial          float64
	//MarginMaintenance      float64
	//MarginRatesInitial     float64
	//MarginRatesMaintenance float64
	//MarginRateLiquidity    float64
	//MarginHedged           float64
	//MarginRateCurrency     float64
	//SwapMode               int32
	//SwapLong               float64
	//SwapShort              float64
	//SwapRollover3Days      float64
	//PermissionsFlags       int32
	//PermissionsBookdepth   int32
}

type AccountInfo struct {
	CoreValues    []interface{}
	TradeSettings []AccountTradeSettings
	ValueO        int32
	ValueN        int32
	ValueC        uint64
	H             [][]interface{}
	Underscore    [][]interface{}
}

type AccountInfoHandler func(info *AccountInfo, msg Message) error

func ParseAccountInfo(respBody []byte) (*AccountInfo, error) {
	offset := 0

	// Define your fields and series sizes
	vl := []ec.Field{
		{PropType: 4}, {PropType: 3}, {PropType: 3}, {PropType: 8}, {PropType: 8},
		{PropType: 11, PropLength: 64}, {PropType: 6}, {PropType: 6}, {PropType: 11, PropLength: 256},
		{PropType: 5}, {PropType: 11, PropLength: 128}, {PropType: 11, PropLength: 256},
		{PropType: 3}, {PropType: 1}, {PropType: 6}, {PropType: 6}, {PropType: 8}, {PropType: 8},
		{PropType: 6}, {PropType: 8}, {PropType: 6}, {PropType: 8}, {PropType: 8}, {PropType: 8},
		{PropType: 6}, {PropType: 6},
	}
	kl, _ := ec.GetSeriesSize(vl)

	_ = kl

	accountTradeSettingsFields := []ec.Field{
		{PropType: 11, PropLength: 256}, {PropType: 3}, {PropType: 3}, {PropType: 6},
		{PropType: 3}, {PropType: 3}, {PropType: 6}, {PropType: 6}, {PropType: 6}, {PropType: 6},
		{PropType: 6}, {PropType: 6}, {PropType: 6}, {PropType: 6}, {PropType: 6}, {PropType: 6},
		{PropType: 18}, {PropType: 6}, {PropType: 6}, {PropType: 6}, {PropType: 18}, {PropType: 18},
		{PropType: 18}, {PropType: 18}, {PropType: 6}, {PropType: 8}, {PropType: 8},
		{PropType: 12, PropLength: 64}, {PropType: 12, PropLength: 64}, {PropType: 8}, {PropType: 8},
		{PropType: 8}, {PropType: 6}, {PropType: 8}, {PropType: 8}, {PropType: 3}, {PropType: 6}, {PropType: 6},
	}
	cl, _ := ec.GetSeriesSize(accountTradeSettingsFields)

	gl := []ec.Field{
		{PropType: 11, PropLength: 256}, {PropType: 6}, {PropType: 11, PropLength: 32}, {PropType: 6}, {PropType: 3},
	}
	yl, _ := ec.GetSeriesSize(gl)

	_l := []ec.Field{
		{PropType: 11, PropLength: 256}, {PropType: 6}, {PropType: 6}, {PropType: 6}, {PropType: 11, PropLength: 32}, {PropType: 6},
	}
	dl, _ := ec.GetSeriesSize(_l)

	info := &AccountInfo{}

	// Parse Core Values
	core, err := parseFields(respBody, &offset, vl)
	if err != nil {
		return nil, err
	}
	info.CoreValues = core

	// Parse TradeSettings
	if offset < len(respBody) {
		count := int(readUint32(respBody, &offset))
		//info.TradeSettings = make([]AccountTradeSettings, count)
		for i := 0; i < count; i++ {
			_, err := parseFields(respBody, &offset, accountTradeSettingsFields)
			if err != nil {
				return nil, err
			}
			//info.TradeSettings[i] = mapToTradeSettings(values)
			offset += cl
		}
	}

	// Parse remaining values
	if offset < len(respBody) {
		info.ValueO = readInt32(respBody, &offset)
	}
	if offset < len(respBody) {
		info.ValueN = readInt32(respBody, &offset)
	}
	if offset < len(respBody) {
		info.ValueC = readBigUint64(respBody, &offset)
	}

	// Parse nested arrays
	info.H, err = parseNestedArray(respBody, &offset, gl, yl)
	if err != nil {
		return nil, err
	}
	info.Underscore, err = parseNestedArray(respBody, &offset, _l, dl)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func DecodeAccountInfo(msg Message) (*AccountInfo, error) {
	if msg.Command != CommandRequestAccount {
		return nil, fmt.Errorf("unexpected command %d for account info", msg.Command)
	}

	info, err := ParseAccountInfo(msg.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode account info: %w", err)
	}

	return info, nil
}

func (c *AppClient) RegisterAccountInfoHandler(handler AccountInfoHandler) {
	if handler == nil {
		return
	}

	c.RegisterHandler(CommandRequestAccount, func(msg Message) error {
		info, err := DecodeAccountInfo(msg)
		if err != nil {
			return err
		}
		return handler(info, msg)
	})
}

func readUint32(buf []byte, offset *int) uint32 {
	v := binary.LittleEndian.Uint32(buf[*offset:])
	*offset += 4
	return v
}

func readInt32(buf []byte, offset *int) int32 {
	v := int32(binary.LittleEndian.Uint32(buf[*offset:]))
	*offset += 4
	return v
}

func readBigUint64(buf []byte, offset *int) uint64 {
	v := binary.LittleEndian.Uint64(buf[*offset:])
	*offset += 8
	return v
}

func readFloat64(buf []byte, offset *int) float64 {
	v := math.Float64frombits(binary.LittleEndian.Uint64(buf[*offset:]))
	*offset += 8
	return v
}

func parseUTF16(buf []byte, offset *int, length int) string {
	runes := []rune{}
	for i := 0; i < length; i += 2 {
		if *offset+i+1 >= len(buf) {
			break
		}
		val := binary.LittleEndian.Uint16(buf[*offset+i:])
		if val == 0 {
			break
		}
		runes = append(runes, rune(val))
	}
	*offset += length
	return string(runes)
}

// --- Generic parser for fields ---
func parseFields(buf []byte, offset *int, fields []ec.Field) ([]interface{}, error) {
	out := make([]interface{}, len(fields))

	for i, f := range fields {
		switch f.PropType {
		case 1, 4: // int8 / uint8
			if *offset+1 > len(buf) {
				out[i] = int8(0)
				*offset = len(buf)
				continue
			}
			out[i] = buf[*offset]
			*offset += 1

		case 2, 5: // int16 / uint16
			if *offset+2 > len(buf) {
				out[i] = int16(0)
				*offset = len(buf)
				continue
			}
			out[i] = int16(binary.LittleEndian.Uint16(buf[*offset:]))
			*offset += 2

		case 3, 6: // int32 / uint32
			if *offset+4 > len(buf) {
				out[i] = int32(0)
				*offset = len(buf)
				continue
			}
			out[i] = int32(binary.LittleEndian.Uint32(buf[*offset:]))
			*offset += 4

		case 8: // float64
			if *offset+8 > len(buf) {
				out[i] = float64(0)
				*offset = len(buf)
				continue
			}
			out[i] = readFloat64(buf, offset)

		case 11: // UTF-16 string
			length := f.PropLength
			if length == 0 {
				length = len(buf) - *offset
			}
			if *offset+length > len(buf) {
				length = len(buf) - *offset
			}
			out[i] = parseUTF16(buf, offset, length)

		case 12: // ArrayBuffer → Float64Array
			length := f.PropLength
			if length == 0 {
				return nil, errors.New("propLength required for PropType 12")
			}
			remaining := len(buf) - *offset
			if remaining < length {
				length = remaining
			}
			out[i] = parseFloat64Array(buf[*offset : *offset+length])
			*offset += length

		case 17: // int64
			if *offset+8 > len(buf) {
				out[i] = int64(0)
				*offset = len(buf)
				continue
			}
			out[i] = int64(binary.LittleEndian.Uint64(buf[*offset:]))
			*offset += 8

		case 18: // uint64
			if *offset+8 > len(buf) {
				out[i] = uint64(0)
				*offset = len(buf)
				continue
			}
			out[i] = readBigUint64(buf, offset)

		default:
			return nil, fmt.Errorf("unsupported PropType: %d", f.PropType)
		}
	}

	return out, nil
}

// parseFloat64Array converts a []byte slice to a slice of float64
func parseFloat64Array(b []byte) []float64 {
	n := len(b) / 8
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = math.Float64frombits(binary.LittleEndian.Uint64(b[i*8:]))
	}
	return out
}

func parseNestedArray(buf []byte, offset *int, fields []ec.Field, seriesSize int) ([][]interface{}, error) {
	if len(buf) <= *offset {
		return nil, nil
	}
	count := int(readInt32(buf, offset))
	result := make([][]interface{}, count)
	for i := 0; i < count; i++ {
		v, err := parseFields(buf, offset, fields)
		if err != nil {
			return nil, err
		}
		*offset += seriesSize
		result[i] = v
	}
	return result, nil
}
