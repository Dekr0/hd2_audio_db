package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func ParseToCHeader(r *StingRayAssetReader) (*ToCHeader, error) {
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

func ParseToCBasic(reader *StingRayAssetReader) (*ToCFile, error) {
    ToC := ToCFile{}

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

	return &ToC, nil
}

func ExtractWwiseSoundbankMediaHeader(data []byte) []*MediaHeader {
	for i := 0; i < len(data) / 12; i++ {
		
	}

	return nil
}

func ExtractWwiseStream(tocFile *os.File) (map[uint64]struct{}, error) {
	reader := &StingRayAssetReader{ File: tocFile }

	ToC, err := ParseToCBasic(reader)
	if err != nil {
		return nil, errors.Join(
			errors.New("Failed to parse ToC File's basic information"), err)
	}

	reader.RelativeSeek(int64(60 + 32 * ToC.NumTypes.Value))

    ToCStart := reader.Head

	logger.Debug("ToCStart Checkpoint",
		"Head", ToCStart,
	)

	streams := make(map[uint64]struct{})

	for i := 0; i < int(ToC.NumFiles.Value); i++ {
        if err = reader.AbsoluteSeek(ToCStart + int64(i) * 80); err != nil {
            return nil, err
        }

		ToCPrev := reader.Head
		header, err := ParseToCHeader(reader)
		if err != nil {
			return nil, errors.Join(
				errors.New("Failed to parse ToC Header"), err)
		}

		logger.Debug("Head change",
			"Before", ToCPrev,
			"After", reader.Head,
			"Diff", reader.Head - ToCPrev,
		)
        logger.Info("Parsed ToC Header File ID", "FileID", header.FileID)

		if header.TypeID.Value == TYPE_WWISE_STREAM {
			if _, in := streams[header.FileID.Value]; in {
				logger.Warn("A Wwise stream with duplicated ToC file ID",
					"ToCFileID", header.FileID.Value,
				)
			}
			streams[header.FileID.Value] = struct{}{}
		}
	}

	return streams, nil
}

func ExtractWwiseSoundbank(tocFile *os.File, rawData bool) (
	map[uint64]*WwiseSoundbank, error) {
    reader := &StingRayAssetReader{ File: tocFile }

	ToC, err := ParseToCBasic(reader)
	if err != nil {
		return nil, errors.Join(
			errors.New("Failed to parse ToC File's basic information"), err)
	}

	reader.RelativeSeek(int64(32 * ToC.NumTypes.Value))

    ToCStart := reader.Head

	logger.Debug("ToCStart Checkpoint",
		"Head", ToCStart,
	)

	banks := make(map[uint64]*WwiseSoundbank)
	bankDeps := make(map[uint64]*WwiseSoundbankDep)

    for i := 0; i < int(ToC.NumFiles.Value); i++ {
        if err = reader.AbsoluteSeek(ToCStart + int64(i) * 80); err != nil {
            return nil, err
        }

		ToCPrev := reader.Head
        header, err := ParseToCHeader(reader)
        if err != nil {
            return nil, err
        }

		logger.Debug("Head change",
			"Before", ToCPrev,
			"After", reader.Head,
			"Diff", reader.Head - ToCPrev,
		)
        logger.Debug("Parsed ToC Header File ID", 
			"FileID", header.FileID.Value,
			"TypeID", header.TypeID.Value)

		if header.TypeID.Value == TYPE_WWISE_BANK {
			if _, in := banks[header.FileID.Value]; in {
				errMsg := fmt.Sprintf("Duplicate Wwise Soundbank file ID: %d", 
				header.FileID.Value)
				return nil, errors.New(errMsg)
			}

			err = reader.AbsoluteSeek(int64(header.ToCDataOffset.Value + 16))
			if err != nil {
				return nil, err
			}

			data := make([]byte, header.ToCDataSize.Value - 16)
			_, err := reader.Read(data)
			if err != nil {
				return nil, err
			}

			banks[header.FileID.Value] = &WwiseSoundbank{
				ToCDataSize: header.ToCDataSize.Value - 16,
				ToCFileId: header.FileID.Value,
				ToCDataOffset: header.ToCDataOffset.Value,
			}
			if rawData {
				banks[header.FileID.Value].RawData = data
			}
		} else if header.TypeID.Value == TYPE_WWISE_DEP {
			if _, in := bankDeps[header.FileID.Value]; in {
				errMsg := fmt.Sprintf(
					"Duplicated Wwise Soundbank dependencies file ID: %d", 
					header.FileID.Value)
				return nil, errors.New(errMsg)
			}

			err = reader.AbsoluteSeek(int64(header.ToCDataOffset.Value + 4))
			if err != nil {
				return nil, err
			}

			size, _, err := reader.ReadUint32()
			if err != nil {
				return nil, err
			}

			buf := make([]byte, size)
			_, err = reader.Read(buf)
			if err != nil {
				return nil, err
			}

			bankDeps[header.FileID.Value] = &WwiseSoundbankDep{
				header.FileID.Value,
				strings.Replace(
					strings.Replace(string(buf), "\u0000", "", -1),
					"/", "_", -1),
			}
		}
    }

	/** 
	 * In case of which Wwise Soundbank and Wwise Soundbank Dependencies are 
	 * put in out of order
	 * */
	 logger.Info("", "banks", len(banks), "deps", len(bankDeps))
	 for id, bank := range banks {
		 bankDep, in := bankDeps[id]
		 if !in {
			 logger.Warn("A Wwise Soundbank without a Wwise Soundbank " +
			 "dependencies", "ToCFileID", id)
			 continue
		 }
		 bank.PathName = bankDep.PathName
	 }

    return banks, nil
}

func ParseToC(tocFile *os.File) (*ToCFile, error) {
    reader := &StingRayAssetReader{ File: tocFile }

	ToC, err := ParseToCBasic(reader)
	if err != nil {
		return nil, errors.Join(
			errors.New("Failed to parse basic information of ToCFile"),
			err)
	}

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
        header, err := ParseToCHeader(reader)
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
    return ToC, err 
}

func ExtractBank(extractBankFlag *string) {
	gameArchiveIDs := strings.Split(*extractBankFlag, ",")

	for _, gameArchiveID := range gameArchiveIDs {
		_, err := os.Stat(gameArchiveID); 
		if err != nil {
			if os.IsNotExist(err) {
				logger.Error("Game archive ID " + gameArchiveID + " does not exist") 
				continue
			} else {
				logger.Error("OS error", "error", err)
				continue
			}
		}

		tocFile, err := os.Open(gameArchiveID)
		if err != nil {
			logger.Error("File open error", "error", err)
			continue
		}
		defer tocFile.Close()

		banks, err := ExtractWwiseSoundbank(tocFile, true)
		if err != nil {
			logger.Error("Failed to parse ToC File", "error", err)
		}
		
		for id, bank := range banks {
			if err = bank.exportToCHeader(); err != nil {
				logger.Error("Failed to export ToC header", "ToCFileID", id, 
				"error", err)
			}
			if err = bank.exportWwiserXML(); err != nil {
				logger.Error("Failed to generate Wwiser XML", "TocFileID", id,
				"error", err)
			}
		}
	}
}
