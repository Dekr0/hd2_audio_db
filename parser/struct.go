package parser

import "sync"

type Archive struct {
	NumTypes      uint32         `json:"NumTypes"`
	NumFiles      uint32         `json:"NumFiles"`
	Unknown       uint32         `json:"Unknown"`
	Unk4Data      [56]uint8      `json:"Unk4Data"`
	AssetTypeCnts []AssetTypeCnt `json:"AssetTypeCnts,omitEmpty"`
	Headers       []AssetHeader  `json:"Headers,omitEmpty"`

	SoundBnks     []uint32
	soundBnkMu    sync.Mutex
	Deps          []uint32
	depMu         sync.Mutex
}

const sizeOfAssetCnt = 16
type AssetTypeCnt struct {
	Type uint64 `json:"Type"`
	Num  uint64 `json:"Num"`
}

const sizeOfAssetHeader = 80

type AssetHeader struct {
	FileID        uint64 `json:"fileID"`
	TypeID        uint64 `json:"typeID"`
	DataOffset    uint64 `json:"dataOffset"`
	StreamOffset  uint64 `json:"streamOffset"`
	GPURsrcOffset uint64 `json:"GPURsrcOffset"`
	UnknownU64A   uint64 `json:"unknownU64A"`
	UnknownU64B   uint64 `json:"unknownU64B"`
	DataSize      uint32 `json:"dataSize"`
	StreamSize    uint32 `json:"streamSize"`
	GPURsrcSize   uint32 `json:"gPURsrcSize"`
	UnknownU32A   uint32 `json:"unknownU32A"`
	UnknownU32B   uint32 `json:"unknownU32B"`
	Idx           uint32 `json:"idx"`

	UUID          string
}

type HIRC struct {
	Header    uint32
	Hierarchy []Hierarchy
	Sound     []Sound
}

type Hierarchy struct {
	Type   HircType
	ID     uint32
	Parent uint32
}

type Sound struct {
	Idx               uint32
	PluginID          uint32
	StreamType        uint8
	SourceID          uint32
	InMemoryMediaSize uint32
	SourceBits        uint8
	PluginParamSize   uint32
}
