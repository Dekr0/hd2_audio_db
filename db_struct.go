package main

var CSV_DIR string

type DBToCWwiseSoundbank struct {
    dbid string
    ref *ToCWwiseSoundbank
	linkedGameArchiveIds map[string]Empty `json:"-"`
}

type DBCAkObject struct {
    dbid string
    ref CAkObject
    linkedSoundbankPathNames map[string]Empty
}

type HireachyResult struct {
    uniqueObjs map[uint32]*DBCAkObject
    uniqueSounds map[uint32]*CAkSound
    uniqueCntrs map[uint32]*CAkRanSeqCntr
}

type BankExtractTaskResult struct {
    gameArchiveID string
    result map[uint64]*ToCWwiseSoundbank
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

type GameArchiveCSVTaskResult struct {
    filename string
    result *GameArchiveCSVResult
    err error
}

type WwiserXMLTaskResult struct {
    pathName string
    result *CAkWwiseBank
    err error
}
