package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

const MAGIC = 4026531857

var DATA string

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

type MediaHeader struct {
	ObjectId uint64
	ShortId uint32
	Offset uint32
	Size uint32
}

type ToCWwiseSoundbank struct {
	ToCDataSize uint32 `json:"ToCDataSize"`
	ToCFileId uint64 `json:"ToCFileId"`
	ToCDataOffset uint64 `json:"ToCDataOffset"`
	PathName string `json:"PathName"`
	RawData []byte `json:"-"`
}

func (w *ToCWwiseSoundbank) exportToCHeader() error { var filename string
	if w.PathName == "" {
		filename = fmt.Sprintf("%d", w.ToCFileId)
	} else {
		filename = w.PathName
	}

	headerFile, err := os.OpenFile(
		filename + ".json", os.O_WRONLY | os.O_CREATE, 0644,
	)
	defer headerFile.Close()

	if err != nil {
		return err
	}

	ToCMetaData, err := json.MarshalIndent(w, "", "    ")
	if err != nil {
		return errors.Join(
			errors.New("Failed to JSON encode ToC header information"),
			err)
	}

	if _, err = headerFile.Write(ToCMetaData); err != nil {
		return errors.Join(errors.New("Failed to write JSON blob into file"), 
		err)
	}

	return nil
}

func (w *ToCWwiseSoundbank) exportWwiserXML(removeBinaryFile bool) error {
	w.RawData[0x08] = 0x8D
	w.RawData[0x09] = 0x00
	w.RawData[0x0A] = 0x00
	w.RawData[0x0B] = 0x00

	var filename string

	if w.PathName == "" {
		filename = fmt.Sprintf("%d", w.ToCFileId)
	} else {
		filename = w.PathName
	}
    // filename = path.Join("xmls", filename)

    // _, err := os.Stat("xmls")
    // if err != nil {
    //     if errors.Is(err, os.ErrNotExist) {
    //         if err := os.Mkdir("xmls", 0644); err != nil {
    //             return err
    //         }
    //     } else {
    //         return err
    //     }
    // }
    
	bankFile, err := os.OpenFile(
		filename, os.O_WRONLY | os.O_CREATE, 0644,
	)
	if err != nil {
        if errors.Is(err, os.ErrExist) {
            panic("Revisiting soundbank since there are name confliction")
        }

		return errors.Join(
			errors.New("Failed to create a empty file for Wwise soundbank"),
			err,
		)
	}
	
	if _, err := bankFile.Write(w.RawData); 
	err != nil {
		bankFile.Close()
		return errors.Join(
			errors.New("Failed to write Wwise soundbank raw data"),
			err,
		)
	}
	
	if err = bankFile.Close(); err != nil {
		return errors.Join(
			errors.New("Failed to close Wwise soundbank file"), 
			err,
		)
	}
	
	cmd := exec.Command("python", "wwiser.pyz", filename)
	if err = cmd.Run(); err != nil {
		return errors.Join(
			errors.New("Failed to generate Wwiser XML output"), 
			err)
	}

    if removeBinaryFile {
        os.Remove(filename)
    }

	return nil
}

func (w *ToCWwiseSoundbank) deleteRawData() {
    w.RawData = nil
}

type WwiseSoundbankDep struct {
	Id uint64
	PathName string
}

func (t *ToCFile) toJSON() (string, error) {
    b, err := json.Marshal(t)
    if err != nil {
        return "", err
    }
    return string(b), nil
}
