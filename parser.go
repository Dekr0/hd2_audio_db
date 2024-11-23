package main

import (
	"encoding/xml"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
)

func matchCAkObjElement(e *xml.StartElement, name string) (bool, error) {
	if len(e.Attr) < 1 {
		return false, errors.New("Mismatch # of XML attribute for CAkObj")
	}
	if e.Attr[0].Name.Local != "na" {
		return false, errors.New("CAkObject without `na`")
	}
	return e.Attr[0].Value == name, nil
}

func parseCAkFldElement(e *xml.StartElement) (*CAkFldElement, error) {
	if len(e.Attr) < 3 {
		return nil, errors.New("Mismatch # of XML attribute for CAkFld")
	}
	if e.Attr[0].Name.Local != "ty" {
		return nil, errors.New("First XML attribute is not `ty` for CAkFld")
	}
	if e.Attr[1].Name.Local != "na" {
		return nil, errors.New("Second XML attribute is not `na` for CAkFld")
	}
	if e.Attr[2].Name.Local != "va" {
		return nil, errors.New("Third XML attribute is not `va` for CAkFld")
	}

	return &CAkFldElement{
		xml.Name{ Local: "fld" },
		e.Attr[1].Value,
		e.Attr[0].Value,
		e.Attr[2].Value,
	}, nil
}

func matchCAkFldElement(e *xml.StartElement, ty string, name string) (
	*string, bool, error) {
	fld, err := parseCAkFldElement(e)
	if err != nil {
		return nil, false, err
	}
	if ty != fld.Type || name != fld.Name {
		return nil, false, nil 
	}
	return &fld.Value, true, nil
}

func getCAkObjectULID(decoder *xml.Decoder) (uint32, error) {
	for {
		token, err := decoder.Token()
		if err != nil {
			return 0, errors.Join(
				errors.New("Failed to obtain ULID of a CAkObject"),
				err)
		}

		switch element := token.(type) {
		case xml.StartElement:
			if value, matched, err := matchCAkFldElement(&element, "sid", "ulID"); 
			err == nil { 
				if !matched {
					break
				}

				ULID, err := strconv.ParseInt(*value, 10, 32)
				if err != nil {
					return 0, errors.Join(
						errors.New("Failed to obtain ULID of a CAkObject"), 
						err)
				}

				return uint32(ULID), nil
			} else {
				return 0, errors.Join(
					errors.New("Failed to obtain ULID of a CAkObject"),
					err)
			}
		}
	}
}

func parseCAkSoundObjectElement(decoder *xml.Decoder) (*CAkSound, error) {
	sound := CAkSound{ 0, 0, 0, false }

	var err error

	logger.Debug("Getting CAkSound ULID")
	sound.ULID, err = getCAkObjectULID(decoder)
	if err != nil {
		return nil, err 
	}
	logger.Debug("CAkSound ULID obtained")

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err 
		}

		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local != "fld" {
				break
			}

			if value, matched, err := matchCAkFldElement(&element, "tid", "sourceID"); 
			err == nil {
				if !matched {
					break
				}

				sourceID, err := strconv.ParseInt(*value, 10, 32)
				if err != nil {
					return nil, err
				}
				sound.SourceID = uint32(sourceID)
				
				return &sound, nil
			} else {
				return nil, err
			}
		}
	}
}

func parseCAkRanSeqCntrObjectElement(decoder *xml.Decoder) (
	*CAkRanSeqCntr, error) {
	var err error

	cntr := CAkRanSeqCntr{ 0, make(map[uint32]*CAkSound), 0 }

	logger.Debug("Getting CAkRanSeqCntr ULID")
	cntr.ULID, err = getCAkObjectULID(decoder)
	if err != nil {
		return nil, err 
	}
	logger.Debug("CAkRanSeqCntr ULID obtained")

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local != "obj" {
				break
			}

			if matched, _ := matchCAkObjElement(&element, "Children"); matched {
				var root CAkChildrenElement

				/** Switch to DOM Mode */
				err = decoder.DecodeElement(&root, &element)
				if err != nil {
					return nil, err
				}

				numChilds, err := strconv.Atoi(root.Nodes[0].Value)
				if err != nil {
					return nil, errors.New("Failed to obtain number of children")
				}

				if numChilds != len(root.Nodes) - 1 {
					return nil, errors.New("Number of children is mismatched " +
					"with the parsing result")
				}

				for _, c := range root.Nodes[1:] {
					if c.Type != "tid" || c.Name != "ulChildID" {
						return nil, errors.New("Malformed children entry")
					}

					ULID, err := strconv.Atoi(c.Value)
					if err != nil {
						return nil, errors.New("Failed to convert ULID to int")
					}

					if _, in := cntr.CAkSounds[uint32(ULID)]; in {
						return nil, errors.New("Duplicate children")
					}

					cntr.CAkSounds[uint32(ULID)] = nil
				}
				
				return &cntr, nil
			}
		}
	}
}

func parseHircChunkXML(decoder *xml.Decoder) (*CAkHirearchy, error) {
	hirearchy := &CAkHirearchy{
		CAkObjects: make(map[uint32]CAkObject),
		ReferencedSounds: make(map[uint32]*CAkSound),
		Sounds: make(map[uint32]*CAkSound),
		RanSeqCntrs:  make(map[uint32]*CAkRanSeqCntr),
	}

	T:
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		switch element := token.(type) {
		case xml.StartElement:
			if matched, _ := matchCAkObjElement(&element, "CAkSound"); matched {
				logger.Debug("Parsing CAkSound")

				sound, err := parseCAkSoundObjectElement(decoder)
				if err != nil {
					return nil, errors.Join(
						errors.New("Failed to parse CAkSound"),
						err)
				}

				if _, in := hirearchy.Sounds[sound.ULID]; in {
					return nil, errors.New("Duplicated CAkSound ULID")
				}
				hirearchy.CAkObjects[sound.ULID] = sound
				hirearchy.Sounds[sound.ULID] = sound

				break
			}

			if matched, _ := matchCAkObjElement(&element, "CAkRanSeqCntr"); matched {
				logger.Debug("Parsing CAkRanSeqCntr")

				cntr, err := parseCAkRanSeqCntrObjectElement(decoder)
				if err != nil {
					return nil, errors.Join(
						errors.New("Failed to parse CAkRanSeqCntr"),
						err)
				}

				if _, in := hirearchy.RanSeqCntrs[cntr.ULID]; in {
					return nil, errors.New("Duplicated CAkRanSeqCntr ULID")
				}
				hirearchy.CAkObjects[cntr.ULID] = cntr
				hirearchy.RanSeqCntrs[cntr.ULID] = cntr

				break
			}
		case xml.EndElement:
			if element.Name.Local == "root" {
				break T
			}
		}
	}

	/**
	 * Group all CAkSound object references into its parent CAkRanSeqCntr 
	 * */
	count := 0
	for cntrULID, cntr := range hirearchy.RanSeqCntrs {
		for soundULID := range cntr.CAkSounds {
			sound, in := hirearchy.Sounds[soundULID]
			if !in {
				if _, in := hirearchy.ReferencedSounds[soundULID]; in {
					return nil, errors.New("Duplicated cross-shared Sound " +
					"object ULID")
				}
				hirearchy.ReferencedSounds[soundULID] = nil
			} else {
				sound.DirectParentID = cntrULID
				cntr.CAkSounds[soundULID] = sound
			}
			count++
		}
	}

	logger.Info("Total Sound objects of all random sequence containers", 
		"Total", count)

	return hirearchy, nil
}

func ParseWwiseSoundBankXML(f io.Reader) error {
	decoder := xml.NewDecoder(f)

	bank := &CAkWwiseBank{ &CAkMediaIndex{ 0 }, nil }

	for {
		t, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		switch v := t.(type) {
		case xml.StartElement:
			/** Validation check */
			if matched, _ := matchCAkObjElement(&v, "MediaHeader"); matched {
				bank.MediaIndex.Count++
				break
			}

			if matched, _ := matchCAkObjElement(&v, "HircChunk"); matched {
				logger.Debug("Parsing HircChunk")
				bank.Hirearchy, err = parseHircChunkXML(decoder)
				if err != nil {
					return errors.Join(errors.New("Failed to parsed HircChunk"), 
					err)
				}

				// b, err := json.MarshalIndent(hirearchy, "", "    ")
				// if err == nil {
				// 	fmt.Println(string(b))
				// }

				logger.Info("Parsing Result", 
					"MediaIndexCount", bank.MediaIndex.Count,
					"ReferencedSounds", len(bank.Hirearchy.ReferencedSounds),
					"SoundObjectCount", len(bank.Hirearchy.Sounds),
					"RanSeqCntrsCount", len(bank.Hirearchy.RanSeqCntrs),
				)

				return nil
			}
		}
	}

	return nil
}

func WwiserOuputParsing(xmlArg *string) error {
	xmls := strings.Split(*xmlArg, ",")

	for _, x := range xmls {
		r, err := os.Open(x)
		if err != nil {
			logger.Warn("Parsing error", "file", x, "error", err)
			continue
		}
		err = ParseWwiseSoundBankXML(r)
		if err != nil {
			logger.Warn("Parsing error", "file", x, "error", err)
			continue
		}
	}

	return nil
}
