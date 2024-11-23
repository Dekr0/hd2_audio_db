package main

import (
	"encoding/binary"
	"encoding/xml"
)

const (
	TYPE_CAKMediaIndex = 0x0
	TYPE_CAKHirearchy = 0x1
	TYPE_CAKSOUND = 0x10
	TYPE_CAKRANSEQCNTR = 0x11
)

type CAkObjElement struct {
	Name string
}

type CAkFldElement struct {
	XMLName xml.Name `xml:"fld" json:"XMLName"`
	Name string `xml:"na,attr" json:"Name"`
	Type string `xml:"ty,attr" json:"Type"`
	Value string `xml:"va,attr" json:"Value"`
}

type CAkChildrenElement struct {
	XMLName xml.Name `xml:"obj" json:"XMLName"`
	Nodes []*CAkFldElement `xml:"fld" json:"Nodes"`
}

type CAkWwiseBank struct {
	MediaIndex *CAkMediaIndex `json:"MediaIndex"`
	Hirearchy *CAkHirearchy `json:"Hirearchy"`
}

type CAkMediaIndex struct {
	Count uint32 `json:"Count"`
}

type CAkObject interface {
	GetULID() uint32
	Marshal() []byte /** This is unused. Experimental */
}

/**
 * Size                  uint32
 * CAkObjectsCount       uint32
 * CAkObjects            sizeof(uint32) * CAkObjectsCount
 * ReferencedSoundsCount uint16
 * ReferencedSounds      sizeof(uint32) * ReferencedSoundsCount
 * SoundsCount           uint16
 * Sounds                CAKSOUND_SIZE * SoundsCount
 * RanSeqCntrsCount      uint16
 */
type CAkHirearchy struct {
	CAkObjects map[uint32]CAkObject `json:"CAkObj"`
	ReferencedSounds map[uint32]*CAkSound `json:"CrossSharedSounds"`
	Sounds map[uint32]*CAkSound `json:"Sounds"`
	RanSeqCntrs map[uint32]*CAkRanSeqCntr `json:"RanSeqCntrs"`
}

func (h *CAkHirearchy) Marshal() []byte {
	size := 
		4 +
		4 +
		4 * len(h.CAkObjects) +
		2 +
		4 * len(h.ReferencedSounds) +
		2 +
		CAKSOUND_SIZE * len(h.Sounds) +
		2
	
	ranSeqCntrsBuf := []byte{}
	for _, r := range h.RanSeqCntrs {
		ranSeqCntrsBuf = append(ranSeqCntrsBuf, r.Marshal()...)
	}
	
	size += len(ranSeqCntrsBuf)

	buf := []byte{}

	buf = binary.LittleEndian.AppendUint32(buf, uint32(size))

	buf = binary.LittleEndian.AppendUint32(buf, uint32(len(h.CAkObjects)))
	for ULID := range h.CAkObjects {
		buf = binary.LittleEndian.AppendUint32(buf, ULID)
	}

	buf = binary.LittleEndian.AppendUint16(buf, uint16(len(h.ReferencedSounds)))
	for ULID := range h.ReferencedSounds {
		buf = binary.LittleEndian.AppendUint32(buf, ULID)
	}

	buf = binary.LittleEndian.AppendUint16(buf, uint16(len(h.Sounds)))
	for _, s := range h.Sounds {
		buf = append(buf, s.Marshal()...)
	}

	buf = binary.LittleEndian.AppendUint16(buf, uint16(len(h.RanSeqCntrs)))
	buf = append(buf, ranSeqCntrsBuf...)

	return buf
}


/**
 * Size           uint32
 * DirectParentID uint32 
 * CAkSoundsCount uint16
 * CAkSounds      CAKSOUND_SIZE * CAkSoundsCount
 * ULID           uint32
 *
 * Total Size     
 * */
type CAkRanSeqCntr struct {
	DirectParentID uint32 `json:"DirectParentID"`
	CAkSounds map[uint32]*CAkSound `json:"CAkSounds"`
	ULID uint32 `json:"ULID"`
}

func (r *CAkRanSeqCntr) Marshal() []byte {
	buf := []byte{}

	size := uint32(
		4 + 
		4 + 
		2 + 
		CAKSOUND_SIZE * len(r.CAkSounds) + 
		4)

	buf = binary.LittleEndian.AppendUint32(buf, size)
	buf = binary.LittleEndian.AppendUint32(buf, r.DirectParentID)
	buf = binary.LittleEndian.AppendUint16(buf, uint16(len(r.CAkSounds)))

	for _, s := range r.CAkSounds {
		buf = append(buf, s.Marshal()...)
	}

	buf = binary.LittleEndian.AppendUint32(buf, r.ULID)

	return buf
}

func (r *CAkRanSeqCntr) GetULID() uint32 {
	return r.ULID
}


/**
 * DirectParentID uint32
 * SourceID       uint32
 * ULID           uint32
 * Foreign        uint8
 *
 * Total Size     13 bytes
 * */
const CAKSOUND_SIZE = 13

type CAkSound struct {
	DirectParentID uint32 `json:"DirectParentID"`
	SourceID uint32 `json:"SourceID"`
	ULID uint32 `json:"ULID"`
	Foreign bool `json:"CrossShared"`
}

func (s *CAkSound) GetULID() uint32 {
	return s.ULID
}

func (s *CAkSound) Marshal() []byte {
	buf := []byte{}
	buf = binary.LittleEndian.AppendUint32(buf, s.DirectParentID)
	buf = binary.LittleEndian.AppendUint32(buf, s.SourceID)
	buf = binary.LittleEndian.AppendUint32(buf, s.ULID)
	if s.Foreign {
		buf = append(buf, 1)
	} else {
		buf = append(buf, 0)
	}
	return buf
}
