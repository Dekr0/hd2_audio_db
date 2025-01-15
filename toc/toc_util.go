package toc

import (
	"errors"
	"os"
)

type StingRayAssetReader struct {
    file os.File
    head int64
}

func (r *StingRayAssetReader) read(dest []byte) (int64, error) {
	head := r.head

    n, err := r.file.Read(dest)
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
func (r *StingRayAssetReader) readUint8() (uint8, int64, error) {
	head := r.head
    buf := make([]byte, 1, 1)
    nread, err := r.file.Read(buf)
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

func (r *StingRayAssetReader) readInt8() (int8, int64, error) {
	head := r.head
    buf := make([]byte, 1, 1)
    nread, err := r.file.Read(buf)
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

func (r *StingRayAssetReader) readUint16() (uint16, int64, error) {
	head := r.head
    buf := make([]byte, 2, 2)
    nread, err := r.file.Read(buf)
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

func (r *StingRayAssetReader) readInt16() (int16, int64, error) {
	head := r.head
    buf := make([]byte, 2, 2)
    nread , err := r.file.Read(buf)
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

func (r *StingRayAssetReader) readUint32() (uint32, int64, error) {
	head := r.head
    buf := make([]byte, 4, 4)
    nread , err := r.file.Read(buf)
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

func (r *StingRayAssetReader) readInt32() (int32, int64, error) {
	head := r.head
    buf := make([]byte, 4, 4)
    nread, err := r.file.Read(buf)
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

func (r *StingRayAssetReader) readUint64() (uint64, int64, error) {
	head := r.head
    buf := make([]byte, 8, 8)
    nread, err := r.file.Read(buf)
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

func (r *StingRayAssetReader) readInt64() (int64, error) {
    buf := make([]byte, 8, 8)
    nread, err := r.file.Read(buf)
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

func (r *StingRayAssetReader) absoluteSeek(offset int64) error {
    var err error = nil

    r.head, err = r.file.Seek(offset, 0)

    return err
}

func (r *StingRayAssetReader) relativeSeek(offset int64) error {
    var err error = nil

    r.head, err = r.file.Seek(offset, 1)

    return err
}

func (r *StingRayAssetReader) syncHead() error {
    return r.relativeSeek(0)
}
