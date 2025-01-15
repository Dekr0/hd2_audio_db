package db

import (
    "dekr0/hd2_audio_db/toc"
    "dekr0/hd2_audio_db/wwise"
)

var CSV_DIR string

type LabelFile struct {
    Sources map[string]string `json:"sources"`
}

type DBToCWwiseSoundbank struct {
    dbID string
    ref *toc.ToCWwiseSoundbank
	gameArchiveDbIDs map[string]Empty `json:"-"`
}

type DBCAkObject struct {
    dbid string
    ref wwise.CAkObject
    soundbankDbIDs map[string]Empty
}

type BankExtractTaskResult struct {
    gameArchiveDbID string
    gameArchiveID string
    result map[uint64]*toc.ToCWwiseSoundbank
    err error
}

type GameArchiveCSVTaskResult struct {
    filename string
    result *GameArchiveCSVResult
    err error
}

type HierarchyResult struct {
    uniqueObjs map[uint32]*DBCAkObject
    uniqueSounds map[uint32]*wwise.CAkSound
    uniqueCntrs map[uint32]*wwise.CAkRanSeqCntr
}

type WwiserXMLTaskResult struct {
    pathName string
    soundbankDbID string
    result *wwise.CAkWwiseBank
    err error
}

type GameArchive struct {
    gameArchiveID string
    uniqueTags map[string]Empty
    uniqueCategories map[string]Empty
}

func NewGameArchive(gameArchiveID string) *GameArchive {
    return &GameArchive{
        gameArchiveID: gameArchiveID,
		uniqueTags: make(map[string]Empty),
		uniqueCategories: make(map[string]Empty),
    }
}

type GameArchiveCSVResult struct {
    uniqueGameArchives map[string]*GameArchive
    perFileLineParsed uint32
    perFileLineRead uint32
}

func newGameArchiveCSVResult() *GameArchiveCSVResult {
    return &GameArchiveCSVResult{ make(map[string]*GameArchive), 0, 0 }
}
