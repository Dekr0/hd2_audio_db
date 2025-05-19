package io

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

var ByteOrder = binary.LittleEndian

var InvalidSeek error = errors.New("Invalid Seek")
var NegativeSeek error = errors.New("Negative Seek")

// A brute force extension to bufio.Reader so that it can seek. It's strongly 
// discourage to seek frequently because it reset the buffer in bufio.Reader 
// completely. The ability to seek is to handle some very niche situation.
type Reader struct {
	p uint
	r io.ReadSeeker
	b bufio.Reader
	o binary.ByteOrder
}

func NewReader(r io.ReadSeeker, o binary.ByteOrder) *Reader {
	return &Reader{0, r, *bufio.NewReaderSize(r, 4096), o}
}

func (r *Reader) ByteOrder() binary.ByteOrder {
	return r.o
}

// Tell return the conception position maintained by Reader. It's not necessary 
// the exact position of r.r because of buffering bufio.Reader.
func (r *Reader) Tell() uint {
	return r.p
}

func (r *Reader) AbsSeekUnsafe(p uint) {
	if err := r.AbsSeek(p); err != nil {
		panic(err)
	}
}

// Absolute seek. Perform absolute seek first, then reset bufio.Reader so that 
// it lands at the right position
func (r *Reader) AbsSeek(p uint) error {
	n, err := r.r.Seek(int64(p), io.SeekStart)
	if err != nil {
		return err
	}
	r.b.Reset(r.r)
	r.p = uint(n)
	return nil
}

func (r *Reader) RelSeekUnsafe(p int) {
	if err := r.RelSeek(p); err != nil {
		panic(err)
	}
}

// Relative seek. I cannot use `p` or `r.p` with io.SeekCurrent because 
// bufio.Reader can advance few bytes ahead for buffering. Thus `r.r` is not 
// at the right expected location.
func (r *Reader) RelSeek(p int) error {
	if p + int(r.p) < 0 {
		return NegativeSeek
	}
	n, err := r.r.Seek(int64(p + int(r.p)), io.SeekStart)
	if err != nil {
		return err
	}
	r.b.Reset(r.r)
	r.p = uint(n)
	return nil
}

func (r *Reader) Read(d []byte) (int, error) {
	return r.b.Read(d)
}

func (r *Reader) ReadFullUnsafe(d []byte) {
	err := r.ReadFull(d)
	if err != nil {
		panic(err)
	}
}

func (r *Reader) ReadFull(d []byte) error {
	nread, err := io.ReadFull(&r.b, d)
	if err != nil {
		return err
	}
	r.p += uint(nread)
	return nil
}

// Discourage because it will lead to mandatory heap allocation, i.e., stack 
// escapes to heap.
func (r *Reader) ReadAllUnsafe() []byte {
	b, err := r.ReadAll()
	if err != nil { panic(err) }
	return b
}

// Discourage because it will lead to mandatory heap allocation, i.e., stack 
// escapes to heap.
func (r *Reader) ReadAll() ([]byte, error) {
	b, err := io.ReadAll(&r.b)
	if err != nil { return nil, err }
	r.p += uint(len(b))
	return b, nil
}

func (r *Reader) U8Unsafe() uint8 {
	v, err := r.U8()
	if err != nil { panic(err) }
	return v
}

func (r *Reader) U8() (uint8, error) {
	var v uint8
	err := binary.Read(&r.b, r.o, &v)
	r.p += 1
	return v, err
}


func (r *Reader) I8Unsafe() int8 {
	v, err := r.I8()
	if err != nil { panic(err) }
	return v
}

func (r *Reader) I8() (int8, error) {
	var v int8
	err := binary.Read(&r.b, r.o, &v)
	r.p += 1
	return v, err
}

func (r *Reader) U16Unsafe() uint16 {
	v, err := r.U16()
	if err != nil { panic(err) }
	return v
}

func (r *Reader) U16() (uint16, error) {
	var v uint16
	err := binary.Read(&r.b, r.o, &v)
	r.p += 2
	return v, err
}

func (r *Reader) I16Unsafe() int16 {
	v, err := r.I16()
	if err != nil { panic(err) }
	return v
}

func (r *Reader) I16() (int16, error) {
	var v int16
	err := binary.Read(&r.b, r.o, &v)
	r.p += 2
	return v, err
}

func (r *Reader) U32Unsafe() uint32 {
	v, err := r.U32()
	if err != nil { panic(err) }
	return v
}

func (r *Reader) U32() (uint32, error) {
	var v uint32
	err := binary.Read(&r.b, r.o, &v)
	r.p += 4
	return v, err
}

func (r *Reader) I32Unsafe() int32 {
	v, err := r.I32()
	if err != nil { panic(err) }
	return v
}

func (r *Reader) I32() (int32, error) {
	var v int32
	err := binary.Read(&r.b, r.o, &v)
	r.p += 4
	return v, err
}

func (r *Reader) F32Unsafe() float32 {
	v, err := r.F32()
	if err != nil { panic(err) }
	return v
}

func (r *Reader) F32() (float32, error) {
	var v float32
	err := binary.Read(&r.b, r.o, &v)
	r.p += 4
	return v, err
}

func (r *Reader) U64Unsafe() uint64 {
	v, err := r.U64()
	if err != nil { panic(err) }
	return v
}

func (r *Reader) U64() (uint64, error) {
	var v uint64
	err := binary.Read(&r.b, r.o, &v)
	r.p += 8
	return v, err
}

func (r *Reader) I64Unsafe() int64 {
	v, err := r.I64()
	if err != nil { panic(err) }
	return v
}

func (r *Reader) I64() (int64, error) {
	var v int64
	err := binary.Read(&r.b, r.o, &v)
	r.p += 8
	return v, err
}

type InPlaceReader struct {
	curr uint
	Buff []byte // Escape hatch for accessing this
	o binary.ByteOrder
}

// NOTES: Make sure pass in a slice of a memory region instead a copy of a 
// memory region
func NewInPlaceReader(buff []byte, o binary.ByteOrder) *InPlaceReader {
	buff = append(buff, 0) // pad one byte for EOF
	return &InPlaceReader{curr: 0, Buff: buff, o: o}
}

func (r *InPlaceReader) Cap() uint {
	return uint(len(r.Buff))
}

func (r *InPlaceReader) Len() uint {
	return r.Cap() - r.curr
}

func (r *InPlaceReader) Tell() uint {
	return r.curr
}

func (r *InPlaceReader) AbsSeekUnsafe(j uint) {
	if err := r.AbsSeek(j); err != nil {
		panic(err)
	}
}

func (r *InPlaceReader) AbsSeek(j uint) error {
	if j >= 0 && j < r.Cap() {
		r.curr = j
		return nil
	}
	return InvalidSeek
}

func (r *InPlaceReader) RelSeekUnsafe(j int) {
	if err := r.RelSeek(j); err != nil {
		panic(err)
	}
}

func (r *InPlaceReader) RelSeek(j int) error {
	if j < 0 {
		flip := uint(-j)
		if flip > r.curr {
			return InvalidSeek
		}
		r.curr -= flip 
	} else {
		j := uint(j)
		if j + r.curr >= r.Cap() {
			return InvalidSeek
		}
		r.curr += j
	}
	return nil
}

func (r *InPlaceReader) NewInPlaceReader(s uint) (*InPlaceReader, error) {
	if s > r.Len() {
		return nil, io.ErrShortBuffer
	}
	nr := NewInPlaceReader(r.Buff[r.curr:r.curr + s], r.o)
	r.curr += s
	return nr, nil
}

func (r *InPlaceReader) NewInPlaceReaderOffset(offset uint, s uint) {}

func (r *InPlaceReader) NewInPlaceReaderUnsafe(s uint) (*InPlaceReader) {
	nr, err := r.NewInPlaceReader(s)
	if err != nil {
		panic(err)
	}
	return nr
}

func (r *InPlaceReader) U8Unsafe() uint8 {
	v, err := r.U8()
	if err != nil { panic(err) }
	return v
}

func (r *InPlaceReader) U8() (uint8, error) {
	if r.Len() < 1 {
		return 0, io.ErrShortBuffer
	}
	var v uint8
	_, err := binary.Decode(r.Buff[r.curr:r.curr + 1], r.o, &v)
	if err == nil {
		r.curr += 1
	} 
	return v, err
}


func (r *InPlaceReader) I8Unsafe() int8 {
	v, err := r.I8()
	if err != nil { panic(err) }
	return v
}

func (r *InPlaceReader) I8() (int8, error) {
	if r.Len() < 1 {
		return 0, io.ErrShortBuffer
	}
	var v int8
	
	_, err := binary.Decode(r.Buff[r.curr:r.curr + 1], r.o, &v)
	if err == nil {
		r.curr += 1
	} 
	return v, err
}

func (r *InPlaceReader) U16Unsafe() uint16 {
	v, err := r.U16()
	if err != nil { panic(err) }
	return v
}

func (r *InPlaceReader) U16() (uint16, error) {
	if r.Len() < 2 {
		return 0, io.ErrShortBuffer
	}
	var v uint16
	
	_, err := binary.Decode(r.Buff[r.curr:r.curr + 2], r.o, &v)
	if err == nil {
		r.curr += 2
	} 
	return v, err
}

func (r *InPlaceReader) I16Unsafe() int16 {
	v, err := r.I16()
	if err != nil { panic(err) }
	return v
}

func (r *InPlaceReader) I16() (int16, error) {
	if r.Len() < 2 {
		return 0, io.ErrShortBuffer
	}
	var v int16
	
	_, err := binary.Decode(r.Buff[r.curr:r.curr + 2], r.o, &v)
	if err == nil {
		r.curr += 2
	} 
	return v, err
}

func (r *InPlaceReader) U32Unsafe() uint32 {
	v, err := r.U32()
	if err != nil { panic(err) }
	return v
}

func (r *InPlaceReader) U32() (uint32, error) {
	if r.Len() < 4 {
		return 0, io.ErrShortBuffer
	}
	var v uint32
	
	_, err := binary.Decode(r.Buff[r.curr:r.curr + 4], r.o, &v)
	if err == nil {
		r.curr += 4
	} 
	return v, err
}

func (r *InPlaceReader) I32Unsafe() int32 {
	v, err := r.I32()
	if err != nil { panic(err) }
	return v
}

func (r *InPlaceReader) I32() (int32, error) {
	if r.Len() < 4 {
		return 0, io.ErrShortBuffer
	}
	var v int32
	
	_, err := binary.Decode(r.Buff[r.curr:r.curr + 4], r.o, &v)
	if err == nil {
		r.curr += 4
	} 
	return v, err
}

func (r *InPlaceReader) F32Unsafe() float32 {
	v, err := r.F32()
	if err != nil { panic(err) }
	return v
}

func (r *InPlaceReader) F32() (float32, error) {
	if r.Len() < 4 {
		return 0, io.ErrShortBuffer
	}
	var v float32
	
	_, err := binary.Decode(r.Buff[r.curr:r.curr + 4], r.o, &v)
	if err == nil {
		r.curr += 4
	} 
	return v, err
}

func (r *InPlaceReader) FourCCNoCopyUnsafe() ([]byte) {
	b, err := r.FourCCNoCopy()
	if err != nil { panic(err) }
	return b
}

func (r *InPlaceReader) FourCCNoCopy() ([]byte, error) {
	if r.Len() < 4 {
		return nil, io.ErrShortBuffer
	}
	b := r.Buff[r.curr:r.curr + 4]
	r.curr += 4
	return b, nil
}

func (r *InPlaceReader) FourCCUnsafe() []byte {
	b, err := r.FourCC()
	if err != nil {
		panic(err)
	}
	return b
}

func (r *InPlaceReader) FourCC() ([]byte, error) {
	if r.Len() < 4 {
		return nil, io.ErrShortBuffer
	}
	
	b := bytes.Clone(r.Buff[r.curr:r.curr + 4])
	r.curr += 4
	return b, nil
}

func (r *InPlaceReader) U64Unsafe() uint64 {
	v, err := r.U64()
	if err != nil { panic(err) }
	return v
}

func (r *InPlaceReader) U64() (uint64, error) {
	if r.Len() < 8 {
		return 0, io.ErrShortBuffer
	}
	var v uint64
	
	_, err := binary.Decode(r.Buff[r.curr:r.curr + 8], r.o, &v)
	if err == nil {
		r.curr += 8
	} 
	return v, err
}

func (r *InPlaceReader) I64Unsafe() int64 {
	v, err := r.I64()
	if err != nil { panic(err) }
	return v
}

func (r *InPlaceReader) I64() (int64, error) {
	if r.Len() < 8 {
		return 0, io.ErrShortBuffer
	}
	var v int64
	
	_, err := binary.Decode(r.Buff[r.curr:r.curr + 8], r.o, &v)
	if err == nil {
		r.curr += 8
	} 
	return v, err
}

func (r *InPlaceReader) ReadNoCopyUnsafe(n uint) []byte {
	b, err := r.ReadNoCopy(n)
	if err != nil {
		panic(err)
	}
	return b
}

func (r *InPlaceReader) ReadNoCopy(n uint) ([]byte, error) {
	if r.Len() < n {
		return nil, io.ErrShortBuffer
	}
	b := r.Buff[r.curr:r.curr + n]
	r.curr += n
	return b, nil
}

func (r *InPlaceReader) ReadUnsafe(n uint) []byte {
	b, err := r.Read(n)
	if err != nil {
		panic(err)
	}
	return b
}

func (r *InPlaceReader) Read(n uint) ([]byte, error) {
	if r.Len() < n {
		return nil, io.ErrShortBuffer
	}
	
	b := bytes.Clone(r.Buff[r.curr:r.curr + n])
	r.curr += n
	return b, nil
}

func (r *InPlaceReader) ExhaustNoCopy() ([]byte) {
	return r.Buff[r.curr:]
}
