package main

import (
	"encoding/binary"
	"encoding/xml"
)

const (
	TYPE_CAKMediaIndex = 0x0
	TYPE_CAKHierarchy = 0x1
	TYPE_CAKSOUND = 0x10
	TYPE_CAKRANSEQCNTR = 0x11
)

var HIERARCHY_TYPE_NAME []string = []string{
    "Sound",
    "Random / Sequence Container",
}

type CAkObjectElement struct {
    XMLName xml.Name `xml:"object" json:"XMLName"`
    Name string `xml:"name,attr" json:"Name"`
    Index string `xml:"index,attr" json:"Index"`
}

type CAkFieldElement struct {
	XMLName xml.Name `xml:"field" json:"XMLName"`
	Name string `xml:"name,attr" json:"Name"`
	Type string `xml:"type,attr" json:"Type"`
	Value string `xml:"value,attr" json:"Value"`
    ValueF string `xml:"valuefmt,attr" json:"ValueF"`
}

type CAkChildrenElement struct {
	XMLName xml.Name `xml:"object" json:"XMLName"`
	Nodes []*CAkFieldElement `xml:"field" json:"Nodes"`
}

type CAkWwiseBank struct {
	MediaIndex *CAkMediaIndex `json:"MediaIndex"`
	Hierarchy *CAkHierarchy `json:"Hierarchy"`
}

type CAkMediaIndex struct {
	Count uint32 `json:"Count"`
}

type CAkObject interface {
    getDirectParentID() uint32 
	getObjectULID() uint32
    getType() string
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
type CAkHierarchy struct {
	CAkObjects map[uint32]CAkObject `json:"CAkObj"`
	ReferencedSounds map[uint32]*CAkSound `json:"CrossSharedSounds"` // Potentially contain nil Sound object
	Sounds map[uint32]*CAkSound `json:"Sounds"`
	RanSeqCntrs map[uint32]*CAkRanSeqCntr `json:"RanSeqCntrs"`
}

// Unused
func (h *CAkHierarchy) Marshal() []byte {
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
    DirectParentObjectULID uint32 `json:"DirectParentObjectULID"`
	ObjectULID uint32 `json:"ObjectULID"`
    ObjectIndex int32 `json:"ObjectIndex"`
    CAkSounds map[uint32]*CAkSound `json:"CAkSounds"`
}

func (r *CAkRanSeqCntr) getDirectParentID() uint32 {
    return r.DirectParentObjectULID
}

func (r *CAkRanSeqCntr) getObjectULID() uint32 {
	return r.ObjectULID
}

func (r *CAkRanSeqCntr) getType() string {
    return HIERARCHY_TYPE_NAME[1]
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
	buf = binary.LittleEndian.AppendUint32(buf, r.DirectParentObjectULID)
	buf = binary.LittleEndian.AppendUint16(buf, uint16(len(r.CAkSounds)))

	for _, s := range r.CAkSounds {
		buf = append(buf, s.Marshal()...)
	}

	buf = binary.LittleEndian.AppendUint32(buf, r.ObjectULID)

	return buf
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
    Foreign bool `json:"Foregin"`
	DirectParentObjectULID uint32 `json:"DirectParentID"`
    ObjectULID uint32 `json:"ObjectULID"`
    ObjectIndex int32 `json:"ObjectIndex"`
	ShortIDs map[uint32]Empty `json:"SourceShortID"`
}

func (s *CAkSound) getDirectParentID() uint32 {
    return s.DirectParentObjectULID
}

func (s *CAkSound) getObjectULID() uint32 {
	return s.ObjectULID
}

func (s *CAkSound) getType() string {
    return HIERARCHY_TYPE_NAME[0]
}

// Unused
func (s *CAkSound) Marshal() []byte {
	buf := []byte{}
	buf = binary.LittleEndian.AppendUint32(buf, s.DirectParentObjectULID)
	// buf = binary.LittleEndian.AppendUint32(buf, s.ShortID)
	buf = binary.LittleEndian.AppendUint32(buf, s.ObjectULID)
	if s.Foreign {
		buf = append(buf, 1)
	} else {
		buf = append(buf, 0)
	}
	return buf
}
