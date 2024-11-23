package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"
)

var TestingToCs = []string{
    "c024a2e8ae639757",
	"e75f556a740e00c9",
}

func ToCHeaderEqual(a *ToCHeader, b *ToCHeader, logger *slog.Logger) (bool, error) {
	if a.FileID != b.FileID {
		err_msg := fmt.Sprintf("a.FileID: %d != b.FileID: %d",
			a.FileID, b.FileID)
		return false, errors.New(err_msg)
	}

	if a.TypeID != b.TypeID {
		err_msg := fmt.Sprintf("a.TypeID: %d != b.TypeID: %d",
			a.TypeID, b.TypeID)
		return false, errors.New(err_msg)
	}

	if a.ToCDataOffset != b.ToCDataOffset {
		err_msg := fmt.Sprintf("a.ToCDataOffset: %d != b.ToCDataOffset: %d",
			a.ToCDataOffset, b.ToCDataOffset)
		return false, errors.New(err_msg)
	}

    if a.StreamFileOffset != b.StreamFileOffset {
		err_msg := fmt.Sprintf("a.StreamFileOffset %d ! = b.StreamFileOffset: %d",
			a.StreamFileOffset, b.StreamFileOffset)
		return false, errors.New(err_msg)
	}
    if a.GPUResourceOffset != b.GPUResourceOffset {
		err_msg := fmt.Sprintf("a.GPUResourceOffset %d ! = b.GPUResourceOffset: %d",
			a.GPUResourceOffset, b.GPUResourceOffset)
		return false, errors.New(err_msg)
	}
    if a.Unknown1 != b.Unknown1 {
		err_msg := fmt.Sprintf("a.Unknown1 %d ! = b.Unknown1: %d",
			a.Unknown1, b.Unknown1)
		return false, errors.New(err_msg)
	}
    if a.Unknown2 != b.Unknown2 {
		err_msg := fmt.Sprintf("a.Unknown2 %d ! = b.Unknown2: %d",
			a.Unknown2, b.Unknown2)
		return false, errors.New(err_msg)
	}
    if a.ToCDataSize != b.ToCDataSize {
		err_msg := fmt.Sprintf("a.ToCDataSize %d ! = b.ToCDataSize: %d",
			a.ToCDataSize, b.ToCDataSize)
		return false, errors.New(err_msg)
	}
    if a.StreamSize != b.StreamSize {
		err_msg := fmt.Sprintf("a.StreamSize %d ! = b.StreamSize: %d",
			a.StreamSize, b.StreamSize)
		return false, errors.New(err_msg)
	}
    if a.GPUResourceSize != b.GPUResourceSize {
		err_msg := fmt.Sprintf("a.GPUResourceSize %d ! = b.GPUResourceSize: %d",
			a.GPUResourceSize, b.GPUResourceSize)
		return false, errors.New(err_msg)
	}
    if a.Unknown3 != b.Unknown3 {
		err_msg := fmt.Sprintf("a.Unknown3 %d ! = b.Unknown3: %d",
			a.Unknown3, b.Unknown3)
		return false, errors.New(err_msg)
	}
    if a.Unknown4 != b.Unknown4 {
		err_msg := fmt.Sprintf("a.Unknown4 %d ! = b.Unknown4: %d",
			a.Unknown4, b.Unknown4)
		return false, errors.New(err_msg)
	}
    if a.EntryIndex != b.EntryIndex {
		err_msg := fmt.Sprintf("a.EntryIndex %d ! = b.EntryIndex: %d",
			a.EntryIndex, b.EntryIndex)
		return false, errors.New(err_msg)
	}

	logger.Info("ToCHeader equal pass", "a", a.FileID, "b", b.FileID)

	return true, nil
}

func ToCFileEqual(a *ToCFile, b *ToCFile, logger *slog.Logger) (bool, error) {
	if a == nil || b == nil {
		return false, errors.New("Nil reference")
	}

	if a.NumFiles != b.NumFiles {
		return false, errors.New("NumFiles is not equal")
	}
	
	if a.NumTypes != b.NumTypes {
		return false, errors.New("NumTypes is not equal")
	}

	if a.Unknown != b.Unknown {
		return false, errors.New("Unknown is not equal")
	}

	if len(a.Unk4Data) != len(b.Unk4Data) {
		return false, errors.New("Unk4Data length is not equal")
	}

	for i := range a.Unk4Data {
		if a.Unk4Data[i] != b.Unk4Data[i] {
			err_msg := fmt.Sprintf("Unk4Data is not equal @ position %d", i)
			return false, errors.New(err_msg)
		}
	}

	if len(a.ToCEntries) != len(b.ToCEntries) {
		return false, errors.New("ToCEntries length is not equal")
	}

	for i := range a.ToCEntries {
		_, err := ToCHeaderEqual(a.ToCEntries[i], b.ToCEntries[i], logger)
		if err != nil {
			err_msg := fmt.Sprintf("ToCEntry is different @ position %d", i)
			return false, errors.Join(errors.New(err_msg), err)
		}
	}

	logger.Info("ToCFile equal pass")

	return true, nil
}

func ToCFileBasicHeaderEqual(a *ToCFile, b *ToCFile, logger *slog.Logger) (bool, error) {
	if a == nil || b == nil {
		return false, errors.New("Nil reference")
	}

	if a.NumFiles != b.NumFiles {
		return false, errors.New("NumFiles is not equal")
	}
	
	if a.NumTypes != b.NumTypes {
		return false, errors.New("NumTypes is not equal")
	}

	if a.Unknown != b.Unknown {
		return false, errors.New("Unknown is not equal")
	}

	if len(a.Unk4Data) != len(b.Unk4Data) {
		return false, errors.New("Unk4Data length is not equal")
	}

	for i := range a.Unk4Data {
		if a.Unk4Data[i] != b.Unk4Data[i] {
			err_msg := fmt.Sprintf("Unk4Data is not equal @ position %d", i)
			return false, errors.New(err_msg)
		}
	}

	if len(a.ToCEntries) != len(b.ToCEntries) {
		return false, errors.New("ToCEntries length is not equal")
	}

	logger.Info("ToCFile basic header equal pass")
	
	return true, nil
}

func LoadExpectedToC() ([]ToCFile, error) { 
    ToCFiles := make([]ToCFile, len(TestingToCs))

    for i, t := range TestingToCs {
        b, err := os.ReadFile(t + "_ToC.json")
        if err != nil {
            return nil, err
        }
        err = json.Unmarshal(b, &ToCFiles[i])
        if err != nil {
            return nil, err
        }
    }

    return ToCFiles, nil
}

func TestParsingBasicHeader(t *testing.T) {
    ToCFiles, err := LoadExpectedToC()
    if err != nil {
        t.Fatal(err)
    }

	handler := slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelWarn,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{}
				}
				return a
			},
		},
	)

    logger := getLogger()(handler)
    for i, testingToC := range TestingToCs {
		expect := ToCFiles[i]
        t.Log("Parsing " + testingToC)
        f, err := os.Open(testingToC)
        if err != nil {
			t.Fatal(err) 
        }
		ToCFile, err := ParseToC(f, logger)
        if err != nil {
            t.Fatal(err)
        }
		if _, err = ToCFileBasicHeaderEqual(ToCFile, &expect, logger); err != nil {
			t.Fatal(err)
		}
    }
}

func TestParsingToCHeader(t *testing.T) {
    ToCFiles, err := LoadExpectedToC()
    if err != nil {
        t.Fatal(err)
    }

	handler := slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelWarn,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{}
				}
				return a
			},
		},
	)

    logger := getLogger()(handler)
    for i, testingToC := range TestingToCs {
		expect := ToCFiles[i]
        t.Log("Parsing " + testingToC)
        f, err := os.Open(testingToC)
        if err != nil {
			t.Fatal(err) 
        }
		ToCFile, err := ParseToC(f, logger)
        if err != nil {
            t.Fatal(err)
        }
		if _, err = ToCFileEqual(ToCFile, &expect, logger); err != nil {
			t.Fatal(err)
		}
    }
}
