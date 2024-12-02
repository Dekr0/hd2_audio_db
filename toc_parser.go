package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

// Main entry point for extracting Wwise Soundbank binary content from a given
// list of game archives. The binary content will be written in disk in the form
// of a file.
// It will generate a json file that contain the bytes offset of a given Wwise
// Soundbank in a game archive. This is useful when using hex editor.
// It will generate a XML file using Wwiser. This XML file contain semi-readable
// information a given Wwise Soundbank
//
// [parameter]
// extractBankFlag - a pointer to string that contain a list of game archive IDs
// separated by ","
func extractBank(extractBankFlag string) {
	gameArchiveIDs := strings.Split(extractBankFlag, ",")

	for _, gameArchiveID := range gameArchiveIDs {
        tocFile, err := openToCFile(gameArchiveID)
        if err != nil {
            DefaultLogger.Error("Failed to open ToC File", "error", err)
            continue
        }
        defer tocFile.Close()

		banks, err := extractWwiseSoundbanks(*tocFile, true)
		if err != nil {
			DefaultLogger.Error("Failed to parse ToC File", "error", err)
            continue
		}
		
		for id, bank := range banks {
			if err = bank.exportToCHeader(); err != nil {
				DefaultLogger.Error("Failed to export ToC header", "ToCFileID", id, 
				"error", err)
			}
			if err = bank.exportWwiserXML(false); err != nil {
				DefaultLogger.Error("Failed to generate Wwiser XML", "ToCFileID", id,
				"error", err)
			}
		}
	}
}

// Under construction
func extractWwiseSoundbankMediaHeader(data []byte) []*MediaHeader {
	for i := 0; i < len(data) / 12; i++ {
		
	}

	return nil
}

// Under construction
func extractWwiseStream(tocFile os.File) (map[uint64]Empty, error) {
	reader := &StingRayAssetReader{ file: tocFile }

	ToC, err := parseToCBasic(reader)
	if err != nil {
		return nil, errors.Join(
			errors.New("Failed to parse ToC File's basic information"), err)
	}

	reader.relativeSeek(int64(60 + 32 * ToC.NumTypes.Value))

    ToCStart := reader.head

	DefaultLogger.Debug("ToCStart Checkpoint",
		"Head", ToCStart,
	)

	streams := make(map[uint64]Empty)

	for i := 0; i < int(ToC.NumFiles.Value); i++ {
        if err = reader.absoluteSeek(ToCStart + int64(i) * 80); err != nil {
            return nil, err
        }

		ToCPrev := reader.head
		header, err := parseToCHeader(reader)
		if err != nil {
			return nil, errors.Join(
				errors.New("Failed to parse ToC Header"), err)
		}

		DefaultLogger.Debug("Head change",
			"Before", ToCPrev,
			"After", reader.head,
			"Diff", reader.head - ToCPrev,
		)
        DefaultLogger.Info("Parsed ToC Header File ID", "fileID", header.FileID)

		if header.TypeID.Value == TYPE_WWISE_STREAM {
			if _, in := streams[header.FileID.Value]; in {
				DefaultLogger.Warn("A Wwise stream with duplicated ToC file ID",
					"ToCFileID", header.FileID.Value,
				)
			}
			streams[header.FileID.Value] = Empty{}
		}
	}

	return streams, nil
}

// Extract all Wwise Soundbank from a given ToC file.
//
// [parameter]
// tocFile - trivial
// rawData - This specifies whether include the all binary content of a given 
// Wwise Soundbank. This is useful for exporting the soundbank into a local 
// file
//
// [return]
// map[uint64]*ToCWwiseSoundbank - All Wwise Soundbanks in a given game archive.
// Nil when an error occur.
// error - trivial
func extractWwiseSoundbanks(tocFile os.File, rawData bool) (
	map[uint64]*ToCWwiseSoundbank, error) {
    reader := &StingRayAssetReader{ file: tocFile }

	ToC, err := parseToCBasic(reader)
	if err != nil {
		return nil, errors.Join(
			errors.New("Failed to parse ToC File's basic information"), err)
	}

	if err = reader.relativeSeek(int64(32 * ToC.NumTypes.Value)); err != nil {
        return nil, err
    }

    ToCStart := reader.head

	DefaultLogger.Debug("ToC starting checkpoint",
		"Head", ToCStart,
	)

	banks := make(map[uint64]*ToCWwiseSoundbank)
	bankDeps := make(map[uint64]*WwiseSoundbankDep)

    for i := 0; i < int(ToC.NumFiles.Value); i++ {
        if err = reader.absoluteSeek(ToCStart + int64(i) * 80); err != nil {
            return nil, err
        }

		ToCPrev := reader.head
        header, err := parseToCHeader(reader)
        if err != nil {
            return nil, err
        }

		DefaultLogger.Debug("Head change",
			"Before", ToCPrev,
			"After", reader.head,
			"Diff", reader.head - ToCPrev,
		)
        DefaultLogger.Debug("Parsed ToC Header File ID", 
			"FileID", header.FileID.Value,
			"TypeID", header.TypeID.Value)

		if header.TypeID.Value == TYPE_WWISE_BANK {
			if _, in := banks[header.FileID.Value]; in {
				errMsg := fmt.Sprintf("Duplicate Wwise Soundbank file ID: %d", 
				header.FileID.Value)
				return nil, errors.New(errMsg)
			}

			err = reader.absoluteSeek(int64(header.ToCDataOffset.Value + 16))
			if err != nil {
				return nil, err
			}

            dataSize := header.ToCDataSize.Value - 16
			data := make([]byte, dataSize, dataSize)
			_, err := reader.read(data)
			if err != nil {
				return nil, err
			}

			banks[header.FileID.Value] = &ToCWwiseSoundbank{
				ToCDataSize: dataSize,
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

			err = reader.absoluteSeek(int64(header.ToCDataOffset.Value + 4))
			if err != nil {
				return nil, err
			}

			size, _, err := reader.readUint32()
			if err != nil {
				return nil, err
			}

			buf := make([]byte, size, size)
			_, err = reader.read(buf)
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
	 DefaultLogger.Debug("Wwise Soundbank extraction result", 
     "NumBanks", len(banks),
     "NumDeps", len(bankDeps))
	 for id, bank := range banks {
         if bank == nil {
             panic("Assertion failure. A nil Wwise Soundbank when organizing " +
             "Wwise dependencies")
         }

		 bankDep, in := bankDeps[id]
		 if !in {
			 DefaultLogger.Warn("A Wwise Soundbank without a Wwise Soundbank " +
			 "dependencies", "ToCFileID", id)
			 continue
		 }
		 bank.PathName = bankDep.PathName
	 }

    return banks, nil
}

// A helper function for opening up a game archive ToC file based on a given 
// game data directory.
// 
// [return]
// *os.File - nil when an error occur
// error - trivial
func openToCFile(gameArchiveID string) (*os.File, error) {
	tocFilePath := path.Join(DATA, gameArchiveID)

	_, err := os.Stat(tocFilePath); 
	if err != nil {
        return nil, err
	}

	tocFile, err := os.Open(tocFilePath)
	if err != nil {
        return nil, err
	}

    return tocFile, nil
}

// [return]
// *ToCFile - trivial. Nil when an error occur
// error - trivial
func parseToC(tocFile os.File) (*ToCFile, error) {
    reader := &StingRayAssetReader{ file: tocFile }

	ToC, err := parseToCBasic(reader)
	if err != nil {
		return nil, errors.Join(
			errors.New("Failed to parse basic information of ToCFile"),
			err)
	}

    err = reader.relativeSeek(int64(32 * ToC.NumTypes.Value))
    ToCStart := reader.head

	DefaultLogger.Debug("ToCStart Checkpoint",
		"Head", ToCStart,
	)
    ToC.ToCEntries = make([]*ToCHeader, ToC.NumFiles.Value)

    for i := 0; i < int(ToC.NumFiles.Value); i++ {
        if err = reader.absoluteSeek(ToCStart + int64(i) * 80); err != nil {
            return nil, err
        }
		ToCPrev := reader.head
        header, err := parseToCHeader(reader)
        if err != nil {
            return nil, err
        }
		DefaultLogger.Debug("Head change",
			"Before", ToCPrev,
			"After", reader.head,
			"Diff", reader.head - ToCPrev,
		)
        DefaultLogger.Info("Parsed ToC Header File ID", "fileID", header.FileID)

        ToC.ToCEntries[i] = header
    }
    return ToC, err 
}

// Parsing logic for parsing the initial basic information in a game archive. 
// This information typicall sit right before all ToC headers and ToC data.
// All basic information will write into the given pointer of a struct that 
// encasuplate the basic structure of a game archive ToC file.
//
// [return]
// *ToCFile - a pointer of struct that encapsulate all information about the ToC 
// file. Nil when an error occur
func parseToCBasic(reader *StingRayAssetReader) (*ToCFile, error) {
    ToC := ToCFile{}

    var err error = nil

	ToC.Magic.Value, ToC.Magic.Address, err = reader.readUint32()
    if err != nil {
        return nil, err
    }

    if ToC.Magic.Value != MAGIC {
        return nil, errors.New("ToC file does not start with MAGIC number")
    }
	DefaultLogger.Debug("Readed MAGIC", 
		"Magic", ToC.Magic.Value,
		"Head", reader.head,
	)

	ToC.NumTypes.Value, ToC.NumTypes.Address, err = reader.readUint32()
    if err != nil {
        return nil, err
    }
	DefaultLogger.Debug("Readed NumTypes", 
		"NumTypes", ToC.NumTypes.Value,
		"Head", reader.head,
	)

	ToC.NumFiles.Value, ToC.NumFiles.Address, err = reader.readUint32()
    if err != nil {
        return nil, err
    }
	DefaultLogger.Debug("Readed NumFiles", 
		"NumTypes", ToC.NumFiles,
		"Head", reader.head,
	)

	ToC.Unknown.Value, ToC.Unknown.Address, err = reader.readUint32()
    if err != nil {
        return nil, err
    }
	DefaultLogger.Debug("Readed Unknown",
		"Unknown", ToC.Unknown,
		"Head", reader.head,
	)

	unk4Data := make([]byte, 56)
	head, err := reader.read(unk4Data)
    if err != nil {
        return nil, err
    }
	for i, d := range unk4Data {
		ToC.Unk4Data[i].Address = int64(i) + head
		ToC.Unk4Data[i].Value = d
	}

	DefaultLogger.Debug("Readed Unk4Data", 
		"Unk4Data", ToC.Unk4Data,
		"Head", reader.head,
	)

	DefaultLogger.Debug("Basic Header Checkpoint",
		"Head", reader.head,
	)

	return &ToC, nil
}

// Consume bytes from the file, and translate those bytes into a struct that 
// encapsulate information of ToC header
//
// [return]
// *ToCHeader - trivial. Nil when an error occurs
// error - trivial
func parseToCHeader(r *StingRayAssetReader) (*ToCHeader, error) {
    var err error

    header := &ToCHeader{}

	header.FileID.Value, header.FileID.Address, err = r.readUint64()
    if err != nil {
        return nil, err
    }

	header.TypeID.Value, header.TypeID.Address, err = r.readUint64()
    if err != nil {
        return nil, err
    }

	header.ToCDataOffset.Value, header.ToCDataOffset.Address, err = r.readUint64()
    if err != nil {
        return nil, err
    }

	header.StreamFileOffset.Value, header.StreamFileOffset.Address, err = r.readUint64()
    if err != nil {
        return nil, err
    }

	header.GPUResourceOffset.Value, header.GPUResourceOffset.Address, err = r.readUint64()
    if err != nil {
        return nil, err
    }

	header.Unknown1.Value, header.Unknown1.Address, err = r.readUint64()
    if err != nil {
        return nil, err
    }

	header.Unknown2.Value, header.Unknown2.Address, err = r.readUint64()
    if err != nil {
        return nil, err
    }

	header.ToCDataSize.Value, header.ToCDataSize.Address, err = r.readUint32()
    if err != nil {
        return nil, err
    }

	header.StreamSize.Value, header.StreamSize.Address, err = r.readUint32()
    if err != nil {
        return nil, err
    }

	header.GPUResourceSize.Value, header.GPUResourceSize.Address, err = r.readUint32()
    if err != nil {
        return nil, err
    }

	header.Unknown3.Value, header.Unknown3.Address, err = r.readUint32()
    if err != nil {
        return nil, err
    }

	header.Unknown4.Value, header.Unknown4.Address, err = r.readUint32()
    if err != nil {
        return nil, err
    }

	header.EntryIndex.Value, header.EntryIndex.Address, err = r.readUint32()
    if err != nil {
        return nil, err
    }

    return header, nil
}
