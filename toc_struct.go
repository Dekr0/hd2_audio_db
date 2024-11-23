package main

import "encoding/json"

const MAGIC = 4026531857

/*
 * ToC Header type constance
 * */
const (
	TYPE_WWISE_STREAM = 5785811756662211598
	TYPE_WWISE_BANK   = 6006249203084351385
    TYPE_WWISE_DEP    = 12624162998411505776
)

/*
 * WWise Bank Hierarchy Type
 * */

type Number interface {
	int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64
}

type Value[K Number] struct {
	Address int64 `json:"Address"`
	Value K `json:"Value"`
}

type ToCFile struct {
    Magic Value[uint32] `json:"Magic"`
    NumTypes Value[uint32] `json:"NumTypes"`
    NumFiles Value[uint32] `json:"NumFiles"`
    Unknown Value[uint32] `json:"Unknown"`
    Unk4Data [56]Value[uint8] `json:"Unk4Data"`
    ToCEntries []*ToCHeader `json:"ToCEntries,omitEmpty"`
}

type ToCHeader struct {
    FileID Value[uint64] `json:"FileID"`
    TypeID Value[uint64] `json:"TypeID"`
    ToCDataOffset Value[uint64] `json:"ToCDataOffset"`
    StreamFileOffset Value[uint64] `json:"StreamFileOffset"`
    GPUResourceOffset Value[uint64] `json:"GPUResourceOffset"`
    Unknown1 Value[uint64] `json:"Unknown1"`
    Unknown2 Value[uint64] `json:"Unknown2"`
    ToCDataSize Value[uint32] `json:"ToCDataSize"`
    StreamSize Value[uint32] `json:"StreamSize"`
    GPUResourceSize Value[uint32] `json:"GPUResourceSize"`
    Unknown3 Value[uint32] `json:"Unknown3"`
    Unknown4 Value[uint32] `json:"Unknown4"`
    EntryIndex Value[uint32] `json:"EntryIndex"`
}

func (t *ToCFile) toJSON() (string, error) {
    b, err := json.Marshal(t)
    if err != nil {
        return "", err
    }
    return string(b), nil
}
