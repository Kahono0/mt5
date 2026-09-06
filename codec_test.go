package mt5

import (
	"encoding/binary"
	"testing"
)

func TestDefaultCodecPackAndUnpack(t *testing.T) {
	codec := NewDefaultCodec()
	payload := []byte{1, 2, 3, 4, 5}

	frame, err := codec.Pack(payload)
	if err != nil {
		t.Fatalf("unexpected pack error: %v", err)
	}

	if len(frame) != len(payload)+frameHeaderLength {
		t.Fatalf("unexpected frame length: got %d", len(frame))
	}

	unpacked, err := codec.Unpack(frame)
	if err != nil {
		t.Fatalf("unexpected unpack error: %v", err)
	}

	if len(unpacked) != len(payload) {
		t.Fatalf("unexpected unpacked length: got %d", len(unpacked))
	}
}

func TestDefaultCodecParseMessage(t *testing.T) {
	codec := NewDefaultCodec()
	plain := make([]byte, 5+3)
	binary.LittleEndian.PutUint16(plain[2:4], 28)
	plain[4] = 0
	copy(plain[5:], []byte{9, 8, 7})

	msg, err := codec.ParseMessage(plain)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if msg.Command != 28 {
		t.Fatalf("unexpected command: got %d", msg.Command)
	}

	if msg.Code != 0 {
		t.Fatalf("unexpected code: got %d", msg.Code)
	}

	if len(msg.Body) != 3 {
		t.Fatalf("unexpected body length: got %d", len(msg.Body))
	}
}

func TestDefaultCodecRejectsShortMessage(t *testing.T) {
	codec := NewDefaultCodec()
	if _, err := codec.ParseMessage([]byte{1, 2, 3}); err == nil {
		t.Fatalf("expected error for short message")
	}
}
