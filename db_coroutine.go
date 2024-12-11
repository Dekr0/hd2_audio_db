package main

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"path"
	"strings"
)

// A go coroutine function (Soundbank extraction of a given archive)
//
// [channel passing]
// *ExtractBankResult - Never nil
//   - result - nil only when openToCFile error or extractWwiseSoundbanks error 
func extractSoundbankTask(gameArchiveID string, c chan *BankExtractTaskResult) {
	payload := &BankExtractTaskResult{ gameArchiveID: gameArchiveID }
	tocFile, err := openToCFile(gameArchiveID)
	if err != nil {
		payload.err = err
		c <- payload
		return
	}
	defer tocFile.Close()

	payload.result, payload.err = extractWwiseSoundbanks(*tocFile, true)
	c <- payload
}

// Game archives spreadsheet parsing
//
// [return]
// *GameArchiveCSVResult - Nil when an error occurs
// error - trivial
func parseGameArchiveCSV(filename string) (*GameArchiveCSVResult, error) {
	splits := strings.Split(path.Base(filename), ".")

	if len(splits) <= 1 {
        return nil, errors.New("Invalid input file: missing file extension") 
    }

    if splits[1] != "csv" {
        return nil, errors.New("Invalid input file: not a CSV file")
    }

    category := splits[0]

	csvFileHandle, err := os.Open(path.Join(CSV_DIR, filename))
    if err != nil {
        return nil, err
    }
    defer csvFileHandle.Close()

    payload := newGameArchiveCSVResult()

    reader := csv.NewReader(csvFileHandle)
    for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
                return payload, nil
			}

            return nil, err
		}

        payload.perFileLineRead++

        tag, gameArchiveIDs := record[0], record[1:]
        for _, gameArchiveID := range gameArchiveIDs {
            if gameArchiveID == "" {
                continue
            }

			if gameArchive, in := payload.uniqueGameArchives[gameArchiveID];
            in {
				if _, in = gameArchive.uniqueCategories[category]; !in {
					gameArchive.uniqueCategories[category] = Empty{}
				}
				if _, in = gameArchive.uniqueTags[tag]; !in {
					gameArchive.uniqueTags[tag] = Empty{}
				}
				continue
			}

            gameArchive := NewGameArchive(gameArchiveID)

			gameArchive.uniqueCategories[category] = Empty{}
			gameArchive.uniqueTags[tag] = Empty{} 

			payload.uniqueGameArchives[gameArchiveID] = gameArchive
        }

        payload.perFileLineParsed++
    }
}

// A go coroutine function (Parsing game archive spreadsheet)
//
// [channel passing]
// - *GameArchiveCSVTaskResult - Never nil
//   - GameArchiveCSVResult - Nil when an error occur when parseGameArchiveCSV 
//   errors
func parseGameArchiveCSVTask(
    filename string, c chan *GameArchiveCSVTaskResult) {
    payload := &GameArchiveCSVTaskResult{}
    defer func() { c <- payload }()

    payload.filename = filename
    payload.result, payload.err = parseGameArchiveCSV(filename)
}

// [return]
// *CAkWwiseBank - Nil when an error occurs
// error - trivial
func exportParseWwiserXML(pathName string, bank *DBToCWwiseSoundbank) (
    *CAkWwiseBank, error) {
    if bank == nil {
        panic("Assertion failure. Receive a Nil DB wrapper of ToC encasuplated " +
        "Wwise Soundbank before exporting Wwiser XML")
    }

    if bank.ref == nil {
        panic("Assertion failure. Receive a Nil ToC encasuplatedWwise Soundbank" +
        " before exporting Wwiser XML")
    }

    err := bank.ref.exportWwiserXML(true)
    if err != nil {
        return nil, err
    }
    bank.ref.deleteRawData()

    filename := pathName + ".xml"
    // filename = path.Join("xmls", filename)
    xmlFile, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer xmlFile.Close()

    return parseWwiseSoundBankXML(xmlFile)
}

// A go coroutine function (exporting and parsing Wwiser XML)
//
// [channel passing]
// *WwiserXMLTaskResult - Never nil
//   result - Nil when exportParserWwiserXML errors
func exportParseWwiserXMLTask(
    pathName string, bank *DBToCWwiseSoundbank, c chan *WwiserXMLTaskResult,
) {
    payload := &WwiserXMLTaskResult{ pathName: pathName }
    defer func() { c <- payload }()

    payload.result, payload.err = exportParseWwiserXML(pathName, bank)
}