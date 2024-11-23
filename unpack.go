package main

import (
	"errors"
	"os"
)

type StingRayAssetReader struct {
    File *os.File
    Head int64
}

func (r *StingRayAssetReader) Read(dest []byte) (int64, error) {
	head := r.Head

    n, err := r.File.Read(dest)
    if err != nil {
        return 0, err
    }
    if n != len(dest) {
        return 0, errors.New("Not enough bytes being read into destination")
    }

    err = r.syncHead()
    
    return head, err
}

/**
 * All the following read / writr functions are in little endian
 */
func (r *StingRayAssetReader) ReadUint8() (uint8, int64, error) {
	head := r.Head
    buf := make([]byte, 1)
    nread, err := r.File.Read(buf)
    if err != nil {
        return 0, 0, err
    }
    if nread != len(buf) {
        return 0, 0, errors.New("Not exactly one byte is being read")
    }

    if err = r.syncHead(); err != nil {
        return 0, 0, err
    }

    return uint8(buf[0]), head, nil
}

func (r *StingRayAssetReader) ReadInt8() (int8, int64, error) {
	head := r.Head
    buf := make([]byte, 1)
    nread, err := r.File.Read(buf)
    if err != nil {
        return 0, 0, err
    }
    if nread != len(buf) {
        return 0, 0, errors.New("Not exactly one byte is being read")
    }

    if err = r.syncHead(); err != nil {
        return 0, 0, err
    }

    return int8(buf[0]), head, nil
}

func (r *StingRayAssetReader) ReadUint16() (uint16, int64, error) {
	head := r.Head
    buf := make([]byte, 2)
    nread, err := r.File.Read(buf)
    if err != nil {
        return 0, 0, err
    }
    if nread != len(buf) {
        return 0, 0, errors.New("Not exactly two bytes are being read") 
    }

    i := uint16(buf[0])
    i |= uint16(buf[1]) << 8

    if err = r.syncHead(); err != nil {
        return 0, 0, err
    }

    return i, head, err
}

func (r *StingRayAssetReader) ReadInt16() (int16, int64, error) {
	head := r.Head
    buf := make([]byte, 2)
    nread , err := r.File.Read(buf)
    if err != nil {
        return 0, 0, err
    }
    if nread != len(buf) {
        return 0, 0, errors.New("Not exactly two bytes are being read") 
    }

    i := int16(buf[0])
    i |= int16(buf[1]) << 8

    if err = r.syncHead(); err != nil {
        return 0, 0, err
    }

    return i, head, err
}

func (r *StingRayAssetReader) ReadUint32() (uint32, int64, error) {
	head := r.Head
    buf := make([]byte, 4)
    nread , err := r.File.Read(buf)
    if err != nil {
        return 0, 0, err
    }
    if nread != len(buf) {
        return 0, 0, errors.New("Not exactly four bytes are being read") 
    }

    i := uint32(buf[0])
    i |= uint32(buf[1]) << 8
    i |= uint32(buf[2]) << 16
    i |= uint32(buf[3]) << 24

    if err = r.syncHead(); err != nil {
        return 0, 0, err
    }

    return i, head, err
}

func (r *StingRayAssetReader) ReadInt32() (int32, int64, error) {
	head := r.Head
    buf := make([]byte, 4)
    nread, err := r.File.Read(buf)
    if err != nil {
        return 0, 0, err
    }
    if nread != len(buf) {
        return 0, 0, errors.New("Not exactly four bytes are being read") 
    }

    i := int32(buf[0])
    i |= int32(buf[1]) << 8
    i |= int32(buf[2]) << 16
    i |= int32(buf[3]) << 24

    if err = r.syncHead(); err != nil {
        return 0, 0, err
    }

    return i, head, err
}

func (r *StingRayAssetReader) ReadUint64() (uint64, int64, error) {
	head := r.Head
    buf := make([]byte, 8)
    nread, err := r.File.Read(buf)
    if err != nil {
        return 0, 0, err
    }
    if nread != len(buf) {
        return 0, 0, errors.New("Not exactly eight bytes are being read") 
    }

    i := uint64(buf[0])
    i |= uint64(buf[1]) << 8
    i |= uint64(buf[2]) << 16
    i |= uint64(buf[3]) << 24
    i |= uint64(buf[4]) << 32
    i |= uint64(buf[5]) << 40
    i |= uint64(buf[6]) << 48
    i |= uint64(buf[7]) << 56

    if err = r.syncHead(); err != nil {
        return 0, 0, err
    }

    return i, head, err
}

func (r *StingRayAssetReader) ReadInt64() (int64, error) {
    buf := make([]byte, 8)
    nread, err := r.File.Read(buf)
    if err != nil {
        return 0, err
    }
    if nread != len(buf) {
        return 0, errors.New("Not exactly eight bytes are being read") 
    }

    i := int64(buf[0])
    i |= int64(buf[1]) << 8
    i |= int64(buf[2]) << 16
    i |= int64(buf[3]) << 24
    i |= int64(buf[4]) << 32
    i |= int64(buf[5]) << 40
    i |= int64(buf[6]) << 48
    i |= int64(buf[7]) << 56

    if err = r.syncHead(); err != nil {
        return 0, err
    }

    return i, err
}

func (r *StingRayAssetReader) AbsoluteSeek(offset int64) error {
    var err error = nil

    r.Head, err = r.File.Seek(offset, 0)

    return err
}

func (r *StingRayAssetReader) RelativeSeek(offset int64) error {
    var err error = nil

    r.Head, err = r.File.Seek(offset, 1)

    return err
}

func (r *StingRayAssetReader) syncHead() error {
    return r.RelativeSeek(0)
}
