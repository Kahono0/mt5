package models

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/kahono0/mt5/internal/ec"
)

type recordReader struct {
	buf    []byte
	offset int
	err    error
}

func (r *recordReader) take(size int) []byte {
	if r.err != nil {
		return nil
	}
	if size < 0 || size > len(r.buf)-r.offset {
		r.err = fmt.Errorf("record too short at byte %d: need %d bytes", r.offset, size)
		return nil
	}
	value := r.buf[r.offset : r.offset+size]
	r.offset += size
	return value
}

func (r *recordReader) readUint8() uint8 {
	buf := r.take(1)
	if buf == nil {
		return 0
	}
	return buf[0]
}

func (r *recordReader) readInt8() int8 { return int8(r.readUint8()) }

func (r *recordReader) readUint16() uint16 {
	buf := r.take(2)
	if buf == nil {
		return 0
	}
	return binary.LittleEndian.Uint16(buf)
}

func (r *recordReader) readInt16() int16 { return int16(r.readUint16()) }

func (r *recordReader) readUint32() uint32 {
	buf := r.take(4)
	if buf == nil {
		return 0
	}
	return binary.LittleEndian.Uint32(buf)
}

func (r *recordReader) readInt32() int32 { return int32(r.readUint32()) }

func (r *recordReader) readUint64() uint64 {
	buf := r.take(8)
	if buf == nil {
		return 0
	}
	return binary.LittleEndian.Uint64(buf)
}

func (r *recordReader) readFloat64() float64 {
	return math.Float64frombits(r.readUint64())
}

func (r *recordReader) readFloat64Array(size int) []float64 {
	buf := r.take(size)
	if buf == nil {
		return nil
	}
	result := make([]float64, size/8)
	for i := range result {
		result[i] = math.Float64frombits(binary.LittleEndian.Uint64(buf[i*8:]))
	}
	return result
}

func (r *recordReader) readString(size int) string {
	return ec.DecodeUTF16(r.take(size))
}

func (r *recordReader) finish() error {
	if r.err != nil {
		return r.err
	}
	if r.offset != len(r.buf) {
		return fmt.Errorf("record decoder consumed %d of %d bytes", r.offset, len(r.buf))
	}
	return nil
}

func readRecordCount(body []byte, offset, recordSize int, name string) (int, int, error) {
	if recordSize <= 0 {
		return 0, offset, fmt.Errorf("invalid %s record size %d", name, recordSize)
	}
	if offset < 0 || offset > len(body) || len(body)-offset < 4 {
		return 0, offset, fmt.Errorf("buffer too short for %s count", name)
	}
	count := binary.LittleEndian.Uint32(body[offset : offset+4])
	offset += 4
	remainingRecords := (len(body) - offset) / recordSize
	if uint64(count) > uint64(remainingRecords) {
		return 0, offset, fmt.Errorf("buffer too short for %s: count %d exceeds available records %d", name, count, remainingRecords)
	}
	return int(count), offset, nil
}
