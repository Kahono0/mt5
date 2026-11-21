package models

import (
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

	// unknown fields, possibly related to permissions or flags
	Unknown1 uint64
	Unknown2 uint64
	Unknown3 uint64
	Unknown4 uint64
	Unknown5 uint64
	Unknown6 uint64
	Unknown7 uint64
}

var AccountTradeSettingsFields = []ec.Field{
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

	// unknown fields, possibly related to permissions or flags
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
}

var AccountTradeSettingsSize, _ = ec.GetSeriesSize(AccountTradeSettingsFields)

func (ats AccountTradeSettings) Fields() []ec.Field {
	return AccountTradeSettingsFields
}

func (ats AccountTradeSettings) FromValues(values []interface{}) any {
	return &AccountTradeSettings{
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

func (ats *AccountTradeSettings) ParseFromBytes(b []byte) error {
	r := recordReader{buf: b}
	ats.SymbolPath = r.readString(256)
	ats.SpreadDiff = r.readUint32()
	ats.SpreadDiffBalance = r.readUint32()
	ats.TradeMode = r.readUint32()
	ats.TradeStopsLevel = r.readInt32()
	ats.TradeFreezeLevel = r.readInt32()
	ats.TradeExemode = r.readUint32()
	ats.TradeFillFlags = r.readUint32()
	ats.TradeTimeFlags = r.readUint32()
	ats.TradeOrderFlags = r.readUint32()
	ats.TradeRequestFlags = r.readUint32()
	ats.TradeRequestTimeout = r.readUint32()
	ats.TradeInstantFlags = r.readUint32()
	ats.TradeInstantTimeout = r.readUint32()
	ats.TradeInstantSlipProfit = r.readUint32()
	ats.TradeInstantSlipLosing = r.readUint32()
	ats.TradeInstantVolume = r.readUint64()
	ats.TradeInstantCheckmode = r.readUint32()
	ats.TradeMarketFlags = r.readUint32()
	ats.TradeExchangeFlags = r.readUint32()
	ats.VolumeMin = r.readUint64()
	ats.VolumeMax = r.readUint64()
	ats.VolumeStep = r.readUint64()
	ats.VolumeLimit = r.readUint64()
	ats.MarginFlags = r.readUint32()
	ats.MarginInitial = r.readFloat64()
	ats.MarginMaintenance = r.readFloat64()
	ats.MarginRatesInitial = r.readFloat64Array(64)
	ats.MarginRatesMaintenance = r.readFloat64Array(64)
	ats.MarginRateLiquidity = r.readFloat64()
	ats.MarginHedged = r.readFloat64()
	ats.MarginRateCurrency = r.readFloat64()
	ats.SwapMode = r.readUint32()
	ats.SwapLong = r.readFloat64()
	ats.SwapShort = r.readFloat64()
	ats.SwapRollover3Days = r.readInt32()
	ats.PermissionsFlags = r.readUint32()
	ats.PermissionsBookdepth = r.readUint32()

	ats.Unknown1 = r.readUint64()
	ats.Unknown2 = r.readUint64()
	ats.Unknown3 = r.readUint64()
	ats.Unknown4 = r.readUint64()
	ats.Unknown5 = r.readUint64()
	ats.Unknown6 = r.readUint64()
	ats.Unknown7 = r.readUint64()
	return r.finish()
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

var CoreProps = []ec.Field{
	{PropType: 4},
	{PropType: 3},
	{PropType: 3},
	{PropType: 8},
	{PropType: 8},
	{PropType: 11, PropLength: 64},
	{PropType: 6},
	{PropType: 6},
	{PropType: 11, PropLength: 256},
	{PropType: 5},
	{PropType: 11, PropLength: 128},
	{PropType: 11, PropLength: 256},
	{PropType: 3},
	{PropType: 1},
	{PropType: 6},
	{PropType: 6},
	{PropType: 8},
	{PropType: 8},
	{PropType: 6},
	{PropType: 8},
	{PropType: 6},
	{PropType: 8},
	{PropType: 8},
	{PropType: 8},
	{PropType: 6},
	{PropType: 6},
}

var CorePropsSize, _ = ec.GetSeriesSize(CoreProps)

func (ac AccountCore) Fields() []ec.Field {
	return CoreProps
}

func (ac AccountCore) FromValues(values []interface{}) any {
	return &AccountCore{
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
	}
}

func (ac *AccountCore) ParseFromBytes(b []byte) error {
	r := recordReader{buf: b}
	ac.AccountType = r.readUint8()
	ac.Rights = r.readInt32()
	ac.PermissionsFlags = r.readInt32()
	ac.Balance = r.readFloat64()
	ac.Credit = r.readFloat64()
	ac.AccountCurrency = r.readString(64)
	ac.CurrencyDigits = r.readUint32()
	ac.MarginLeverage = r.readUint32()
	ac.AccountName = r.readString(256)
	ac.ServerBuild = r.readUint16()
	ac.ServerName = r.readString(128)
	ac.Company = r.readString(256)
	ac.TimezoneShift = r.readInt32()
	ac.DaylightMode = r.readInt8()
	ac.MarginMode = r.readUint32()
	ac.MarginFreeMode = r.readUint32()
	ac.MarginSoCall = r.readFloat64()
	ac.MarginSoSo = r.readFloat64()
	ac.MarginSoMode = r.readUint32()
	ac.MarginVirtual = r.readFloat64()
	ac.MarginFreeProfitMode = r.readUint32()
	ac.AccProfit = r.readFloat64()
	ac.CommissionDaily = r.readFloat64()
	ac.CommissionMonthly = r.readFloat64()
	ac.AuthPasswordMin = r.readUint32()
	ac.OTPStatus = r.readUint32()
	return r.finish()
}

type Unknown1M struct {
	SymbolPath  string
	UnknownInt1 uint32
	Currency    string
	UnknownInt2 uint32
	UnknownInt3 int32
}

func (u Unknown1M) Fields() []ec.Field {
	return Unknown1
}

func (u Unknown1M) FromValues(values []interface{}) any {
	return &Unknown1M{
		SymbolPath:  values[0].(string),
		UnknownInt1: values[1].(uint32),
		Currency:    values[2].(string),
		UnknownInt2: values[3].(uint32),
		UnknownInt3: values[4].(int32),
	}
}

func (u *Unknown1M) ParseFromBytes(b []byte) error {
	r := recordReader{buf: b}
	u.SymbolPath = r.readString(256)
	u.UnknownInt1 = r.readUint32()
	u.Currency = r.readString(32)
	u.UnknownInt2 = r.readUint32()
	u.UnknownInt3 = r.readInt32()
	return r.finish()
}

var Unknown1 = []ec.Field{
	{PropType: 11, PropLength: 256},
	{PropType: 6},
	{PropType: 11, PropLength: 32},
	{PropType: 6},
	{PropType: 3},
}

var Unknown1Size, _ = ec.GetSeriesSize(Unknown1)

var Unknown2 = []ec.Field{
	{PropType: 11, PropLength: 256},
	{PropType: 6},
	{PropType: 6},
	{PropType: 6},
	{PropType: 11, PropLength: 32},
	{PropType: 6},
}

type Unknown2M struct {
	UnknownString  string
	UnknownInt1    uint32
	UnknownInt2    uint32
	UnknownInt3    uint32
	UnknownString2 string
	UnknownInt4    uint32
}

func (u Unknown2M) Fields() []ec.Field {
	return Unknown2
}

func (u Unknown2M) FromValues(values []interface{}) any {
	return &Unknown2M{
		UnknownString:  values[0].(string),
		UnknownInt1:    values[1].(uint32),
		UnknownInt2:    values[2].(uint32),
		UnknownInt3:    values[3].(uint32),
		UnknownString2: values[4].(string),
		UnknownInt4:    values[5].(uint32),
	}
}

func (u *Unknown2M) ParseFromBytes(b []byte) error {
	r := recordReader{buf: b}
	u.UnknownString = r.readString(256)
	u.UnknownInt1 = r.readUint32()
	u.UnknownInt2 = r.readUint32()
	u.UnknownInt3 = r.readUint32()
	u.UnknownString2 = r.readString(32)
	u.UnknownInt4 = r.readUint32()
	return r.finish()
}

var Unknown2Size, _ = ec.GetSeriesSize(Unknown2)

// --- AccountInfo struct ---
type AccountInfo struct {
	Core          AccountCore
	TradeSettings []AccountTradeSettings
	ValueO        int32
	ValueN        int32
	ValueC        uint64
	H             []Unknown1M
	Underscore    []Unknown2M
}

var AccountInfoFields = []ec.Field{
	{PropType: 12, PropLength: CorePropsSize, InnerFields: CoreProps, FromValues: AccountCore{}.FromValues},
	{PropType: 13, PropLength: AccountTradeSettingsSize, PropLengthType: 6, InnerFields: AccountTradeSettingsFields, FromValues: AccountTradeSettings{}.FromValues},
	{PropType: 3},
	{PropType: 3},
	{PropType: 18},
	{PropType: 13, PropLength: Unknown1Size, PropLengthType: 6, InnerFields: Unknown1, FromValues: Unknown1M{}.FromValues},
	{PropType: 13, PropLength: Unknown2Size, PropLengthType: 6, InnerFields: Unknown2, FromValues: Unknown2M{}.FromValues},
}

var AccountInfoSize, _ = ec.GetSeriesSize(AccountInfoFields)

var SchemaAccountInfo = ec.CompileSchema(AccountInfoFields)

func (ai AccountInfo) Fields() []ec.Field {
	return AccountInfoFields
}

func (ai AccountInfo) FromValues(values []interface{}) any {
	tradeSettings := make([]AccountTradeSettings, len(values[1].([]interface{})))
	for i, ts := range values[1].([]interface{}) {
		tradeSettings[i] = *ts.(*AccountTradeSettings)
	}

	h := make([]Unknown1M, len(values[5].([]interface{})))
	for i, u := range values[5].([]interface{}) {
		h[i] = *u.(*Unknown1M)
	}

	underscore := make([]Unknown2M, len(values[6].([]interface{})))
	for i, u := range values[6].([]interface{}) {
		underscore[i] = *u.(*Unknown2M)
	}
	return &AccountInfo{
		Core:          *values[0].(*AccountCore),
		TradeSettings: tradeSettings,
		ValueO:        values[2].(int32),
		ValueN:        values[3].(int32),
		ValueC:        values[4].(uint64),
		H:             h,
		Underscore:    underscore,
	}
}

func (ai *AccountInfo) ParseFromBytes(b []byte) error {
	core := AccountCore{}
	if err := core.ParseFromBytes(b[:CorePropsSize]); err != nil {
		return fmt.Errorf("failed to parse AccountCore: %w", err)
	}
	ai.Core = core

	offset := CorePropsSize

	count, newOffset, err := readRecordCount(b, offset, AccountTradeSettingsSize, "AccountTradeSettings")
	if err != nil {
		return fmt.Errorf("failed to read AccountTradeSettings count: %w", err)
	}
	offset = newOffset

	ai.TradeSettings = make([]AccountTradeSettings, count)
	for i := 0; i < count; i++ {
		ts := AccountTradeSettings{}
		if err := ts.ParseFromBytes(b[offset : offset+AccountTradeSettingsSize]); err != nil {
			return fmt.Errorf("failed to parse AccountTradeSettings at index %d: %w", i, err)
		}
		ai.TradeSettings[i] = ts
		offset += AccountTradeSettingsSize
	}

	valuesSize := 4 + 4 + 8 // int32 + int32 + uint64
	r := recordReader{buf: b[offset : offset+valuesSize]}
	ai.ValueO = r.readInt32()
	ai.ValueN = r.readInt32()
	ai.ValueC = r.readUint64()

	if err := r.finish(); err != nil {
		return fmt.Errorf("failed to parse ValueO, ValueN, ValueC: %w", err)
	}
	offset += valuesSize

	countH, newOffset, err := readRecordCount(b, offset, Unknown1Size, "Unknown1M")
	if err != nil {
		return fmt.Errorf("failed to read Unknown1M count: %w", err)
	}
	offset = newOffset

	ai.H = make([]Unknown1M, countH)
	for i := 0; i < countH; i++ {
		u := Unknown1M{}
		if err := u.ParseFromBytes(b[offset : offset+Unknown1Size]); err != nil {
			return fmt.Errorf("failed to parse Unknown1M at index %d: %w", i, err)
		}
		ai.H[i] = u
		offset += Unknown1Size
	}

	countUnderscore, newOffset, err := readRecordCount(b, offset, Unknown2Size, "Unknown2M")
	if err != nil {
		return fmt.Errorf("failed to read Unknown2M count: %w", err)
	}
	offset = newOffset

	ai.Underscore = make([]Unknown2M, countUnderscore)
	for i := 0; i < countUnderscore; i++ {
		u := Unknown2M{}
		if err := u.ParseFromBytes(b[offset : offset+Unknown2Size]); err != nil {
			return fmt.Errorf("failed to parse Unknown2M at index %d: %w", i, err)
		}
		ai.Underscore[i] = u
		offset += Unknown2Size
	}

	return nil
}

func readRecordCount(body []byte, offset, recordSize int, name string) (int, int, error) {
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
