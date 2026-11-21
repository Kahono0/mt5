package mt5

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/kahono0/mt5/internal/ec"
)

// TODO: Fix parsing account info and improve dev experience on flags and permissions (add something like enum for those fields)

// --- AccountTradeSettings struct ---
type AccountTradeSettings struct {
	SymbolPath             string
	SpreadDiff             uint32
	SpreadDiffBalance      uint32
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

// AccountInfoHandler is a function that processes decoded account data.
// Handle errors within the callback; only framework-level decode
// errors are routed to [Config.OnError] and the error channel.
type AccountInfoHandler func(info *AccountInfo, msg Message)

// RequestAccountInfo sends a request for account information.
// The client must be authenticated.
func (c *AppClient) RequestAccountInfo() error {
	if !c.IsAuthenticated() {
		return ErrNotAuthenticated
	}
	if err := c.core.SendCommand(CommandRequestAccountInfo, nil); err != nil {
		return fmt.Errorf("failed to request account info: %w", err)
	}
	return nil
}

// GetAccountInfo sends a request for account information and blocks until
// the response is received or the context is cancelled. This is a synchronous
// wrapper around [RequestAccountInfo] and [RegisterAccountInfoHandler].
func (c *AppClient) GetAccountInfo(ctx context.Context) (*AccountInfo, error) {
	if !c.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}

	type result struct {
		info *AccountInfo
		err  error
	}
	ch := make(chan result, 1)

	cancel := c.RegisterAccountInfoHandler(func(info *AccountInfo, msg Message) {
		ch <- result{info, nil}
	}, WithErrorHandler(func(err error) {
		ch <- result{nil, err}
	}))
	defer cancel()

	if err := c.RequestAccountInfo(); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-ch:
		return r.info, r.err
	}
}

// RegisterAccountInfoHandler registers a handler that receives decoded
// account information. Use [WithErrorHandler] to route errors from this
// handler to a dedicated callback. Returns a function that, when called,
// removes the handler.
func (c *AppClient) RegisterAccountInfoHandler(handler AccountInfoHandler, opts ...HandlerOption) func() {
	if handler == nil {
		return func() {}
	}

	return c.RegisterHandler(CommandRequestAccountInfo, func(msg Message) error {
		info, err := DecodeAccountInfo(msg)
		if err != nil {
			return err
		}
		handler(info, msg)
		return nil
	}, opts...)
}

func DecodeAccountInfo(msg Message) (*AccountInfo, error) {
	if msg.Command != CommandRequestAccountInfo {
		return nil, &CommandMismatchError{Expected: CommandRequestAccountInfo, Got: msg.Command}
	}

	info, err := ParseAccountInfo(msg.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode account info: %w", err)
	}

	return info, nil
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

func parseFloat64ArrayWithOffset(b []byte, offset int) (interface{}, error) {
	if offset < 0 || offset >= len(b) {
		return nil, fmt.Errorf("offset %d out of bounds for byte slice of length %d", offset, len(b))
	}
	n := (len(b) - offset) / 8
	out := make([]float64, n)
	for i := range n {
		out[i] = math.Float64frombits(binary.LittleEndian.Uint64(b[offset+i*8:]))
	}
	return out, nil
}

// --- Map parsed values to AccountTradeSettings ---
func mapToTradeSettings(values []any) AccountTradeSettings {
	return AccountTradeSettings{
		SymbolPath:             values[0].(string),
		SpreadDiff:             values[1].(uint32),
		SpreadDiffBalance:      values[2].(uint32),
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

func ParseAccountInfo(respBody []byte) (*AccountInfo, error) {
	offset := 0

	// Define your fields and series sizes
	vl := []ec.Field{
		{PropType: 4}, {PropType: 3}, {PropType: 3}, {PropType: 8}, {PropType: 8},
		{
			PropType:   11,
			PropLength: 64,
		}, {PropType: 6}, {PropType: 6}, {PropType: 11, PropLength: 256},
		{PropType: 5}, {PropType: 11, PropLength: 128}, {PropType: 11, PropLength: 256},
		{PropType: 3}, {PropType: 1}, {PropType: 6}, {PropType: 6}, {PropType: 8}, {PropType: 8},
		{PropType: 6}, {PropType: 8}, {PropType: 6}, {PropType: 8}, {PropType: 8}, {PropType: 8},
		{PropType: 6}, {PropType: 6},
	}
	accountTradeSettingsFields := []ec.Field{
		{PropType: 11, PropLength: 256},
		{PropType: 6},
		{PropType: 6},
		{PropType: 6},
		{PropType: 3},
		{PropType: 3},
		{PropType: 6},
		{PropType: 6},
		{PropType: 6},
		{PropType: 6},
		{PropType: 6},
		{PropType: 6},
		{PropType: 6},
		{PropType: 6},
		{PropType: 6},
		{PropType: 6},
		{PropType: 18},
		{PropType: 6},
		{PropType: 6},
		{PropType: 6},
		{PropType: 18},
		{PropType: 18},
		{PropType: 18},
		{PropType: 18},
		{PropType: 6},
		{PropType: 8},
		{PropType: 8},
		{PropType: 12, PropLength: 64, Parser: parseFloat64ArrayWithOffset},
		{PropType: 12, PropLength: 64, Parser: parseFloat64ArrayWithOffset},
		{PropType: 8},
		{PropType: 8},
		{PropType: 8},
		{PropType: 6},
		{PropType: 8},
		{PropType: 8},
		{PropType: 3},
		{PropType: 6},
		{PropType: 6},

		{PropType: 8},
		{PropType: 8},
		{PropType: 8},
		{PropType: 8},
		{PropType: 8},
		{PropType: 8},
		{PropType: 8},
	}
	gl := []ec.Field{
		{
			PropType:   11,
			PropLength: 256,
		}, {PropType: 6}, {PropType: 11, PropLength: 32}, {PropType: 6}, {PropType: 3},
	}

	_l := []ec.Field{
		{
			PropType:   11,
			PropLength: 256,
		}, {PropType: 6}, {PropType: 6}, {PropType: 6}, {PropType: 11, PropLength: 32}, {PropType: 6},
	}

	_ = gl
	_ = _l

	info := &AccountInfo{}

	// Parse Core Values
	core, err := ec.Parse(respBody, vl, offset)
	if err != nil {
		return nil, err
	}
	s, err := ec.GetSeriesSize(vl)
	if err != nil {
		return nil, err
	}

	offset += s
	info.Core, err = mapToAccountCore(core)
	if err != nil {
		return nil, err
	}
	if info.Core.AuthPasswordMin == 0 {
		info.Core.AuthPasswordMin = 8
	}
	s, err = ec.GetSeriesSize(accountTradeSettingsFields)
	if err != nil {
		return nil, err
	}
	tradeSettingsSchema := ec.CompileSchema(accountTradeSettingsFields)
	// Parse TradeSettings
	if offset < len(respBody) {
		count := binary.LittleEndian.Uint32(respBody[offset:])
		offset += 4
		info.TradeSettings = make([]AccountTradeSettings, 0, count)

		for i := range count {
			v, err := tradeSettingsSchema.Parse(respBody, offset)
			if err != nil {
				return nil, err
			}
			offset += s
			if i >= 0 {
				info.TradeSettings = append(info.TradeSettings, mapToTradeSettings(v))
			}
		}
	}

	// Parse remaining values
	if offset < len(respBody) {
		info.ValueO = int32(binary.LittleEndian.Uint32(respBody[offset:]))
		offset += 4
	}
	if offset < len(respBody) {
		info.ValueN = int32(binary.LittleEndian.Uint32(respBody[offset:]))
		offset += 4
	}
	if offset < len(respBody) {
		info.ValueC = binary.LittleEndian.Uint64(respBody[offset:])
		offset += 8
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
