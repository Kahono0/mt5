package mt5

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

type Symbol struct {
	Name          string
	Description   string
	Digits        uint32
	ID            uint32
	Path          string
	TradeCalcMode uint32
	Basis         string
	Sector        uint16
}

type SymbolList struct {
	Symbols []Symbol
}

// QuerySymbolsHandler is a function that processes decoded symbol data.
// Handle errors within the callback; only framework-level decode
// errors are routed to [Config.OnError] and the error channel.
type QuerySymbolsHandler func(symbols *SymbolList, msg Message)

// QuerySymbols sends a request for all symbols
// The client must be authenticated
func (c *AppClient) QuerySymbols() error {
	if !c.IsAuthenticated() {
		return ErrNotAuthenticated
	}

	if err := c.core.SendCommand(CommandQuerySymbols, nil); err != nil {
		return fmt.Errorf("failed to query symbols: %w", err)
	}
	return nil
}

// RegisterQuerySymbolsHandler registers a handler that receives decoded
// symbol data. Use [WithErrorHandler] to route errors from this handler
// to a dedicated callback.
func (c *AppClient) RegisterQuerySymbolsHandler(handler QuerySymbolsHandler, opts ...HandlerOption) {
	if handler == nil {
		return
	}

	c.RegisterHandler(CommandQuerySymbols, func(msg Message) error {
		info, err := DecodeSymbols(msg)
		if err != nil {
			return err
		}
		handler(info, msg)
		return nil
	}, opts...)
}

func DecodeSymbols(msg Message) (*SymbolList, error) {
	if msg.Command != CommandQuerySymbols {
		return nil, &CommandMismatchError{Expected: CommandQuerySymbols, Got: msg.Command}
	}
	body := msg.Body
	const recordSize = 526

	if len(body) < 4 {
		return nil, fmt.Errorf("symbol response too short: %d bytes", len(body))
	}

	count := binary.LittleEndian.Uint32(body[:4])
	remaining := len(body) - 4

	if uint64(count)*recordSize > uint64(remaining) {
		return nil, fmt.Errorf(
			"invalid symbol count %d: need %d bytes, have %d",
			count,
			uint64(count)*recordSize,
			remaining,
		)
	}

	result := &SymbolList{
		Symbols: make([]Symbol, 0, count),
	}

	offset := 4

	i := 0
	for range count {
		record := body[offset : offset+recordSize]

		symbol := Symbol{
			Name:          readUTF16LE(record[0:64]),
			Description:   readUTF16LE(record[64:192]),
			Digits:        binary.LittleEndian.Uint32(record[192:196]),
			ID:            binary.LittleEndian.Uint32(record[196:200]),
			Path:          readUTF16LE(record[200:456]),
			TradeCalcMode: binary.LittleEndian.Uint32(record[456:460]),
			Basis:         readUTF16LE(record[460:524]),
			Sector:        binary.LittleEndian.Uint16(record[524:526]),
		}

		i++

		fmt.Printf("Parsed symbol %d Name: %s\n", i, symbol.Name)

		result.Symbols = append(result.Symbols, symbol)
		offset += recordSize
	}

	return result, nil
}

func readUTF16LE(buf []byte) string {
	if len(buf)%2 != 0 {
		buf = buf[:len(buf)-1]
	}

	units := make([]uint16, len(buf)/2)
	for i := range units {
		units[i] = binary.LittleEndian.Uint16(buf[i*2:])
		if units[i] == 0 {
			units = units[:i]
			break
		}
	}

	return string(utf16.Decode(units))
}
