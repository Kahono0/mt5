package mthelpers

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"strings"
)

// ErrInvalidHexLength is returned when a hex string has odd length.
var ErrInvalidHexLength = errors.New("invalid hex string length")

// HexToBuffer converts a hex string into a byte slice.
func HexToBuffer(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, ErrInvalidHexLength
	}
	return hex.DecodeString(s)
}

// Obfuscation shifts each character code by -1, with special cases for
// control characters 28 → '&' and 23 → '!'.
func Obfuscation(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case 28:
			sb.WriteRune('&')
		case 23:
			sb.WriteRune('!')
		default:
			sb.WriteRune(r - 1)
		}
	}
	return sb.String()
}

// DeviceIDFromString hashes a device ID string into a 16-byte identifier
// using SHA-1.
func DeviceIDFromString(idStr string) []byte {
	hash := sha1.Sum([]byte(idStr))
	return hash[:16]
}
