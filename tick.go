package mt5

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Tick represents a single price tick received from the server.
type Tick struct {
	// SymbolID is the numeric identifier for the symbol.
	SymbolID uint32
	// TickTime is the tick timestamp in milliseconds (seconds * 1000).
	TickTime int64
	// Fields is a bitmask indicating which price fields are present.
	Fields uint32
	// Bid is the bid price.
	Bid float64
	// Ask is the ask price.
	Ask float64
	// Last is the last trade price.
	Last float64
	// TickVolume is the tick volume.
	TickVolume int64
	// TimeMSDelta is the millisecond offset within the second.
	TimeMSDelta uint32
	// Flags contains additional tick flags.
	Flags uint16
	// TickTimeMS is the precise tick timestamp (TickTime + TimeMSDelta).
	TickTimeMS int64
}

// TickHandler is a function that processes decoded tick data.
type TickHandler func(ticks []Tick, msg Message) error

const tickRecordSize = 50

// DecodeTicks parses a tick update message into a slice of [Tick] values.
// The message must have command [CommandTickUpdate].
func DecodeTicks(msg Message) ([]Tick, error) {
	if msg.Command != CommandTickUpdate {
		return nil, fmt.Errorf("unexpected command %d for ticks", msg.Command)
	}

	body := msg.Body
	if len(body) == 0 {
		return []Tick{}, nil
	}

	count := len(body) / tickRecordSize
	ticks := make([]Tick, 0, count)

	for i := 0; i < count; i++ {
		off := i * tickRecordSize
		tickTimeSec := int32(binary.LittleEndian.Uint32(body[off+4 : off+8]))
		tickVolume := int64(binary.LittleEndian.Uint64(body[off+36 : off+44]))
		timeMSDelta := binary.LittleEndian.Uint32(body[off+44 : off+48])
		tickTime := int64(tickTimeSec) * 1000
		tick := Tick{
			SymbolID:    binary.LittleEndian.Uint32(body[off : off+4]),
			TickTime:    tickTime,
			Fields:      binary.LittleEndian.Uint32(body[off+8 : off+12]),
			Bid:         math.Float64frombits(binary.LittleEndian.Uint64(body[off+12 : off+20])),
			Ask:         math.Float64frombits(binary.LittleEndian.Uint64(body[off+20 : off+28])),
			Last:        math.Float64frombits(binary.LittleEndian.Uint64(body[off+28 : off+36])),
			TickVolume:  tickVolume,
			TimeMSDelta: timeMSDelta,
			Flags:       binary.LittleEndian.Uint16(body[off+48 : off+50]),
			TickTimeMS:  tickTime + int64(timeMSDelta),
		}
		ticks = append(ticks, tick)
	}

	return ticks, nil
}

// RegisterTickHandler registers a handler that receives decoded [Tick]
// data for [CommandTickUpdate] messages.
func (c *AppClient) RegisterTickHandler(handler TickHandler) {
	if handler == nil {
		return
	}

	c.RegisterHandler(CommandTickUpdate, func(msg Message) error {
		ticks, err := DecodeTicks(msg)
		if err != nil {
			return err
		}
		return handler(ticks, msg)
	})
}
