package ec

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"unicode/utf16"
)

// Field describes a single typed value in a binary serialization sequence.
type Field struct {
	PropType   int
	PropValue  interface{}
	PropLength int
}

// Serialize encodes a sequence of fields into a binary byte slice.
func Serialize(fields []Field) ([]byte, error) {
	size, err := seriesSize(fields)
	if err != nil {
		return nil, err
	}
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
			offset++
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
			offset++
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
		case 10: // ASCII string, padded to PropLength
			s, ok := f.PropValue.(string)
			if !ok {
				return nil, errors.New("expected string for PropType 10")
			}
			length := f.PropLength
			if length == 0 {
				length = len(s)
			}
			copy(buf[offset:], []byte(s))
			offset += length
		case 11: // UTF-16LE string, padded to PropLength
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
			offset += length
		case 12: // Raw bytes
			v, ok := f.PropValue.([]byte)
			if !ok {
				return nil, errors.New("expected []byte for PropType 12")
			}
			copy(buf[offset:], v)
			offset += len(v)
		case 17: // Int64
			v, ok := toInt64(f.PropValue)
			if !ok {
				return nil, errors.New("expected int64 for PropType 17")
			}
			binary.LittleEndian.PutUint64(buf[offset:], uint64(v))
			offset += 8
		case 18: // Uint64
			v, ok := toUint64(f.PropValue)
			if !ok {
				return nil, errors.New("expected uint64 for PropType 18")
			}
			binary.LittleEndian.PutUint64(buf[offset:], v)
			offset += 8
		default:
			return nil, fmt.Errorf("unsupported PropType: %d", f.PropType)
		}
	}

	return buf, nil
}

func seriesSize(fields []Field) (int, error) {
	size := 0
	for _, f := range fields {
		switch f.PropType {
		case 1, 4:
			size++
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

func GetSeriesSize(fields []Field) (int, error) {
	return seriesSize(fields)
}

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
