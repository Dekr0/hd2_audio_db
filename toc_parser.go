package main

import (
	"errors"
	"log/slog"
	"os"
)

func parseToCHeader(r *StingRayAssetReader) (*ToCHeader, error) {
    var err error

    header := &ToCHeader{}

	header.FileID.Value, header.FileID.Address, err = r.ReadUint64()
    if err != nil {
        return nil, err
    }

	header.TypeID.Value, header.TypeID.Address, err = r.ReadUint64()
    if err != nil {
        return nil, err
    }

	header.ToCDataOffset.Value, header.ToCDataOffset.Address, err = r.ReadUint64()
    if err != nil {
        return nil, err
    }

	header.StreamFileOffset.Value, header.StreamFileOffset.Address, err = r.ReadUint64()
    if err != nil {
        return nil, err
    }

	header.GPUResourceOffset.Value, header.GPUResourceOffset.Address, err = r.ReadUint64()
    if err != nil {
        return nil, err
    }

	header.Unknown1.Value, header.Unknown1.Address, err = r.ReadUint64()
    if err != nil {
        return nil, err
    }

	header.Unknown2.Value, header.Unknown2.Address, err = r.ReadUint64()
    if err != nil {
        return nil, err
    }

	header.ToCDataSize.Value, header.ToCDataSize.Address, err = r.ReadUint32()
    if err != nil {
        return nil, err
    }

	header.StreamSize.Value, header.StreamSize.Address, err = r.ReadUint32()
    if err != nil {
        return nil, err
    }

	header.GPUResourceSize.Value, header.GPUResourceSize.Address, err = r.ReadUint32()
    if err != nil {
        return nil, err
    }

	header.Unknown3.Value, header.Unknown3.Address, err = r.ReadUint32()
    if err != nil {
        return nil, err
    }

	header.Unknown4.Value, header.Unknown4.Address, err = r.ReadUint32()
    if err != nil {
        return nil, err
    }

	header.EntryIndex.Value, header.EntryIndex.Address, err = r.ReadUint32()
    if err != nil {
        return nil, err
    }

    return header, nil
}

func ParseToC(f *os.File, logger *slog.Logger) (*ToCFile, error) {
    ToC := ToCFile{}

    reader := &StingRayAssetReader{ File: f }

    var err error = nil

	ToC.Magic.Value, ToC.Magic.Address, err = reader.ReadUint32()
    if err != nil {
        return nil, err
    }

    if ToC.Magic.Value != MAGIC {
        return nil, errors.New("ToC file does not start with MAGIC number")
    }
	logger.Debug("Readed MAGIC", 
		"Magic", ToC.Magic.Value,
		"Head", reader.Head,
	)

	ToC.NumTypes.Value, ToC.NumTypes.Address, err = reader.ReadUint32()
    if err != nil {
        return nil, err
    }
	logger.Debug("Readed NumTypes", 
		"NumTypes", ToC.NumTypes.Value,
		"Head", reader.Head,
	)

	ToC.NumFiles.Value, ToC.NumFiles.Address, err = reader.ReadUint32()
    if err != nil {
        return nil, err
    }
	logger.Debug("Readed NumFiles", 
		"NumTypes", ToC.NumFiles,
		"Head", reader.Head,
	)

	ToC.Unknown.Value, ToC.Unknown.Address, err = reader.ReadUint32()
    if err != nil {
        return nil, err
    }
	logger.Debug("Readed Unknown",
		"Unknown", ToC.Unknown,
		"Head", reader.Head,
	)

	unk4Data := make([]byte, 56)
	head, err := reader.Read(unk4Data)
    if err != nil {
        return nil, err
    }
	for i, d := range unk4Data {
		ToC.Unk4Data[i].Address = int64(i) + head
		ToC.Unk4Data[i].Value = d
	}

	logger.Debug("Readed Unk4Data", 
		"Unk4Data", ToC.Unk4Data,
		"Head", reader.Head,
	)

	logger.Debug("Basic Header Checkpoint",
		"Head", reader.Head,
	)

    err = reader.RelativeSeek(int64(32 * ToC.NumTypes.Value))
    ToCStart := reader.Head

	logger.Debug("ToCStart Checkpoint",
		"Head", ToCStart,
	)
    ToC.ToCEntries = make([]*ToCHeader, ToC.NumFiles.Value)

    for i := 0; i < int(ToC.NumFiles.Value); i++ {
        if err = reader.AbsoluteSeek(ToCStart + int64(i) * 80); err != nil {
            return nil, err
        }
		ToCPrev := reader.Head
        header, err := parseToCHeader(reader)
        if err != nil {
            return nil, err
        }
		logger.Debug("Head change",
			"Before", ToCPrev,
			"After", reader.Head,
			"Diff", reader.Head - ToCPrev,
		)
        logger.Info("Parsed ToC Header File ID", "FileID", header.FileID)

        ToC.ToCEntries[i] = header
    }
    return &ToC, err 
}
