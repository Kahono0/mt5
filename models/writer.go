package models

import (
	"encoding/binary"
	"fmt"
	"math"
)

type recordWriter struct {
	buf    []byte
	offset int
}

func newRecordWriter(size int) *recordWriter {
	return &recordWriter{
		buf:    make([]byte, size),
		offset: 0,
	}
}

func (w *recordWriter) writeUint8(value uint8) {
	w.buf[w.offset] = value
	w.offset++
}

func (w *recordWriter) writeInt8(value int8) {
	w.writeUint8(uint8(value))
}

func (w *recordWriter) writeUint16(value uint16) {
	binary.LittleEndian.PutUint16(w.buf[w.offset:w.offset+2], value)
	w.offset += 2
}

func (w *recordWriter) writeInt16(value int16) {
	w.writeUint16(uint16(value))
}

func (w *recordWriter) writeUint32(value uint32) {
	binary.LittleEndian.PutUint32(w.buf[w.offset:w.offset+4], value)
	w.offset += 4
}

func (w *recordWriter) writeInt32(value int32) {
	w.writeUint32(uint32(value))
}

func (w *recordWriter) writeUint64(value uint64) {
	binary.LittleEndian.PutUint64(w.buf[w.offset:w.offset+8], value)
	w.offset += 8
}

func (w *recordWriter) writeInt64(value int64) {
	w.writeUint64(uint64(value))
}

func (w *recordWriter) writeFloat64(value float64) {
	w.writeUint64(math.Float64bits(value))
}

func (w *recordWriter) writeUTF16(value string, length int) {
	utf16Bytes := []byte{}
	for _, r := range value {
		utf16Bytes = append(utf16Bytes, byte(r), byte(r>>8))
	}
	copy(w.buf[w.offset:w.offset+length], utf16Bytes[:min(length, len(utf16Bytes))])
	w.offset += length
}

func (w *recordWriter) writeBytes(value []byte, length int) {
	copy(w.buf[w.offset:w.offset+length], value[:min(length, len(value))])
	w.offset += length
}

func (w *recordWriter) finish() ([]byte, error) {
	if w.offset != len(w.buf) {
		return nil, fmt.Errorf("record size mismatch: expected %d bytes, wrote %d bytes", len(w.buf), w.offset)
	}
	return w.buf, nil
}
