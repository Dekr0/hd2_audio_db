package parser

import (
	"dekr0/hd2_audio_db/io"
	"errors"
	"sync"
)

var MaxParser uint32 = 4

const defaultBnkPerArchive = 2

var NotHelldiversGameArchive error = errors.New(
	"Not a game archive used by Helldivers 2",
)

type AssetType uint64

const (
	AssetTypeSoundBank       AssetType = 6006249203084351385
	AssetTypeWwiseDependency           = 12624162998411505776
	AssetTypeWwiseStream               = 5785811756662211598
)

const MagicValue uint32 = 0xF0000011

func ParseArchiveHeader(a *Archive, r *io.Reader) {
	if r.U32Unsafe() != MagicValue {
		panic(NotHelldiversGameArchive)
	}
	a.NumTypes = r.U32Unsafe()
	a.NumFiles = r.U32Unsafe()
	a.Unknown = r.U32Unsafe()
	r.ReadUnsafe(a.Unk4Data[:])
	a.AssetTypeCnts = make([]AssetTypeCnt, a.NumTypes, a.NumTypes)
	for i := range a.AssetTypeCnts {
		r.U64Unsafe()
		a.AssetTypeCnts[i].Type = r.U64Unsafe()
		a.AssetTypeCnts[i].Num = r.U64Unsafe()
		if a.AssetTypeCnts[i].Type == uint64(AssetTypeSoundBank) {
			a.SoundBnks = make([]uint32, 0, a.AssetTypeCnts[i].Num)
		} else if a.AssetTypeCnts[i].Type == uint64(AssetTypeWwiseDependency) {
			a.Deps = make([]uint32, 0, a.AssetTypeCnts[i].Num)
		}
		r.U32Unsafe()
		r.U32Unsafe()
	}
	a.Headers = make([]AssetHeader, a.NumFiles, a.NumFiles)
}

func ParseArchiveHeaderCache(a *Archive, r *io.Reader) {
	data := make([]byte, 72, 72)
	r.ReadUnsafe(data)

	ir := io.NewInPlaceReader(data, io.ByteOrder)
	if ir.U32Unsafe() != MagicValue {
		panic(NotHelldiversGameArchive)
	}
	a.NumTypes = ir.U32Unsafe()
	a.NumFiles = ir.U32Unsafe()
	a.Unknown = ir.U32Unsafe()
	a.Unk4Data = [56]byte(ir.ReadNoCopyUnsafe(56))
	a.AssetTypeCnts = make([]AssetTypeCnt, a.NumTypes, a.NumTypes)

	size := a.NumTypes * (sizeOfAssetCnt + 16)
	data = make([]byte, size, size)
	r.ReadUnsafe(data)

	ir = io.NewInPlaceReader(data, io.ByteOrder)
	for i := range a.AssetTypeCnts {
		ir.RelSeekUnsafe(8)
		a.AssetTypeCnts[i].Type = ir.U64Unsafe()
		a.AssetTypeCnts[i].Num = ir.U64Unsafe()
		if a.AssetTypeCnts[i].Type == uint64(AssetTypeSoundBank) {
			a.SoundBnks = make([]uint32, 0, a.AssetTypeCnts[i].Num)
		} else if a.AssetTypeCnts[i].Type == uint64(AssetTypeWwiseDependency) {
			a.Deps = make([]uint32, 0, a.AssetTypeCnts[i].Num)
		}
		ir.RelSeekUnsafe(8)
	}
	a.Headers = make([]AssetHeader, a.NumFiles, a.NumFiles)
}

func ParseAssetHeadersCachePerThread(a *Archive, r *io.Reader) {
	if MaxParser > 1 && MaxParser < a.NumFiles {
		var w sync.WaitGroup
		base := a.NumFiles / MaxParser
		prev := uint32(0)
		rest := a.NumFiles % MaxParser
		for range MaxParser {
			head := prev
			tail := head + base
			if rest > 0 {
				tail += 1
				rest -= 1
			}
			prev = tail
			lower := head * sizeOfAssetHeader
			upper := tail * sizeOfAssetHeader
			data := make([]byte, upper-lower, upper-lower)
			r.ReadUnsafe(data)
			r := io.NewInPlaceReader(data, io.ByteOrder)
			w.Add(1)
			go parseAssetHeader(&w, r, head, tail, a)
		}
		w.Wait()
	} else {
		data := make([]byte, a.NumFiles*sizeOfAssetHeader, a.NumFiles*sizeOfAssetHeader)
		r.ReadUnsafe(data)
		r := io.NewInPlaceReader(data, io.ByteOrder)
		for i := range a.NumFiles {
			a.Headers[i].FileID = r.U64Unsafe()
			a.Headers[i].TypeID = r.U64Unsafe()
			a.Headers[i].DataOffset = r.U64Unsafe()
			a.Headers[i].StreamOffset = r.U64Unsafe()
			a.Headers[i].GPURsrcOffset = r.U64Unsafe()
			a.Headers[i].UnknownU64A = r.U64Unsafe()
			a.Headers[i].UnknownU64B = r.U64Unsafe()
			a.Headers[i].DataSize = r.U32Unsafe()
			a.Headers[i].StreamSize = r.U32Unsafe()
			a.Headers[i].GPURsrcSize = r.U32Unsafe()
			a.Headers[i].UnknownU32A = r.U32Unsafe()
			a.Headers[i].UnknownU32B = r.U32Unsafe()
			a.Headers[i].Idx = r.U32Unsafe()
			if a.Headers[i].TypeID == uint64(AssetTypeSoundBank) {
				a.SoundBnks = append(a.SoundBnks, i)
			} else if a.Headers[i].TypeID == uint64(AssetTypeWwiseDependency) {
				a.Deps = append(a.Deps, i)
			}
		}
	}
}

func ParseAssetHeaders(a *Archive, r *io.Reader) {
	if MaxParser > 1 && MaxParser < a.NumFiles {
		data := make([]byte, a.NumFiles*80, a.NumFiles*80)
		r.ReadUnsafe(data)
		var w sync.WaitGroup
		base := a.NumFiles / MaxParser
		prev := uint32(0)
		rest := a.NumFiles % MaxParser
		for range MaxParser {
			head := prev
			tail := head + base
			if rest > 0 {
				tail += 1
				rest -= 1
			}
			prev = tail
			lower := head * sizeOfAssetHeader
			upper := tail * sizeOfAssetHeader
			r := io.NewInPlaceReader(data[lower:upper], io.ByteOrder)
			w.Add(1)
			go parseAssetHeader(&w, r, head, tail, a)
		}
		w.Wait()
	} else {
		data := make([]byte, a.NumFiles * sizeOfAssetHeader, a.NumFiles * sizeOfAssetHeader)
		r.ReadUnsafe(data)
		r := io.NewInPlaceReader(data, io.ByteOrder)
		for i := range a.NumFiles {
			a.Headers[i].FileID = r.U64Unsafe()
			a.Headers[i].TypeID = r.U64Unsafe()
			a.Headers[i].DataOffset = r.U64Unsafe()
			a.Headers[i].StreamOffset = r.U64Unsafe()
			a.Headers[i].GPURsrcOffset = r.U64Unsafe()
			a.Headers[i].UnknownU64A = r.U64Unsafe()
			a.Headers[i].UnknownU64B = r.U64Unsafe()
			a.Headers[i].DataSize = r.U32Unsafe()
			a.Headers[i].StreamSize = r.U32Unsafe()
			a.Headers[i].GPURsrcSize = r.U32Unsafe()
			a.Headers[i].UnknownU32A = r.U32Unsafe()
			a.Headers[i].UnknownU32B = r.U32Unsafe()
			a.Headers[i].Idx = r.U32Unsafe()
			if a.Headers[i].TypeID == uint64(AssetTypeSoundBank) {
				a.SoundBnks = append(a.SoundBnks, i)
			} else if a.Headers[i].TypeID == uint64(AssetTypeWwiseDependency) {
				a.Deps = append(a.Deps, i)
			}
		}
	}
}

func parseAssetHeader(
	w *sync.WaitGroup,
	r *io.InPlaceReader,
	i uint32,
	j uint32,
	a *Archive,
) {
	soundBnks := []uint32{}
	deps := []uint32{}
	for ; i < j; i++ {
		a.Headers[i].FileID = r.U64Unsafe()
		a.Headers[i].TypeID = r.U64Unsafe()
		a.Headers[i].DataOffset = r.U64Unsafe()
		a.Headers[i].StreamOffset = r.U64Unsafe()
		a.Headers[i].GPURsrcOffset = r.U64Unsafe()
		a.Headers[i].UnknownU64A = r.U64Unsafe()
		a.Headers[i].UnknownU64B = r.U64Unsafe()
		a.Headers[i].DataSize = r.U32Unsafe()
		a.Headers[i].StreamSize = r.U32Unsafe()
		a.Headers[i].GPURsrcSize = r.U32Unsafe()
		a.Headers[i].UnknownU32A = r.U32Unsafe()
		a.Headers[i].UnknownU32B = r.U32Unsafe()
		a.Headers[i].Idx = r.U32Unsafe()
		if a.Headers[i].TypeID == uint64(AssetTypeSoundBank) {
			soundBnks = append(soundBnks, i)
		} else if a.Headers[i].TypeID == uint64(AssetTypeWwiseDependency) {
			deps = append(deps, i)
		}
	}
	a.soundBnkMu.Lock()
	a.SoundBnks = append(a.SoundBnks, soundBnks...)
	a.soundBnkMu.Unlock()
	a.depMu.Lock()
	a.Deps = append(a.Deps, deps...)
	a.depMu.Unlock()
	w.Done()
}
