package mt5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/kahono0/mt5/internal/ec"
)

// TODO: Fix parsing account info and improve dev experience on flags and permissions (add something like enum for those fields)

// --- AccountTradeSettings struct ---
type AccountTradeSettings struct {
	SymbolPath             string
	SpreadDiff             int32
	SpreadDiffBalance      int32
	TradeMode              uint32
	TradeStopsLevel        int32
	TradeFreezeLevel       int32
	TradeExemode           uint32
	TradeFillFlags         uint32
	TradeTimeFlags         uint32
	TradeOrderFlags        uint32
	TradeRequestFlags      uint32
	TradeRequestTimeout    uint32
	TradeInstantFlags      uint32
	TradeInstantTimeout    uint32
	TradeInstantSlipProfit uint32
	TradeInstantSlipLosing uint32
	TradeInstantVolume     uint64
	TradeInstantCheckmode  uint32
	TradeMarketFlags       uint32
	TradeExchangeFlags     uint32
	VolumeMin              uint64
	VolumeMax              uint64
	VolumeStep             uint64
	VolumeLimit            uint64
	MarginFlags            uint32
	MarginInitial          float64
	MarginMaintenance      float64
	MarginRatesInitial     []float64
	MarginRatesMaintenance []float64
	MarginRateLiquidity    float64
	MarginHedged           float64
	MarginRateCurrency     float64
	SwapMode               uint32
	SwapLong               float64
	SwapShort              float64
	SwapRollover3Days      int32
	PermissionsFlags       uint32
	PermissionsBookdepth   uint32
}

// AccountCore contains the fixed-size account fields at the beginning of the response.
type AccountCore struct {
	AccountType          uint8
	Rights               int32
	PermissionsFlags     int32
	Balance              float64
	Credit               float64
	AccountCurrency      string
	CurrencyDigits       uint32
	MarginLeverage       uint32
	AccountName          string
	ServerBuild          uint16
	ServerName           string
	Company              string
	TimezoneShift        int32
	DaylightMode         int8
	MarginMode           uint32
	MarginFreeMode       uint32
	MarginSoCall         float64
	MarginSoSo           float64
	MarginSoMode         uint32
	MarginVirtual        float64
	MarginFreeProfitMode uint32
	AccProfit            float64
	CommissionDaily      float64
	CommissionMonthly    float64
	AuthPasswordMin      uint32
	OTPStatus            uint32
}

// --- AccountInfo struct ---
type AccountInfo struct {
	Core          AccountCore
	TradeSettings []AccountTradeSettings
	ValueO        int32
	ValueN        int32
	ValueC        uint64
	H             [][]any
	Underscore    [][]any
}

func mapToAccountCore(values []any) (AccountCore, error) {
	if len(values) != 26 {
		return AccountCore{}, fmt.Errorf("expected 26 core account values, got %d", len(values))
	}
	return AccountCore{
		AccountType:          values[0].(uint8),
		Rights:               values[1].(int32),
		PermissionsFlags:     values[2].(int32),
		Balance:              values[3].(float64),
		Credit:               values[4].(float64),
		AccountCurrency:      values[5].(string),
		CurrencyDigits:       values[6].(uint32),
		MarginLeverage:       values[7].(uint32),
		AccountName:          values[8].(string),
		ServerBuild:          values[9].(uint16),
		ServerName:           values[10].(string),
		Company:              values[11].(string),
		TimezoneShift:        values[12].(int32),
		DaylightMode:         values[13].(int8),
		MarginMode:           values[14].(uint32),
		MarginFreeMode:       values[15].(uint32),
		MarginSoCall:         values[16].(float64),
		MarginSoSo:           values[17].(float64),
		MarginSoMode:         values[18].(uint32),
		MarginVirtual:        values[19].(float64),
		MarginFreeProfitMode: values[20].(uint32),
		AccProfit:            values[21].(float64),
		CommissionDaily:      values[22].(float64),
		CommissionMonthly:    values[23].(float64),
		AuthPasswordMin:      values[24].(uint32),
		OTPStatus:            values[25].(uint32),
	}, nil
}

// --- Helper functions ---
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
func parseFields(buf []byte, offset *int, fields []ec.Field) ([]any, error) {
	out := make([]any, len(fields))

	for i, f := range fields {
		switch f.PropType {
		case 1: // int8
			if *offset+1 > len(buf) {
				out[i] = int8(0)
				*offset = len(buf)
				continue
			}
			out[i] = int8(buf[*offset])
			*offset += 1

		case 4: // uint8
			if *offset+1 > len(buf) {
				out[i] = uint8(0)
				*offset = len(buf)
				continue
			}
			out[i] = uint8(buf[*offset])
			*offset += 1

		case 2: // int16
			if *offset+2 > len(buf) {
				out[i] = int16(0)
				*offset = len(buf)
				continue
			}
			out[i] = int16(binary.LittleEndian.Uint16(buf[*offset:]))
			*offset += 2

		case 5: // uint16
			if *offset+2 > len(buf) {
				out[i] = uint16(0)
				*offset = len(buf)
				continue
			}
			out[i] = uint16(binary.LittleEndian.Uint16(buf[*offset:]))
			*offset += 2

		case 3: // int32
			if *offset+4 > len(buf) {
				out[i] = int32(0)
				*offset = len(buf)
				continue
			}
			out[i] = int32(binary.LittleEndian.Uint32(buf[*offset:]))
			*offset += 4

		case 6: // uint32
			if *offset+4 > len(buf) {
				out[i] = uint32(0)
				*offset = len(buf)
				continue
			}
			out[i] = binary.LittleEndian.Uint32(buf[*offset:])
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
	for i := range n {
		out[i] = math.Float64frombits(binary.LittleEndian.Uint64(b[i*8:]))
	}
	return out
}

// --- Map parsed values to AccountTradeSettings ---
func mapToTradeSettings(values []any) AccountTradeSettings {
	return AccountTradeSettings{
		SymbolPath:             values[0].(string),
		SpreadDiff:             values[1].(int32),
		SpreadDiffBalance:      values[2].(int32),
		TradeMode:              values[3].(uint32),
		TradeStopsLevel:        values[4].(int32),
		TradeFreezeLevel:       values[5].(int32),
		TradeExemode:           values[6].(uint32),
		TradeFillFlags:         values[7].(uint32),
		TradeTimeFlags:         values[8].(uint32),
		TradeOrderFlags:        values[9].(uint32),
		TradeRequestFlags:      values[10].(uint32),
		TradeRequestTimeout:    values[11].(uint32),
		TradeInstantFlags:      values[12].(uint32),
		TradeInstantTimeout:    values[13].(uint32),
		TradeInstantSlipProfit: values[14].(uint32),
		TradeInstantSlipLosing: values[15].(uint32),
		TradeInstantVolume:     values[16].(uint64),
		TradeInstantCheckmode:  values[17].(uint32),
		TradeMarketFlags:       values[18].(uint32),
		TradeExchangeFlags:     values[19].(uint32),
		VolumeMin:              values[20].(uint64),
		VolumeMax:              values[21].(uint64),
		VolumeStep:             values[22].(uint64),
		VolumeLimit:            values[23].(uint64),
		MarginFlags:            values[24].(uint32),
		MarginInitial:          values[25].(float64),
		MarginMaintenance:      values[26].(float64),
		MarginRatesInitial:     values[27].([]float64),
		MarginRatesMaintenance: values[28].([]float64),
		MarginRateLiquidity:    values[29].(float64),
		MarginHedged:           values[30].(float64),
		MarginRateCurrency:     values[31].(float64),
		SwapMode:               values[32].(uint32),
		SwapLong:               values[33].(float64),
		SwapShort:              values[34].(float64),
		SwapRollover3Days:      values[35].(int32),
		PermissionsFlags:       values[36].(uint32),
		PermissionsBookdepth:   values[37].(uint32),
	}
}

func parseH(buf []byte, offset *int, fields []ec.Field) ([][]any, error) {
	if len(buf) <= *offset {
		return nil, nil
	}
	count := int(readInt32(buf, offset))
	if count == -1 {
		return [][]any{}, nil
	}
	if count < 0 {
		return nil, fmt.Errorf("invalid H record count: %d", count)
	}
	result := make([][]any, 0, count)
	valueFields := []ec.Field{
		{PropType: 8}, {PropType: 8}, {PropType: 8}, {PropType: 8},
	}
	for range count {
		values, err := parseFields(buf, offset, fields)
		if err != nil {
			return nil, err
		}
		nestedCount := int(values[4].(int32))
		if nestedCount == -1 {
			nestedCount = 0
		}
		if nestedCount < 0 {
			return nil, fmt.Errorf("invalid H nested record count: %d", nestedCount)
		}
		nested := make([][]any, 0, nestedCount)
		for j := 0; j < nestedCount; j++ {
			value, err := parseFields(buf, offset, valueFields)
			if err != nil {
				return nil, err
			}
			nested = append(nested, value)
		}
		result = append(result, append(values, nested))
	}
	return result, nil
}

func parseUnderscore(buf []byte, offset *int, fields []ec.Field) ([][]any, error) {
	if len(buf) <= *offset {
		return nil, nil
	}
	count := int(readInt32(buf, offset))
	if count == -1 {
		return [][]any{}, nil
	}
	if count < 0 {
		return nil, fmt.Errorf("invalid underscore record count: %d", count)
	}
	fmt.Println("count is", count)
	result := make([][]any, 0, count)
	valueFields := []ec.Field{
		{PropType: 6}, {PropType: 6}, {PropType: 8}, {PropType: 8},
		{PropType: 8}, {PropType: 8}, {PropType: 8}, {PropType: 11, PropLength: 32},
	}
	for range count {
		values, err := parseFields(buf, offset, fields)
		if err != nil {
			return nil, err
		}
		nestedCount := int(readUint32(buf, offset))
		nested := make([][]any, 0, nestedCount)
		for range nestedCount {
			value, err := parseFields(buf, offset, valueFields)
			if err != nil {
				return nil, err
			}
			nested = append(nested, value)
		}
		result = append(result, append(values, nested))
	}
	return result, nil
}

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
	accountTradeSettingsFields := []ec.Field{
		{PropType: 11, PropLength: 256}, {PropType: 3}, {PropType: 3}, {PropType: 6},
		{PropType: 3}, {PropType: 3}, {PropType: 6}, {PropType: 6}, {PropType: 6}, {PropType: 6},
		{PropType: 6}, {PropType: 6}, {PropType: 6}, {PropType: 6}, {PropType: 6}, {PropType: 6},
		{PropType: 18}, {PropType: 6}, {PropType: 6}, {PropType: 6}, {PropType: 18}, {PropType: 18},
		{PropType: 18}, {PropType: 18}, {PropType: 6}, {PropType: 8}, {PropType: 8},
		{PropType: 12, PropLength: 64}, {PropType: 12, PropLength: 64}, {PropType: 8}, {PropType: 8},
		{PropType: 8}, {PropType: 6}, {PropType: 8}, {PropType: 8}, {PropType: 3}, {PropType: 6}, {PropType: 6},
	}
	gl := []ec.Field{
		{PropType: 11, PropLength: 256}, {PropType: 6}, {PropType: 11, PropLength: 32}, {PropType: 6}, {PropType: 3},
	}

	_l := []ec.Field{
		{PropType: 11, PropLength: 256}, {PropType: 6}, {PropType: 6}, {PropType: 6}, {PropType: 11, PropLength: 32}, {PropType: 6},
	}

	_ = gl
	_ = _l

	info := &AccountInfo{}

	// Parse Core Values
	core, err := parseFields(respBody, &offset, vl)
	if err != nil {
		return nil, err
	}
	info.Core, err = mapToAccountCore(core)
	if err != nil {
		return nil, err
	}
	if info.Core.AuthPasswordMin == 0 {
		info.Core.AuthPasswordMin = 8
	}
	fmt.Println("offset ", offset)
	// Parse TradeSettings
	if offset < len(respBody) {
		count := int(readUint32(respBody, &offset))
		info.TradeSettings = make([]AccountTradeSettings, 0, count)

		for i := range count {
			v, err := parseFields(respBody, &offset, accountTradeSettingsFields)
			if err != nil {
				return nil, err
			}
			if i == 0 {
				info.TradeSettings = append(info.TradeSettings, mapToTradeSettings(v))
			}

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

	// // Parse nested arrays
	// info.H, err = parseH(respBody, &offset, gl)
	// if err != nil {
	// 	return nil, err
	// }
	// info.Underscore, err = parseUnderscore(respBody, &offset, _l)
	// if err != nil {
	// 	return nil, err
	// }

	return info, nil
}
