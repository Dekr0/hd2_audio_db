package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
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

func extractWwiseSoundbank(tocFile *os.File) (
	map[uint64]*WwiseSoundbank, error) {
    reader := &StingRayAssetReader{ File: tocFile }

    var err error = nil

	magic, _,  err := reader.ReadUint32()
    if err != nil {
        return nil, err
    }

    if magic != MAGIC {
        return nil, errors.New("ToC file does not start with MAGIC number")
    }
	logger.Debug("Readed MAGIC", 
		"Magic", magic,
		"Head", reader.Head,
	)

	numTypes, _, err := reader.ReadUint32()
    if err != nil {
        return nil, err
    }
	logger.Debug("Readed NumTypes", 
		"NumTypes", numTypes,
		"Head", reader.Head,
	)

	numFiles, _, err := reader.ReadUint32()
    if err != nil {
        return nil, err
    }
	logger.Debug("Readed NumFiles", 
		"NumTypes", numFiles,
		"Head", reader.Head,
	)

	reader.RelativeSeek(int64(60 + 32 * numTypes))

    ToCStart := reader.Head

	logger.Debug("ToCStart Checkpoint",
		"Head", ToCStart,
	)

	banks := make(map[uint64]*WwiseSoundbank)
	bankDeps := make(map[uint64]*WwiseSoundbankDep)

    for i := 0; i < int(numFiles); i++ {
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
				Id: header.FileID.Value,
				RawData: data,
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
	 for id, bank := range banks {
		 bankDep, in := bankDeps[id]
		 if !in {
			 logger.Warn("A Wwise Soundbank without a Wwise Soundbank " +
			 "dependencies", "ToCFileID", id)
			 continue
		 }
		 bank.PathName = bankDep.PathName
	 }

    return nil, nil
}


func ParseToC(tocFile *os.File) (*ToCFile, error) {
    ToC := ToCFile{}

    reader := &StingRayAssetReader{ File: tocFile }

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

func genWwiserXML(gameArchiveIDsArg *string) {
	gameArchiveIDs := strings.Split(*gameArchiveIDsArg, ",")

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

		banks, err := extractWwiseSoundbank(tocFile)
		if err != nil {
			logger.Error("Failed to parse ToC File", "error", err)
		}
		
		for id, bank := range banks {
			if err = bank.genWwiserXML(); err != nil {
				logger.Error("Failed to generate Wwiser XML", "TocFileID", id,
				"error", err)
			}
		}
	}
}
