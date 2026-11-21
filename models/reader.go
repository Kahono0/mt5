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
