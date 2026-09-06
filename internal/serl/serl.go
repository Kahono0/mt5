package serl

import (
	"crypto/rand"
	"encoding/binary"
)

// Command prepends a 4-byte command header (2 random bytes + uint16
// command ID) to the payload.
func Command(cmd uint16, payload []byte) []byte {
	length := 4
	if payload != nil {
		length += len(payload)
	}

	out := make([]byte, length)
	if _, err := rand.Read(out[0:2]); err != nil {
		panic("failed to generate random bytes")
	}

	binary.LittleEndian.PutUint16(out[2:4], cmd)

	if payload != nil {
		copy(out[4:], payload)
	}

	return out
}
