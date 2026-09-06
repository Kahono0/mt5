package mt5

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestDecodeTicksParsesSingleRecord(t *testing.T) {
	body := make([]byte, 50)
	binary.LittleEndian.PutUint32(body[0:4], 77)         // symbol_id
	binary.LittleEndian.PutUint32(body[4:8], 1710000000) // tick_time seconds
	binary.LittleEndian.PutUint32(body[8:12], 5)         // fields
	binary.LittleEndian.PutUint64(body[12:20], math.Float64bits(1.2345))
	binary.LittleEndian.PutUint64(body[20:28], math.Float64bits(1.2347))
	binary.LittleEndian.PutUint64(body[28:36], math.Float64bits(1.2346))
	binary.LittleEndian.PutUint64(body[36:44], uint64(99)) // tick_volume
	binary.LittleEndian.PutUint32(body[44:48], 250)
	binary.LittleEndian.PutUint16(body[48:50], 3)

	ticks, err := DecodeTicks(Message{Command: CommandTickUpdate, Body: body})
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if len(ticks) != 1 {
		t.Fatalf("expected 1 tick, got %d", len(ticks))
	}

	tick := ticks[0]
	if tick.SymbolID != 77 || tick.Fields != 5 || tick.TickVolume != 99 || tick.Flags != 3 {
		t.Fatalf("unexpected parsed tick values: %+v", tick)
	}
	if tick.TimeMSDelta != 250 {
		t.Fatalf("expected delta 250, got %d", tick.TimeMSDelta)
	}
	if tick.TickTimeMS != tick.TickTime+250 {
		t.Fatalf("tick time ms mismatch")
	}
}

func TestDecodeTicksRejectsWrongCommand(t *testing.T) {
	_, err := DecodeTicks(Message{Command: CommandRequestAccount})
	if err == nil {
		t.Fatalf("expected command mismatch error")
	}
}

func TestDecodeTicksEmptyBody(t *testing.T) {
	ticks, err := DecodeTicks(Message{Command: CommandTickUpdate, Body: []byte{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ticks) != 0 {
		t.Fatalf("expected 0 ticks, got %d", len(ticks))
	}
}

func TestBuildTickSubscriptionPayload(t *testing.T) {
	payload := BuildTickSubscriptionPayload([]uint32{11, 22, 33})
	if len(payload) != 16 {
		t.Fatalf("expected payload length 16, got %d", len(payload))
	}

	count := binary.LittleEndian.Uint32(payload[0:4])
	if count != 3 {
		t.Fatalf("expected count 3, got %d", count)
	}

	if binary.LittleEndian.Uint32(payload[4:8]) != 11 ||
		binary.LittleEndian.Uint32(payload[8:12]) != 22 ||
		binary.LittleEndian.Uint32(payload[12:16]) != 33 {
		t.Fatalf("unexpected symbol id sequence in payload")
	}
}

func TestBuildTradeHistoryPayload(t *testing.T) {
	payload := BuildTradeHistoryPayload(-10, 20)
	if len(payload) != 8 {
		t.Fatalf("expected payload length 8, got %d", len(payload))
	}

	from := int32(binary.LittleEndian.Uint32(payload[0:4]))
	to := int32(binary.LittleEndian.Uint32(payload[4:8]))
	if from != -10 || to != 20 {
		t.Fatalf("unexpected from/to values: %d %d", from, to)
	}
}
