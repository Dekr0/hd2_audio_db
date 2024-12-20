package main

import (
	// "encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

/*
Decoding logic when a StartElement <object name="..." index="..."> is encountered
Attribute `name` and `index` is not necessary always there for every StartElement
<object>. `name` typical occurs for every StartElement <object>

[return]
*CAkObjElement - a pointer of struct that encapsulate <object name="..." index="...">
Nil when an error occurs.
error - trivial
*/
func decodeCAkObjectStartElement(e *xml.StartElement) (*CAkObjectElement, error) {
    if e.Name.Local != "object" {
        return nil, errors.New("This is not a <object> element.")
    }

    object := &CAkObjectElement{
        XMLName: xml.Name{ Space: "", Local: "object" },
    }
    for _, a := range e.Attr {
        switch a.Name.Local {
        case "name":
            object.Name = a.Value
        case "index":
            object.Index = a.Value
        }
    }
    
    return object, nil
}

/*
Decoding logic when a StartElement <field type="..." name="..." value="..." 
valuefmt="..."> is encountered. Not all attributes are always there for every 
StartElement <field>. `type`, `name`, and `value` typical occur for every 
StartElement <field>

[return]
*CAkObjElement - a pointer of struct that encapsulate <field ...>. Nil when an 
error occurs.
error - trivial
*/
func decodeCAkFldStartElement(e *xml.StartElement) (*CAkFieldElement, error) {
    if e.Name.Local != "field" {
        return nil, errors.New("This is not a <field> element")
    }

    field := &CAkFieldElement{
        XMLName: xml.Name{ Space: "", Local: "field" },
    }

    for _, a := range e.Attr {
        switch a.Name.Local {
        case "name":
            field.Name = a.Value
        case "type":
            field.Type = a.Value
        case "value":
            field.Value = a.Value
        case "valuefmt":
            field.ValueF = a.Value
        }
    }

    return field, nil
}

// Parsing logic for obtaining the ULID (uint32) for a CAkObject.
//
// [return]
// uint32 - ULID of a CAkObject.
// error - trivial
func getCAkObjectULID(decoder *xml.Decoder) (uint32, error) {
	for {
		token, err := decoder.Token()
		if err != nil {
			return 0, err
        }

		switch element := token.(type) {
		case xml.StartElement:
            if element.Name.Local != "field" {
                continue
            }

            field, err := decodeCAkFldStartElement(&element)
            if err != nil {
                return 0, err
            }

            if field.Type == "sid" && field.Name == "ulID" {
                v, err := strconv.ParseInt(field.Value, 10, 32)
                if err != nil {
                    return 0, err
                }

                return uint32(v), nil
            }
		}
	}
}

/*
Parsing logic for when a StartElement <object na="CAkRanSeqCntr"> is encounted.
It only extract some basic information and all the ULID of its children 
CAkobjects.

[return]
*CAkRanSeqCntr - a pointer to a struct that encapsulates partial information 
about CAkRanSeqCntr. Nil only when object ULID cannon be obtained
error - trivial
*/
func parseCAkRanSeqCntrObjectElement(decoder *xml.Decoder) (
	*CAkRanSeqCntr, error) {
	var err error

	cntr := CAkRanSeqCntr{ 0, 0, 0, make(map[uint32]*CAkSound) }

	cntr.ObjectULID, err = getCAkObjectULID(decoder)
	if err != nil {
		return nil, errors.Join(
            errors.New("Failed to obtain CAkRanSeqCntr ULID"),
            err,
        )
	}

    /* 
    Once <object na="Children"> is parsed, it will immedately return.
    <field na="DirectParentID"> typically occurs before <object name="Children">
     */
	for {
		token, err := decoder.Token()
		if err != nil {
			return &cntr, err
		}

		switch element := token.(type) {
		case xml.StartElement:
            if element.Name.Local == "field" {
                field, err := decodeCAkFldStartElement(&element)
                if err != nil {
                    return &cntr, nil
                }

                if field.Type != "tid" || field.Name != "DirectParentID" {
                    continue
                }

                v, err := strconv.ParseInt(field.Value, 10, 32)
                if err != nil {
                    return &cntr, err
                }

                cntr.DirectParentObjectULID = uint32(v)

                break
            }

            if element.Name.Local == "object" {
                object, err := decodeCAkObjectStartElement(&element)
                if err != nil {
                    return &cntr, err
                }

                if object.Name != "Children" {
                    continue
                }

			    var children CAkChildrenElement

			    err = decoder.DecodeElement(&children, &element)
			    if err != nil {
			    	return &cntr, err
			    }

			    numChilds, err := strconv.Atoi(children.Nodes[0].Value)
			    if err != nil {
			    	return &cntr, errors.New("Failed to obtain number of children")
			    }

			    if numChilds != len(children.Nodes) - 1 {
			    	return &cntr, errors.New("Number of children is mismatched " +
			    	"with the parsing result")
			    }

			    for _, c := range children.Nodes[1:] {
			    	if c.Type != "tid" || c.Name != "ulChildID" {
			    		return &cntr, errors.New("Malformed children entry")
			    	}

			    	ULID, err := strconv.Atoi(c.Value)
			    	if err != nil {
			    		return &cntr, errors.New("Failed to convert object ULID to int")
			    	}

			    	if _, in := cntr.CAkSounds[uint32(ULID)]; in {
			    		return &cntr, errors.New("Duplicate children")
			    	}

			    	cntr.CAkSounds[uint32(ULID)] = nil
			    }
			    
			    return &cntr, nil
            }
		}
	}
}

/*
Parsing logic for when a StartElement <object na="CAkSound"> is encounted.
It only extract some basic information (ULID, source ID, and direct parent ULID)

[return]
*CAkSound - a pointer to a struct that encapsulates partial information about 
CAkSound. Nil only when the object ULID cannot be obtained.
error - trivial
*/
func parseCAkSoundObjectElement(decoder *xml.Decoder) (*CAkSound, error) {
	sound := CAkSound{ false, 0, 0, 0, make(map[uint32]Empty) }

	var err error

	sound.ObjectULID, err = getCAkObjectULID(decoder)
	if err != nil {
		return nil, errors.Join(
            errors.New("Failed to obtain CAkSound ULID"),
            err,
        )
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			return &sound, err 
		}

		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local != "field" {
				break
			}

            field, err := decodeCAkFldStartElement(&element)
            if err != nil {
                return &sound, err
            }

            if field.Type != "tid" {
                continue
            }

            if field.Name == "sourceID" {
                v, err := strconv.ParseInt(field.Value, 10, 32)
                if err != nil {
                    return &sound, err
                }
                sound.ShortIDs[uint32(v)] = Empty{}
                break
           } 

           if field.Name == "DirectParentID" {
               v, err := strconv.ParseInt(field.Value, 10, 32)
               if err != nil {
                   return &sound, err
               }
               sound.DirectParentObjectULID = uint32(v)
               return &sound, nil
           }
        }
    }
}

/*
Parsing logic when a StartElement <object na="HircChunk"> is encounted

[parameter]
decoder - a reference to StartElement.

[return]
*CAkHierarchy - a pointer to a struct that encapsulates partial information 
about Wwise Hierarchy. Nil when an error is occured.
error - trivial
*/
func parseHircChunkXML(decoder *xml.Decoder) (*CAkHierarchy, error) {
	hierarchy := &CAkHierarchy{
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
            if element.Name.Local != "object" {
                continue
            }

            object, err := decodeCAkObjectStartElement(&element)
            if err != nil {
                return nil, err
            }

            if object.Name == "CAkSound" {
				sound, err := parseCAkSoundObjectElement(decoder)
				if err != nil {
                    errMsg := "Failed to parse CAkSound object element."
                    if sound != nil {
                        errMsg += fmt.Sprintf("Object ULID: %d", sound.ObjectULID)
                    }
					return nil, errors.Join(errors.New(errMsg), err)
				}

				if _, in := hierarchy.Sounds[sound.ObjectULID]; in {
                    errMsg := fmt.Sprintf("Duplicated CAkSound object ULID: %d", 
                    sound.ObjectULID)
					return nil, errors.New(errMsg)
				}

                v, err := strconv.ParseInt(object.Index, 10, 32)
                if err != nil {
                    errMsg := fmt.Sprintf("Failed to obtain index for CAkSound. " + 
                    "ObjectULID: %d.", sound.ObjectULID)

                    return nil, errors.Join(errors.New(errMsg), err)
                }

                sound.ObjectIndex = int32(v)
				hierarchy.CAkObjects[sound.ObjectULID] = sound
				hierarchy.Sounds[sound.ObjectULID] = sound

				break
            }

            if object.Name == "CAkRanSeqCntr" {
				cntr, err := parseCAkRanSeqCntrObjectElement(decoder)
				if err != nil {
                    errMsg := "Failed to parse CAkRanSeqCntr object element."
                    if cntr != nil {
                        errMsg += fmt.Sprintf("Object ULID: %d", cntr.ObjectULID)
                    }
					return nil, errors.Join(errors.New(errMsg), err)
				}

				if _, in := hierarchy.RanSeqCntrs[cntr.ObjectULID]; in {
                    errMsg := fmt.Sprintf("Duplicated CAkRanSeqCntr object ULID: %d", 
                    cntr.ObjectULID)
					return nil, errors.New(errMsg)
				}

                v, err := strconv.ParseInt(object.Index, 10, 32)
                if err != nil {
                    errMsg := fmt.Sprintf("Failed to obtain index for CAkRanSeqCntr. " + 
                    "ObjectULID: %d.", cntr.ObjectULID)
                    return nil, errors.Join(errors.New(errMsg), err)
                } 

                cntr.ObjectIndex = int32(v)
				hierarchy.CAkObjects[cntr.ObjectULID] = cntr
				hierarchy.RanSeqCntrs[cntr.ObjectULID] = cntr

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
	for cntrULID, cntr := range hierarchy.RanSeqCntrs {
        if cntr == nil {
            panic("Assertion failed. Nil Random / Sequence object when " +
            "grouping Sound objects into Random / Sequence Containers")
        }

		for soundULID := range cntr.CAkSounds {
			sound, in := hierarchy.Sounds[soundULID]
			if !in {
				if _, in := hierarchy.ReferencedSounds[soundULID]; in {
					return nil, errors.New("Duplicated referenced Sound " +
					"object ULID")
				}
				hierarchy.ReferencedSounds[soundULID] = nil
                continue
			}

            if sound == nil {
                panic("Assertion failed. Nil Sound object when grouping " + 
                "Sound objects into Random / Sequence Containers")
            }

            if sound.DirectParentObjectULID != cntrULID {
                errMsg := fmt.Sprintf("CAkSound %d has a parent ULID of %d" +
                "but CAkRanSeqCntr contain this CAkSound has a ULID of %d",
                sound.ObjectULID, sound.DirectParentObjectULID, cntrULID,
                )
                return nil, errors.New(errMsg)
            }
			sound.DirectParentObjectULID = cntrULID
			cntr.CAkSounds[soundULID] = sound
		}
	}

	return hierarchy, nil
}


// Main entry point for parsing a single Wwiser XML file. CAkWwiseBank.Hierarchy
// can be nil if a Wwise Soundbank does not contain HirchChunk section.
// 
// [parameter]
// f io.Reader - A io.Reader to a Wwsier XML file
// 
// [return]
//
// CAkWwiseBank - a pointer to a struct that encapsulates partial information 
// about Wwise Soundbank. Nil when an error occur.
// error - trivial 
func parseWwiseSoundBankXML(f io.Reader) (*CAkWwiseBank, error) {
	decoder := xml.NewDecoder(f)

	bank := &CAkWwiseBank{ &CAkMediaIndex{ 0 }, nil }

	for {
		t, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
                return nil, errors.New("A potential malformed Soundbank XML or " + 
                "a Soundbank without HircChunk")
			}
			return nil, err
		}
		switch element := t.(type) {
		case xml.StartElement:
			/** Validation check */
            if element.Name.Local != "object" {
                continue
            }

            obj, err := decodeCAkObjectStartElement(&element)
            if err != nil {
                return nil, err
            }

            if obj.Name == "MediaHeader" {
                bank.MediaIndex.Count++
                break
            }

            if obj.Name == "HircChunk" {
				bank.Hierarchy, err = parseHircChunkXML(decoder)
				if err != nil {
					return nil, errors.Join(
                        errors.New("Failed to parsed HircChunk object element"), 
                        err)
				}

                return bank, nil
            }
		}
	}
}

// Entry point for parsing Wwiser XML files. Locate all given and existing XML 
// files and perform parsing.
// 
// [parameter]
// xmlsArg - a list of XML files name separated by ","
//
// [return]
// error - trivial
func parseWwiserXML(xmlsArg string) error {
	xmls := strings.Split(xmlsArg, ",")

	for _, xml := range xmls {
		r, err := os.Open(xml)
		if err != nil {
			DefaultLogger.Warn("Wwiser XML parsing error", "file", xml, "error", err)
			continue
		}

        bank, err := parseWwiseSoundBankXML(r)
		if err != nil {
			DefaultLogger.Warn("Wwiser XML parsing error", "file", xml, "error", err)
			continue
		}

        /*
        output, err := json.MarshalIndent(bank, "", "    ")
        if err != nil {
            DefaultLogger.Error("Failed to generate JSON Wwise soundbank",
                "file", xml,
                "error", err,
            )
            continue
        }
        */

		DefaultLogger.Info("Parsing Result", 
			"mediaIndexCount", bank.MediaIndex.Count,
			"referencedSounds", len(bank.Hierarchy.ReferencedSounds),
			"soundObjectCount", len(bank.Hierarchy.Sounds),
			"ranSeqCntrsCount", len(bank.Hierarchy.RanSeqCntrs),
		)

        // fmt.Println(string(output))
	}

	return nil
}
