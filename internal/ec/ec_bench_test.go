package ec

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

var utf16BenchmarkResult string

func BenchmarkParseUTF16(b *testing.B) {
	cases := []struct {
		name string
		text string
	}{
		{name: "ascii", text: "EURUSD-1234567890"},
		{name: "bmp", text: "café 東京"},
		{name: "surrogate_pair", text: "trade 🚀"},
	}

	for _, tc := range cases {
		buf := encodeUTF16BenchmarkInput(tc.text, 64)
		b.Run(tc.name+"/builder", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				utf16BenchmarkResult = parseUTF16(buf)
			}
		})
		b.Run(tc.name+"/stdlib", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				utf16BenchmarkResult = parseUTF16Std(buf)
			}
		})
	}
}

func parseUTF16Std(buf []byte) string {
	units := bytesToUint16Slice(buf)
	for i, unit := range units {
		if unit == 0 {
			units = units[:i]
			break
		}
	}
	return string(utf16.Decode(units))
}

func encodeUTF16BenchmarkInput(text string, byteLength int) []byte {
	units := utf16.Encode([]rune(text))
	buf := make([]byte, byteLength)
	for i, unit := range units {
		if i*2+1 >= len(buf) {
			break
		}
		binary.LittleEndian.PutUint16(buf[i*2:], unit)
	}
	return buf
}
