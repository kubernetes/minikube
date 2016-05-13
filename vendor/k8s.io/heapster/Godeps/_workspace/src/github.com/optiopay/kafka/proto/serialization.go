package proto

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var ErrNotEnoughData = errors.New("not enough data")

type decoder struct {
	buf []byte
	r   io.Reader
	err error
}

func NewDecoder(r io.Reader) *decoder {
	return &decoder{
		r:   r,
		buf: make([]byte, 1024),
	}
}

func (d *decoder) DecodeInt8() int8 {
	if d.err != nil {
		return 0
	}
	b := d.buf[:1]
	n, err := io.ReadFull(d.r, b)
	if err != nil {
		d.err = err
		return 0
	}
	if n != 1 {
		d.err = ErrNotEnoughData
		return 0
	}
	return int8(b[0])
}

func (d *decoder) DecodeInt16() int16 {
	if d.err != nil {
		return 0
	}
	b := d.buf[:2]
	n, err := io.ReadFull(d.r, b)
	if err != nil {
		d.err = err
		return 0
	}
	if n != 2 {
		d.err = ErrNotEnoughData
		return 0
	}
	return int16(binary.BigEndian.Uint16(b))
}

func (d *decoder) DecodeInt32() int32 {
	if d.err != nil {
		return 0
	}
	b := d.buf[:4]
	n, err := io.ReadFull(d.r, b)
	if err != nil {
		d.err = err
		return 0
	}
	if n != 4 {
		d.err = ErrNotEnoughData
		return 0
	}
	return int32(binary.BigEndian.Uint32(b))
}

func (d *decoder) DecodeUint32() uint32 {
	if d.err != nil {
		return 0
	}
	b := d.buf[:4]
	n, err := io.ReadFull(d.r, b)
	if err != nil {
		d.err = err
		return 0
	}
	if n != 4 {
		d.err = ErrNotEnoughData
		return 0
	}
	return binary.BigEndian.Uint32(b)
}

func (d *decoder) DecodeInt64() int64 {
	if d.err != nil {
		return 0
	}
	b := d.buf[:8]
	n, err := io.ReadFull(d.r, b)
	if err != nil {
		d.err = err
		return 0
	}
	if n != 8 {
		d.err = ErrNotEnoughData
		return 0
	}
	return int64(binary.BigEndian.Uint64(b))
}

func (d *decoder) DecodeString() string {
	if d.err != nil {
		return ""
	}
	slen := d.DecodeInt16()
	if d.err != nil {
		return ""
	}
	if slen < 1 {
		return ""
	}

	var b []byte
	if int(slen) > len(d.buf) {
		b = make([]byte, slen)
	} else {
		b = d.buf[:int(slen)]
	}
	n, err := io.ReadFull(d.r, b)
	if err != nil {
		d.err = err
		return ""
	}
	if n != int(slen) {
		d.err = ErrNotEnoughData
		return ""
	}
	return string(b)
}

func (d *decoder) DecodeArrayLen() int {
	return int(d.DecodeInt32())
}

func (d *decoder) DecodeBytes() []byte {
	if d.err != nil {
		return nil
	}
	slen := d.DecodeInt32()
	if d.err != nil {
		return nil
	}
	if slen < 1 {
		return nil
	}

	b := make([]byte, slen)
	n, err := io.ReadFull(d.r, b)
	if err != nil {
		d.err = err
		return nil
	}
	if n != int(slen) {
		d.err = ErrNotEnoughData
		return nil
	}
	return b
}

func (d *decoder) Err() error {
	return d.err
}

type encoder struct {
	w   io.Writer
	err error
	buf [8]byte
}

func NewEncoder(w io.Writer) *encoder {
	return &encoder{w: w}
}

func (e *encoder) Encode(value interface{}) {
	if e.err != nil {
		return
	}
	var b []byte

	switch val := value.(type) {
	case int8:
		_, e.err = e.w.Write([]byte{byte(val)})
	case int16:
		b = e.buf[:2]
		binary.BigEndian.PutUint16(b, uint16(val))
	case int32:
		b = e.buf[:4]
		binary.BigEndian.PutUint32(b, uint32(val))
	case int64:
		b = e.buf[:8]
		binary.BigEndian.PutUint64(b, uint64(val))
	case uint16:
		b = e.buf[:2]
		binary.BigEndian.PutUint16(b, val)
	case uint32:
		b = e.buf[:4]
		binary.BigEndian.PutUint32(b, val)
	case uint64:
		b = e.buf[:8]
		binary.BigEndian.PutUint64(b, val)
	case string:
		buf := e.buf[:2]
		binary.BigEndian.PutUint16(buf, uint16(len(val)))
		e.err = writeAll(e.w, buf)
		if e.err == nil {
			e.err = writeAll(e.w, []byte(val))
		}
	case []byte:
		buf := e.buf[:4]

		if val == nil {
			no := int32(-1)
			binary.BigEndian.PutUint32(buf, uint32(no))
			e.err = writeAll(e.w, buf)
			return
		}

		binary.BigEndian.PutUint32(buf, uint32(len(val)))
		e.err = writeAll(e.w, buf)
		if e.err == nil {
			e.err = writeAll(e.w, val)
		}
	case []int32:
		e.EncodeArrayLen(len(val))
		for _, v := range val {
			e.Encode(v)
		}
	default:
		e.err = fmt.Errorf("cannot encode type %T", value)
	}

	if b != nil {
		e.err = writeAll(e.w, b)
		return
	}
}

func (e *encoder) EncodeInt8(val int8) {
	if e.err != nil {
		return
	}

	_, e.err = e.w.Write([]byte{byte(val)})
}

func (e *encoder) EncodeInt16(val int16) {
	if e.err != nil {
		return
	}

	b := e.buf[:2]
	binary.BigEndian.PutUint16(b, uint16(val))
	e.err = writeAll(e.w, b)
}

func (e *encoder) EncodeInt32(val int32) {
	if e.err != nil {
		return
	}

	b := e.buf[:4]
	binary.BigEndian.PutUint32(b, uint32(val))
	e.err = writeAll(e.w, b)
}

func (e *encoder) EncodeInt64(val int64) {
	if e.err != nil {
		return
	}

	b := e.buf[:8]
	binary.BigEndian.PutUint64(b, uint64(val))
	e.err = writeAll(e.w, b)
}

func (e *encoder) EncodeUint32(val uint32) {
	if e.err != nil {
		return
	}

	b := e.buf[:4]
	binary.BigEndian.PutUint32(b, val)
	e.err = writeAll(e.w, b)
}

func (e *encoder) EncodeBytes(val []byte) {
	if e.err != nil {
		return
	}

	buf := e.buf[:4]

	if val == nil {
		no := int32(-1)
		binary.BigEndian.PutUint32(buf, uint32(no))
		e.err = writeAll(e.w, buf)
		return
	}

	binary.BigEndian.PutUint32(buf, uint32(len(val)))
	e.err = writeAll(e.w, buf)
	if e.err == nil {
		e.err = writeAll(e.w, val)
	}
}

func (e *encoder) EncodeString(val string) {
	if e.err != nil {
		return
	}

	buf := e.buf[:2]

	binary.BigEndian.PutUint16(buf, uint16(len(val)))
	e.err = writeAll(e.w, buf)
	if e.err == nil {
		e.err = writeAll(e.w, []byte(val))
	}
}

func (e *encoder) EncodeError(err error) {
	b := e.buf[:2]

	if err == nil {
		binary.BigEndian.PutUint16(b, uint16(0))
		e.err = writeAll(e.w, b)
		return
	}
	kerr, ok := err.(*KafkaError)
	if !ok {
		e.err = fmt.Errorf("cannot encode error of type %T", err)
	}

	binary.BigEndian.PutUint16(b, uint16(kerr.errno))
	e.err = writeAll(e.w, b)
}

func (e *encoder) EncodeArrayLen(length int) {
	e.EncodeInt32(int32(length))
}

func (e *encoder) Err() error {
	return e.err
}

func writeAll(w io.Writer, b []byte) error {
	n, err := w.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return fmt.Errorf("cannot write %d: %d written", len(b), n)
	}
	return nil
}
