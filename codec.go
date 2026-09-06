package mt5

import (
	"encoding/binary"

	"github.com/kahono0/mt5/internal/serl"
)

const frameHeaderLength = 8

// DefaultCodec implements the [Codec] interface using the standard MT5
// binary framing and command layout.
type DefaultCodec struct{}

// NewDefaultCodec returns a new [DefaultCodec].
func NewDefaultCodec() *DefaultCodec {
	return &DefaultCodec{}
}

// EncodeCommand prepends a 4-byte command header (2 random bytes + uint16
// command) to the payload.
func (c *DefaultCodec) EncodeCommand(command uint16, body []byte) ([]byte, error) {
	return serl.Command(command, body), nil
}

// Pack wraps a payload in an 8-byte length-prefixed frame.
func (c *DefaultCodec) Pack(payload []byte) ([]byte, error) {
	frame := make([]byte, frameHeaderLength+len(payload))
	frame[0] = byte(len(payload) >> 24)
	frame[1] = byte(len(payload) >> 16)
	frame[2] = byte(len(payload) >> 8)
	frame[3] = byte(len(payload))
	copy(frame[frameHeaderLength:], payload)
	return frame, nil
}

// Unpack strips the 8-byte frame header, returning the inner payload.
func (c *DefaultCodec) Unpack(frame []byte) ([]byte, error) {
	if len(frame) <= frameHeaderLength {
		return nil, ErrMessageTooShort
	}
	return frame[frameHeaderLength:], nil
}

// ParseMessage interprets a decrypted payload as a [Message].
func (c *DefaultCodec) ParseMessage(plain []byte) (Message, error) {
	if len(plain) < 5 {
		return Message{}, ErrMessageTooShort
	}

	msg := Message{
		Command: binary.LittleEndian.Uint16(plain[2:4]),
		Code:    plain[4],
		Body:    plain[5:],
		Raw:     plain,
	}

	return msg, nil
}
