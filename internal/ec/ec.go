package ec

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode/utf16"
)

// Field mirrors a JS field: [propType, propValue, propLength, propLengthType]
type Field struct {
	PropType       int
	PropValue      interface{}
	PropLength     int
	PropLengthType int
	Parser         func([]byte, int) (interface{}, error) // optional custom parser for PropType 12
	InnerFields    []Field                                // optional inner fields for PropType 12/13
	FromValues     func([]interface{}) any                // optional function to convert parsed values to a struct
}

func GetSeriesSize(fields []Field) (int, error) {
	size := 0

	for _, f := range fields {
		switch f.PropType {
		case 1, 4:
			size += 1
		case 2, 5:
			size += 2
		case 3, 6, 7:
			size += 4
		case 8, 9, 17, 18:
			size += 8
		case 10:
			if f.PropLength > 0 {
				size += f.PropLength
			} else if s, ok := f.PropValue.(string); ok {
				size += len(s)
			} else {
				return 0, fmt.Errorf("expected string for PropType 10, got %T", f.PropValue)
			}
		case 11:
			if f.PropLength > 0 {
				size += f.PropLength
			} else if s, ok := f.PropValue.(string); ok {
				size += 2 * len(s)
			} else {
				return 0, fmt.Errorf("expected string for PropType 11, got %T", f.PropValue)
			}
		case 12:
			if f.PropLength > 0 {
				size += f.PropLength
			} else if b, ok := f.PropValue.([]byte); ok {
				size += len(b)
			} else {
				return 0, fmt.Errorf("expected []byte for PropType 12, got %T", f.PropValue)
			}
		default:
			return 0, fmt.Errorf("unsupported PropType: %d", f.PropType)
		}
	}

	return size, nil
}

// ---------- Helpers ----------
func getSeriesSize(fields []Field) int {
	size, _ := GetSeriesSize(fields)
	return size
}

func setCharString(buf []byte, str string, offset int) {
	for i := 0; i < len(str); i++ {
		buf[offset+i] = str[i]
	}
}

// ---------- Serialize ----------
func Serialize(fields []Field) ([]byte, error) {
	// calculate total size
	size := getSeriesSize(fields)
	buf := make([]byte, size)
	offset := 0

	for _, f := range fields {
		switch f.PropType {
		case 1: // Int8
			v, ok := toInt64(f.PropValue)
			if !ok {
				return nil, errors.New("expected numeric value for PropType 1")
			}
			buf[offset] = byte(int8(v))
			offset += 1
		case 2: // Int16
			v, ok := toInt64(f.PropValue)
			if !ok {
				return nil, errors.New("expected numeric value for PropType 2")
			}
			binary.LittleEndian.PutUint16(buf[offset:], uint16(int16(v)))
			offset += 2
		case 3: // Int32
			v, ok := toInt64(f.PropValue)
			if !ok {
				return nil, errors.New("expected numeric value for PropType 3")
			}
			binary.LittleEndian.PutUint32(buf[offset:], uint32(int32(v)))
			offset += 4
		case 4: // Uint8
			v, ok := toUint64(f.PropValue)
			if !ok {
				return nil, errors.New("expected numeric value for PropType 4")
			}
			buf[offset] = byte(v)
			offset += 1
		case 5: // Uint16
			v, ok := toUint64(f.PropValue)
			if !ok {
				return nil, errors.New("expected numeric value for PropType 5")
			}
			binary.LittleEndian.PutUint16(buf[offset:], uint16(v))
			offset += 2
		case 6: // Uint32
			v, ok := toUint64(f.PropValue)
			if !ok {
				return nil, errors.New("expected numeric value for PropType 6")
			}
			binary.LittleEndian.PutUint32(buf[offset:], uint32(v))
			offset += 4
		case 7: // Float32
			v, ok := toFloat64(f.PropValue)
			if !ok {
				return nil, errors.New("expected numeric value for PropType 7")
			}
			binary.LittleEndian.PutUint32(buf[offset:], math.Float32bits(float32(v)))
			offset += 4
		case 8: // Float64
			v, ok := toFloat64(f.PropValue)
			if !ok {
				return nil, errors.New("expected numeric value for PropType 8")
			}
			binary.LittleEndian.PutUint64(buf[offset:], math.Float64bits(v))
			offset += 8
		case 10: // ASCII / byte string, padded to PropLength
			s, ok := f.PropValue.(string)
			if !ok {
				return nil, errors.New("expected string for PropType 10")
			}
			length := f.PropLength
			if length == 0 {
				length = len(s)
			}
			copy(buf[offset:], []byte(s))
			if len(s) < length {
				for i := len(s); i < length; i++ {
					buf[offset+i] = 0
				}
			}
			offset += length
		case 11: // UTF-16 string
			s, ok := f.PropValue.(string)
			if !ok {
				return nil, errors.New("expected string for PropType 11")
			}
			utf16codes := utf16.Encode([]rune(s))
			for i, code := range utf16codes {
				binary.LittleEndian.PutUint16(buf[offset+i*2:], code)
			}
			length := f.PropLength
			if length == 0 {
				length = len(utf16codes) * 2
			}
			// pad if necessary
			if len(utf16codes)*2 < length {
				for i := len(utf16codes) * 2; i < length; i++ {
					buf[offset+i] = 0
				}
			}
			offset += length
		case 12: // []byte / ArrayBuffer
			v, ok := f.PropValue.([]byte)
			if !ok {
				return nil, errors.New("expected []byte for PropType 12")
			}
			copy(buf[offset:], v)
			offset += len(v)
		case 17: // int64
			v, ok := toInt64(f.PropValue)
			if !ok {
				return nil, errors.New("expected int64 for PropType 17")
			}
			binary.LittleEndian.PutUint64(buf[offset:], uint64(v))
			offset += 8
		case 18: // uint64 / BigInt
			v, ok := toUint64(f.PropValue)
			if !ok {
				return nil, errors.New("expected uint64 for PropType 18")
			}
			binary.LittleEndian.PutUint64(buf[offset:], v)
			offset += 8
		default:
			return nil, errors.New("unsupported PropType")
		}
	}

	return buf, nil
}

// helper functions

func toInt64(val interface{}) (int64, bool) {
	switch v := val.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	default:
		return 0, false
	}
}

func toUint64(val interface{}) (uint64, bool) {
	switch v := val.(type) {
	case int:
		return uint64(v), true
	case int8:
		return uint64(v), true
	case int16:
		return uint64(v), true
	case int32:
		return uint64(v), true
	case int64:
		return uint64(v), true
	case uint8:
		return uint64(v), true
	case uint16:
		return uint64(v), true
	case uint32:
		return uint64(v), true
	case uint64:
		return v, true
	default:
		return 0, false
	}
}

func toFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

// Parse parses a byte buffer according to field definitions
func Parse(buf []byte, fields []Field, offset int) ([]interface{}, error) {
	out := make([]interface{}, len(fields))

	for i, f := range fields {
		switch f.PropType {
		case 1:
			out[i] = int8(buf[offset])
			offset += 1
		case 2:
			out[i] = int16(binary.LittleEndian.Uint16(buf[offset:]))
			offset += 2
		case 3:
			out[i] = int32(binary.LittleEndian.Uint32(buf[offset:]))
			offset += 4
		case 4:
			out[i] = buf[offset]
			offset += 1
		case 5:
			out[i] = binary.LittleEndian.Uint16(buf[offset:])
			offset += 2
		case 6:
			out[i] = binary.LittleEndian.Uint32(buf[offset:])
			offset += 4
		case 7:
			out[i] = math.Float32frombits(binary.LittleEndian.Uint32(buf[offset:]))
			offset += 4
		case 8:
			out[i] = math.Float64frombits(binary.LittleEndian.Uint64(buf[offset:]))
			offset += 8
		case 10:
			length := f.PropLength
			if length == 0 {
				length = len(f.PropValue.(string))
			}
			out[i] = getCharString(buf[offset : offset+length])
			offset += length
		case 11:
			//length := f.PropLength
			//if length == 0 {
			//	length = len(f.PropValue.(string)) * 2
			//}
			//out[i] = parseUTF16(buf[offset : offset+length])
			//offset += length

			length := f.PropLength
			if length == 0 {
				length = len(buf) - offset
			}
			if offset+length > len(buf) {
				length = len(buf) - offset
			}
			out[i] = parseUTF16(buf[offset : offset+length])
			offset += length
		case 12:
			// Determine the length
			length := f.PropLength
			if f.PropLengthType != 0 {
				switch f.PropLengthType {
				case 4:
					length = int(buf[offset])
					offset += 1
				case 5:
					length = int(binary.LittleEndian.Uint16(buf[offset:]))
					offset += 2
				case 6:
					length = int(binary.LittleEndian.Uint32(buf[offset:]))
					offset += 4
				default:
					return nil, errors.New("unsupported PropLengthType")
				}
			}

			// Check buffer bounds
			if offset+length > len(buf) {
				return nil, fmt.Errorf(
					"buffer too short for PropType 12, need %d bytes, got %d",
					length,
					len(buf)-offset,
				)
			}

			// Use field parser if provided
			if f.Parser != nil {
				val, err := f.Parser(buf[offset:offset+length], 0)
				if err != nil {
					return nil, fmt.Errorf("PropType 12 parser error: %w", err)
				}
				out[i] = val
			} else if len(f.InnerFields) != 0 {
				val, err := Parse(buf[offset:offset+length], f.InnerFields, 0)
				if err != nil {
					return nil, fmt.Errorf("PropType 12 inner fields parse error: %w", err)
				}
				if f.FromValues != nil {
					out[i] = f.FromValues(val)
				} else {
					out[i] = val
				}
			} else {
				// fallback: raw bytes
				out[i] = buf[offset : offset+length]
			}

			offset += length
		case 13:
			// Determine the length
			length := f.PropLength
			count := 1
			if f.PropLengthType != 0 {
				switch f.PropLengthType {
				case 4:
					count = int(buf[offset])
					offset += 1
				case 5:
					count = int(binary.LittleEndian.Uint16(buf[offset:]))
					offset += 2
				case 6:
					count = int(binary.LittleEndian.Uint32(buf[offset:]))
					offset += 4
				default:
					return nil, errors.New("unsupported PropLengthType")
				}
			}

			totalLength := length * count
			// Check buffer bounds
			if offset+totalLength > len(buf) {
				return nil, fmt.Errorf(
					"buffer too short for PropType 13, need %d bytes, got %d",
					totalLength,
					len(buf)-offset,
				)
			}

			// Use field parser if provided
			if f.Parser != nil {
				outVal := make([]interface{}, count)
				for i := range count {
					val, err := f.Parser(buf[offset:offset+length], 0)
					if err != nil {
						return nil, fmt.Errorf("PropType 13 parser error at index %d: %w", i, err)
					}
					outVal[i] = val
					offset += length
				}
				out[i] = outVal
			} else if len(f.InnerFields) != 0 {
				outVal := make([]interface{}, count)
				for i := range count {
					val, err := Parse(buf[offset:offset+length], f.InnerFields, 0)
					if err != nil {
						return nil, fmt.Errorf("PropType 13 inner fields parse error at index %d: %w", i, err)
					}
					if f.FromValues != nil {
						outVal[i] = f.FromValues(val)
					} else {
						outVal[i] = val
					}
					offset += length
				}
				out[i] = outVal
			} else {
				// fallback: raw bytes
				out[i] = buf[offset : offset+length]

				offset += length
			}

		case 17:
			out[i] = int64(binary.LittleEndian.Uint64(buf[offset:]))
			offset += 8
		case 18:
			out[i] = binary.LittleEndian.Uint64(buf[offset:])
			offset += 8
		default:
			return nil, errors.New("unsupported PropType")
		}
	}

	return out, nil
}

// parseUTF16 converts little-endian UTF-16 bytes to string, stops at first null
func parseUTF16(buf []byte) string {
	// A UTF-16 code unit expands to at most three UTF-8 bytes. A surrogate
	// pair takes four bytes for two code units, so this capacity is sufficient
	// for the entire decoded string and avoids an intermediate []rune.
	var result strings.Builder
	result.Grow((len(buf) / 2) * 3)

	for i := 0; i+1 < len(buf); {
		first := binary.LittleEndian.Uint16(buf[i:])
		if first == 0 {
			break
		}
		i += 2

		r := rune(first)
		if first >= 0xD800 && first <= 0xDBFF {
			if i+1 < len(buf) {
				second := binary.LittleEndian.Uint16(buf[i:])
				if second >= 0xDC00 && second <= 0xDFFF {
					r = utf16.DecodeRune(r, rune(second))
					i += 2
				} else {
					r = '\uFFFD'
				}
			} else {
				r = '\uFFFD'
			}
		} else if first >= 0xDC00 && first <= 0xDFFF {
			r = '\uFFFD'
		}
		result.WriteRune(r)
	}
	return result.String()
}

func bytesToUint16Slice(buf []byte) []uint16 {
	if len(buf)%2 != 0 {
		buf = buf[:len(buf)-1] // trim last byte if odd length
	}
	u16s := make([]uint16, len(buf)/2)
	for i := 0; i < len(u16s); i++ {
		u16s[i] = binary.LittleEndian.Uint16(buf[i*2:])
	}
	return u16s
}

// getCharString reads ASCII / byte string and trims trailing zero bytes
func getCharString(buf []byte) string {
	i := len(buf)
	for i > 0 && buf[i-1] == 0 {
		i--
	}
	return string(buf[:i])
}
