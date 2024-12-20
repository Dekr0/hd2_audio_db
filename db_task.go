package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"os"
	"strconv"
	"strings"

	"dekr0/hd2_audio_db/internal/database"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

/*
A helper function that find the primary key for a record that contain the
string representation of a Wwise hierarchy object type. Error only arise when
a type doesn't exist or *t is nil

[parameter]
rows - rows of database record about Wwise hierarchy object types
*t - a string representation of a Wwiser hierarchy object type

[return]
string - primary key for a record whose string representaion of a Wwise.
hierarchy object type matches the one provided by *t. Empty string when an
error occurs.
error - trivial
*/
func getHierarchyObjectTypePrimaryKey(
    rows []database.HierarchyObjectType, t string,
) (string, error) {
    for _, row := range rows {
        if row.Type == t {
            return row.DbID, nil
        }
    }
    return "", errors.New("Type " + t + "does not exist")
}

/*
A helper function for generate a Database primary key using UUID4

[return]
string - Return Empty string when an error occurs
error - trivial
*/
func genDBID() (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	
	idS := id.String()
	if idS == "" {
		return "", errors.New("Empty string when generate UUID4 string")
	}

	return idS, nil
}

func getDBConn() (*sql.DB, error) {
	godotenv.Load()

	dbPath := os.Getenv("DB_PATH")
	db, err := sql.Open("sqlite3", dbPath)
    return db, err
}

/*
Rewriting all database records about Helldivers 2 game archives. It will 
completely all database records previously written. The run through of this 
is parsing every .csv file in the ./csv/archives/ folder. Each .csv file is 
named after a category associated with the archives that file contains.
Each .csv file following this format
- first column is always a tag that is assigned to all game archievs in a 
given row
- the rest of the columns are game archive IDs
Since an game archive can contain multiple Wwise Soundbank (e.g. e75f556a740e00c9),
, a game archive can have more than one tag, and can have more than one 
category.
*/
func rewriteAllGameArchivesSpreadsheet(dir string, ctx context.Context) error {
	if DefaultLogger == nil {
		return errors.New("Logger cannot be nil")
	}

	csvFiles, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

    db, err := getDBConn()
	defer db.Close()

	dbQueries := database.New(db)

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queriesWithTx := dbQueries.WithTx(tx)
	if err = queriesWithTx.DeleteAllGameArchive(ctx); err != nil {
		return err
	}

    // Report checking
	totalLineRead := 0
    totalLineParsed := 0
	totalFileParsed := 0
    errorCSVs := make(map[string]error)
    // End

    // For sync. 
    // Todo: Add context cancellation for go coroutine if main thread encouter 
    // a fatal error.
	uniqueGameArchives := make(map[string]*GameArchive)
    resultChannel := make(chan *GameArchiveCSVTaskResult)
    finishedTasks := 0
    for _, csvFile := range csvFiles {
        go parseGameArchiveCSVTask(csvFile.Name(), resultChannel)
    }

    for finishedTasks < len(csvFiles) {
        select {
        case payload := <- resultChannel:
            if payload == nil {
                panic("Assertion failure. Receive a nil game archive CSV task " +
                "result")
            }

            finishedTasks++
            if payload.err != nil {
                if _, in := errorCSVs[payload.filename]; !in {
                    errorCSVs[payload.filename] = payload.err
                } else {
                    DefaultLogger.Error("Visit the same file more than once", 
                    "error", err)
                }
                break
            }

            if payload.result == nil {
                panic("Assertion failure. Receive a nil game archive CSV result")
            }
            
            filename := payload.filename
            result := payload.result

            for key, val := range result.uniqueGameArchives {
                gameArchive, in := uniqueGameArchives[key]
                if !in {
                    uniqueGameArchives[key] = val
                    continue
                }

                for tag := range val.uniqueTags {
                    if _, in := gameArchive.uniqueTags[tag]; !in {
                        gameArchive.uniqueTags[tag] = Empty{}
                    }
                }
                for category := range val.uniqueCategories {
                    if _, in := gameArchive.uniqueCategories[category]; !in {
                        gameArchive.uniqueCategories[category] = Empty{}
                    }
                }
            }

            totalLineParsed += int(result.perFileLineParsed)
            totalLineRead += int(result.perFileLineRead)

            totalFileParsed++

            DefaultLogger.Info("Finished parsing one file",
                "file", filename,
                "perFileLineParsed", result.perFileLineParsed,
                "perFileLineRead", result.perFileLineRead,
                "totalLineParsed", totalLineParsed,
                "totalLineRead", totalLineRead,
                "totalFileParsed", totalFileParsed,
            )
        }
    }

    close(resultChannel)
	DefaultLogger.Info("Finished parsing all files",
        "totalLineParsed", totalFileParsed,
		"totalLineRead", totalLineRead,
		"totalFileParse", totalFileParsed,
	)

    DefaultLogger.Error("Error Report", "numErrorCSVFiles", len(errorCSVs))
    for filename, err := range errorCSVs {
        DefaultLogger.Error(filename, "error", err)
    }
    // End


    // Writing game archive records.
    totalRows := 0
	for _, gameArchive := range uniqueGameArchives {
		categories := []string{} 
		for category := range gameArchive.uniqueCategories {
			categories = append(categories, category)
		}
		tags := []string{}
		for tag := range gameArchive.uniqueTags {
			tags = append(tags, tag)
		}

        dbId, err := genDBID()
        if err != nil {
            return errors.Join(
                errors.New("Failed to generate UUID4"),
                err,
            )
        }

		if err = queriesWithTx.CreateGameArchive(
			ctx,
			database.CreateGameArchiveParams{
				DbID: dbId,
				GameArchiveID: gameArchive.gameArchiveID,
				Tags: strings.Join(tags, ";"),
				Categories: strings.Join(categories, ";"),
			},
		); err != nil {
			return err
		}
		totalRows += 1
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	DefaultLogger.Info("Finished rewriting all game archives ID",
		"totalRows", totalRows,
	)
    // end

	return nil
}

func rewriteAllGameArchives(ctx context.Context) error {
    if DefaultLogger == nil {
        return errors.New("Logger cannot be nil")
    }

    db, err := getDBConn()
    if err != nil {
        return err
    }
    defer db.Close()

    files, err := os.ReadDir(DATA)
    if err != nil {
        return err
    }

    dbQueries := database.New(db)

    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    queriesWithTx := dbQueries.WithTx(tx)
    if err = queriesWithTx.DeleteAllGameArchive(ctx); err != nil {
        return err
    }

    totalRows := 0

    type ErrorGameArchive struct {
        filename string
        err error
    }

    errorGameArchive := []ErrorGameArchive{}
    for _, file := range files {
        /* Ignore patches */
        if strings.Contains(file.Name(), "patch") {
            continue
        }
        if !strings.Contains(file.Name(), "stream") {
            continue
        }


        splits := strings.Split(file.Name(), ".")
        if len(splits) == 1 {
            errorGameArchive = append(
                errorGameArchive, 
                ErrorGameArchive{
                    file.Name(), 
                    errors.New(file.Name() + "is missing stream extension"),
                },
            )
            continue
        }

        gameArchivePath := path.Join(DATA, splits[0])
        _, err := os.Stat(gameArchivePath)
        if errors.Is(err, os.ErrNotExist) {
            errorGameArchive = append(
                errorGameArchive,
                ErrorGameArchive{
                    file.Name(),
                    errors.New(file.Name() + "does not have ToC file"),
                },
            )
            continue
        }

        dbId, err := genDBID()
        if err != nil {
            return errors.Join(errors.New("Failed to generate UUID4"), err)
        }
        if err = queriesWithTx.CreateGameArchive(
            ctx,
            database.CreateGameArchiveParams{
                DbID: dbId, GameArchiveID: splits[0], Tags: "", Categories: "",
            },
        ); err != nil {
            return err
        }
        totalRows += 1
    }

    if err = tx.Commit(); err != nil {
        return err
    }

    DefaultLogger.Info("Finished rewriting all game archives ID")
    DefaultLogger.Info("Info status", "totalRows", totalRows)
    DefaultLogger.Error("Error status")
    for _, f := range errorGameArchive {
        DefaultLogger.Error("Malformed filename", "filename", f)
    }

    return nil
}

/*
Rewriting all database records about all possible hierarchy object type in 
Wwise soundbank. It will completely erase all records previously written.
*/
func rewriteAllHierarchyObjectTypes(ctx context.Context) error {
	if DefaultLogger == nil {
		return errors.New("Logger cannot be nil")
	}

	db, err := getDBConn() 
	if err != nil {
		return err
	}
	defer db.Close()

	dbQueries := database.New(db)

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

    queriesWithTx := dbQueries.WithTx(tx)
    if err = queriesWithTx.DeleteAllHierarchyObjectType(ctx); err != nil {
        return err
    }

    totalRows := 0
    for _, name := range HIERARCHY_TYPE_NAME {
        dbId, err := genDBID()
        if err != nil {
            return errors.Join(
                errors.New("Failed to generate UUID4"),
                err,
            )
        }
        
        err = queriesWithTx.CreateHierarchyObjectType(
            ctx,
            database.CreateHierarchyObjectTypeParams{
                DbID: dbId, 
                Type: name,
            },
        )
        if err != nil {
            return err
        }
        totalRows++
    }

	if err = tx.Commit(); err != nil {
		return err
	}

    DefaultLogger.Info("Update Hierarchy Object Type Stat",
        "numHierarchyType", len(HIERARCHY_TYPE_NAME),
        "totalRows", totalRows,
    )

    return nil
}

func rewriteAllSoundBank(ctx context.Context) error {
	if DefaultLogger == nil {
		return errors.New("Logger cannot be nil")
	}

	db, err := getDBConn() 
	if err != nil {
		return err
	}
	defer db.Close()

    banks, err := extractAllSoundbanks(db, false, ctx)
    if err != nil {
        return err
    }
    if err = writeAllSoundbanks(banks, db, ctx); err != nil {
        return err
    }

    return nil
}

/*
Completely erase all existing database record of Wwise Soundbanks, its 
hierarchy objects, and relationship, and rewriting new one.

General process is the following:
1. `SELECT` all rows in the `helldiver_game_archive`. Extract the game archive 
ID from each row.
2. For each game archive ID, join with the Helldivers 2 game data path (
specified by `HELLDIVER_DATA` environmental variable). This is the path to 
a specific game archive ToC file with the given ID.
3. Parse ToC file of each game archive, and extract all Wwise Soundbanks and 
Wwise Dependencies. Notice that there are might be some game archives that 
no longer exist; but they're still listed in the google spreadsheet.
4. For each Wwise Soundbank, write its record into `helldiver_soundbank`.
5. For each Wwise Soundbank, write its binary data to a file, and pipe them 
through wwiser so that wwiser can generate a XML file.
6. Parse through each XML file, and obtain a `*CAkWwiseBank`
7. Transfer its objects into the hashmap that stores all objects from every 
single Wwise Soundbank in Helldivers 2 uniquely.
*/
func rewriteAllSoundAssets(ctx context.Context) error {
	if DefaultLogger == nil {
		return errors.New("Logger cannot be nil")
	}

	db, err := getDBConn() 
	if err != nil {
		return err
	}
	defer db.Close()

    banks, err := extractAllSoundbanks(db, true, ctx)
    if err != nil {
        return err
    }
    if err = writeAllSoundbanks(banks, db, ctx); err != nil {
        return err
    }

    hierarchyResult, err := extractAllHierarchyObjs(banks, ctx)
    if err != nil {
        return err
    }

    if err = writeAllHierarchyObjs(hierarchyResult, db, ctx); 
    err != nil {
        return err
    }

    return nil
}

func writeAllSoundbanks(
    banks []*DBToCWwiseSoundbank, db *sql.DB, ctx context.Context,
) error {
    dbQueries := database.New(db)

    // Write Sounbank data into database
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queriesWithTx := dbQueries.WithTx(tx)
	if err = queriesWithTx.DeleteAllSoundbank(ctx); err != nil {
		return err
	}

	totalRows := 0
	for _, b := range banks {
		linkedGameArchiveIds := []string{}
		for gid := range b.linkedGameArchiveIds {
			linkedGameArchiveIds = append(linkedGameArchiveIds, gid)
		}
		dbId, err := genDBID()
		if err != nil {
			DefaultLogger.Error(
				"Failed to generate UUID4 for a new game archive",
				"toCFileID", b.ref.ToCFileId,
				"error", err,
			)
			continue
		}
		if err = queriesWithTx.CreateSoundbank(ctx,
			database.CreateSoundbankParams{
				DbID: dbId,
				TocFileID: strconv.FormatUint(b.ref.ToCFileId, 10),
				SoundbankPathName: b.ref.PathName,
				SoundbankReadableName: "",
				Categories: "",
				LinkedGameArchiveIds: strings.Join(linkedGameArchiveIds, ";"),
			},
		); err != nil {
			return err
		}
		totalRows++

        b.dbid = dbId
	}
	if err = tx.Commit(); err != nil {
		return err
	}
    // End
	
	DefaultLogger.Info("Finished rewriting Wwise Sounbank records",
		"totalRows", totalRows,
	)
    return nil
}

/*
Main thread for extracting all Wwise Soundbanks from all game archives. The 
main thread is responsible for organizing unique and duplicate Wwise 
Soundbanks from the result a single go routine return.
Each go routine is responsible to extracting all Wwise Soundbanks of a single 
game archive with a given game archive ID.

[return]
[]*DBToCWwiseSoundbank - nil when an error occur
error - trivial
*/
func extractAllSoundbanks(db *sql.DB, rawData bool, ctx context.Context) (
    []*DBToCWwiseSoundbank, error) {
    dbQueries := database.New(db)

	gameArchives, err := dbQueries.GetAllGameArchives(ctx)
	if err != nil {
		return nil, err
	}

    // Extract Soundbank
    bankChannel := make(chan *BankExtractTaskResult)
    erroredGameArchives := make(map[string]error)
	uniqueBanks := make(map[string]*DBToCWwiseSoundbank)
    // For sync. 
    // Todo: this should be encapsulated into worker pool. Add context 
    // cancellation for go coroutine if main thread encouter a fatal error.
    finishedTasks := 0
    totalBankVisits := 0
    dispatchTasks := 0
    activeWorkers := 0
    for finishedTasks < len(gameArchives) {
        select {
        case payload := <- bankChannel:
            finishedTasks++
            activeWorkers--

            if payload == nil {
                panic("Assertion failure. Received a Nil Soundbank extraction payload")
            }

            gameArchiveID := payload.gameArchiveID

            if payload.err != nil {
                if _, in := erroredGameArchives[gameArchiveID]; in {
                    errMsg := fmt.Sprintf(
                        "Revisited a game archive. Check key constraint. %s", 
                        payload.gameArchiveID,
                    )
                    return nil, errors.New(errMsg)
                } else {
                    erroredGameArchives[payload.gameArchiveID] = payload.err
                }
                break
            }

            if payload.result == nil {
                panic("Assertion failure. Received a Nil Soundbank extraction result")
            }

            totalBankVisits += len(payload.result)
		    for _, bank := range payload.result {
                pathName := &bank.PathName
                
                key := *pathName + ";" + strconv.FormatUint(bank.ToCFileId, 10)
                // existBank, in := uniqueBanks[*pathName]
		    	existBank, in := uniqueBanks[key]
		    	if !in {
		    		// uniqueBanks[*pathName] = &DBToCWwiseSoundbank{
                    //     "",
                    //     bank,
                    //     make(map[string]Empty),
                    // }
                    // uniqueBanks[*pathName].linkedGameArchiveIds[id] = Empty{} 
		    		uniqueBanks[key] = &DBToCWwiseSoundbank{
                        "",
                        bank,
                        make(map[string]Empty),
                    }
                    uniqueBanks[key].linkedGameArchiveIds[gameArchiveID] = Empty{} 
		    	} else {
		    		if _, in := existBank.linkedGameArchiveIds[gameArchiveID]; !in {
		    			existBank.linkedGameArchiveIds[gameArchiveID] = Empty{} 
		    		}
		    	}
		    }
        default:
            for activeWorkers < 4 && dispatchTasks < len(gameArchives) {
                go extractSoundbankTask(
                    gameArchives[dispatchTasks].GameArchiveID,
                    rawData,
                    bankChannel,
                )
                activeWorkers++
                dispatchTasks++
            }
        }
    }
    if activeWorkers != 0 {
        return nil, errors.New("Leaked go coroutine when Wwise Soundbank " + 
        "extraction completed")
    }
    DefaultLogger.Info("Wwise Soundbank extraction complete",
		"totalArchiveVisits", finishedTasks,
        "totalBankVisits", totalBankVisits,
		"totalUniqueBanks", len(uniqueBanks),
    )

    DefaultLogger.Error("Errored game archives")
    for gameArchiveID, err := range erroredGameArchives {
        DefaultLogger.Error("Game archive with extraction error",
            "gameArchiveID", gameArchiveID,
            "error", err,
        )
    }

    close(bankChannel)
    // End

    // Free map memory
    uniqueBanksSlice := make([]*DBToCWwiseSoundbank, 0, len(uniqueBanks))
    for _, b := range uniqueBanks {
        uniqueBanksSlice = append(uniqueBanksSlice, b)
    }

	return uniqueBanksSlice, nil
}

func writeAllHierarchyObjs(
    hierarchy *HierarchyResult, db *sql.DB, ctx context.Context,
) error {
    dbQueries := database.New(db)

    objTypes, err := dbQueries.GetAllHierarchyObjectTypes(ctx)
    if err != nil {
        return err
    }

    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    queriesWithTx := dbQueries.WithTx(tx)
    if err = queriesWithTx.DeleteAllRandomSeqContainer(ctx); err != nil {
        return err
    }
    if err = queriesWithTx.DeleteAllSound(ctx); err != nil {
        return err
    }
    if err = queriesWithTx.DeleteAllHierarchyObject(ctx); err != nil {
        return err
    }

    totalObjRow := 0
    totalSoundRow := 0
    totalRandomSeqCntrRow := 0
    for ulid, obj := range hierarchy.uniqueObjs {
        linkedSoundbankPathNames := make([]string, 0, len(obj.linkedSoundbankPathNames))
        for dbid := range obj.linkedSoundbankPathNames {
            linkedSoundbankPathNames = append(linkedSoundbankPathNames, dbid)
        }

        dbid, err := genDBID()
        if err != nil {
            return errors.Join(
                errors.New("Failed to generate UUID4 for a given hierarchy object"),
                err,
            )
        }

        typeID, err := getHierarchyObjectTypePrimaryKey(
            objTypes, obj.ref.getType(),
        )
        if err != nil {
            return err
        }

        parentULID := int(obj.ref.getDirectParentID())

        if err = queriesWithTx.CreateHierarchyObject(
            ctx,
            database.CreateHierarchyObjectParams{
                DbID: dbid,
                WwiseObjectID: strconv.Itoa(int(ulid)),
                TypeDbID: typeID,
                ParentWwiseObjectID: strconv.Itoa(parentULID),
                LinkedSoundbankPathNames: strings.Join(linkedSoundbankPathNames, ";"),
            },
        ); err != nil {
            return err
        }
        totalObjRow++

        obj.dbid = dbid

        switch obj.ref.getType() {
        case HIERARCHY_TYPE_NAME[0]: // Sound
            sound, in := hierarchy.uniqueSounds[ulid]
            if !in { // invariant check
                errMsg := "A sound object is not in the hierarchy" + 
                "object structure."
                errMsg += fmt.Sprintf(" ULID: %d", ulid)
                return errors.New(errMsg)
            }
            for shortID := range sound.ShortIDs {
                if err = queriesWithTx.CreateSound(
                    ctx,
                    database.CreateSoundParams{
                        DbID: obj.dbid,
                        WwiseShortID: strconv.Itoa(int(shortID)),
                        Label: "",
                        Tags: "",
                    },
                ); err != nil {
                    return err
                }
            }
            totalSoundRow++
        case HIERARCHY_TYPE_NAME[1]: // Random / Sequence Container
            _, in := hierarchy.uniqueCntrs[ulid]
            if !in { // invariant check
                errMsg := "A random / sequence container object is not in " +
                "the hierarchy object struct"
                errMsg = fmt.Sprintf(" ULID: %d", ulid)
                return errors.New(errMsg)
            }
            if err = queriesWithTx.CreateRandomSeqContainer(
                ctx,
                database.CreateRandomSeqContainerParams{
                    DbID: obj.dbid,
                    Label: "",
                    Tags: "",
                },
            ); err != nil {
                return err
            }
            totalRandomSeqCntrRow++
        }
    }

    if err = tx.Commit(); err != nil {
        return err
    }

    DefaultLogger.Info("Finished rewrite all Hierarchy objects in all Wwise " +
    "Soundbanks Hierarchy objects", 
    "totalObjRow", totalObjRow,
    "totalSoundRow", totalSoundRow,
    "totalRandomSeqCntrRow", totalRandomSeqCntrRow,
    )

    return nil
}

/*
Main thread of exporting and parsing Wwiser XML. The main thread is 
responsible for checking invariant of parsing result, organizing unique and 
duplicated hierarchy objects. Each go routine is responsible for export and 
parsing Wwiser XML of a single soundbank. If an invarinat check is failed, 
immedately panic.

[return]
*ExtractHierarchyResult - Nil when an error occurs.
error - trivial
*/
func extractAllHierarchyObjs(
    banks []*DBToCWwiseSoundbank, _ context.Context,
) (*HierarchyResult, error) {
    uniqueObjs := make(map[uint32]*DBCAkObject)
    uniqueSounds := make(map[uint32]*CAkSound)
    uniqueCntrs := make(map[uint32]*CAkRanSeqCntr)

    erroredBanks := make(map[string]error)
    erroredObj := make(map[uint32]error)

    // For sync. 
    // Todo: this should be encapsulated into worker pool. Add context 
    // cancellation for go coroutine if main thread encouter a fatal error.
    parsedXMLChannel := make(chan *WwiserXMLTaskResult)
    finishedTasks := 0
    dispatchTasks := 0
    activeWorkers := 0
    for finishedTasks < len(banks) {
        select {
        case payload := <- parsedXMLChannel:
            if payload == nil {
                panic("Assertion failure. Received a nil payload from exporting" +
                " and parsing Wwiser XML tasks")
            }

            finishedTasks++
            activeWorkers--

            if payload.err != nil {
                if _, in := erroredBanks[payload.pathName]; in {
                    errMsg := fmt.Sprintf(
                        "Revisited a soundbank. Soundbank: %s",
                        payload.pathName,
                    )
                    return nil, errors.New(errMsg)
                } else {
                    erroredBanks[payload.pathName] = payload.err
                }
                break
            }

            if payload.result == nil {
                panic("Assertion failure. Received a nil payload from exporting" +
                " and parsing Wwiser XML tasks")
            }

            if payload.result.Hierarchy == nil {
                panic("Assertion failure. Parsing the result of Wwise Soundbank" +
                " without HircChunk")
            }

            for ulid, obj := range payload.result.Hierarchy.CAkObjects {
                if dbObj, in := uniqueObjs[ulid]; in {
                    dbObj.linkedSoundbankPathNames[payload.pathName] = Empty{}

                    // Invariant check
                    if dbObj.ref.getDirectParentID() != obj.getDirectParentID() {
                        errMsg := fmt.Sprintf(
                            "Parent ID in map: %d. Parent ID in received: %d.",
                            dbObj.ref.getDirectParentID(),
                            obj.getDirectParentID(),
                        )
                        erroredObj[ulid] = errors.New(errMsg)
                    }

                } else {
                    uniqueObjs[ulid] = &DBCAkObject{
                        ref: obj,
                        linkedSoundbankPathNames: make(map[string]Empty),
                    }
                    uniqueObjs[ulid].linkedSoundbankPathNames[payload.pathName] = Empty{}
                }
            }

            newSounds := 0
            for ulid, sound := range payload.result.Hierarchy.Sounds {
                if existSound, in := uniqueSounds[ulid]; !in {
                    uniqueSounds[ulid] = sound
                    newSounds++
                } else {

                    // invariant check
                    if existSound.getDirectParentID() != sound.getDirectParentID() {
                        errMsg := fmt.Sprintf("Two sound with same object ULID" +
                        "mismatch parent ULID. ULID: %d", ulid)
                        return nil, errors.New(errMsg)
                    }

                    for shortID := range sound.ShortIDs {
                        if _, in := existSound.ShortIDs[shortID]; !in {
                            existSound.ShortIDs[shortID] = Empty{}
                        }
                    }
                }
            }

            newCntrs := 0
            for ulid, cntr := range payload.result.Hierarchy.RanSeqCntrs {
                if existCntr, in := uniqueCntrs[ulid]; !in {
                    uniqueCntrs[ulid] = cntr
                    newCntrs++
                } else {

                    // invariant check
                    if existCntr.getDirectParentID() != existCntr.getDirectParentID() {
                        errMsg := fmt.Sprintf(
                            "Two container with same object ULID mismatch " +
                            "parent ULID. ULID: %d.",
                            ulid, 
                        )
                        return nil, errors.New(errMsg)
                    }

                }
            }
		    DefaultLogger.Info("Parsing Result", 
                "pathName", payload.pathName,
		    	"mediaIndexCount", payload.result.MediaIndex.Count,
		    	"referencedSounds", len(payload.result.Hierarchy.ReferencedSounds),
		    	"soundObjectCount", len(payload.result.Hierarchy.Sounds),
		    	"ranSeqCntrsCount", len(payload.result.Hierarchy.RanSeqCntrs),
                "newSounds", newSounds,
                "newCntrs", newCntrs,
		    )
        default:
            for activeWorkers < 4 && dispatchTasks < len(banks) {
                go exportParseWwiserXMLTask(
                    banks[dispatchTasks].ref.PathName,
                    banks[dispatchTasks],
                    parsedXMLChannel)
                dispatchTasks++
                activeWorkers++
            }
        }
    }
    if activeWorkers != 0 {
        errMsg := "Leaked go coroutine when Wwiser XML export and parsing " + 
        "extraction completed"
        return nil, errors.New(errMsg)
    }
    DefaultLogger.Info("Wwise Soundbank extraction complete")
    close(parsedXMLChannel)

    DefaultLogger.Error("Errored Soundbanks", 
        "numErroredBanks", len(erroredBanks),
    )
    for pathName, err := range erroredBanks {
        DefaultLogger.Error(pathName, "error", err)
    }

    DefaultLogger.Error("Errored Objects",
        "numErroredObjects", len(erroredObj),
    )
    for ulid, err := range erroredObj {
        DefaultLogger.Error(strconv.Itoa(int(ulid)), "error", err)
    }

    result := &HierarchyResult{ uniqueObjs, uniqueSounds, uniqueCntrs }

    return result, nil
}

/*
 Read the data from all files supported by 
*/
func updateSoundLabelsFromFileNoStruct(
    pathArgs string, ctx context.Context,
) error {
    paths := strings.Split(pathArgs, ";")

    db, err := getDBConn()
    if err != nil {
        return err
    }
    defer db.Close()

    dbQueries := database.New(db)

    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }

    queryWithTx := dbQueries.WithTx(tx)

    for _, p := range paths {
        var data LabelFile

        buf, err := os.ReadFile(p)
        if err != nil {
            DefaultLogger.Warn("Error opening file", "file_path", p, "error", err)
            continue
        }

        err = json.Unmarshal(buf, &data)
        if err != nil {
            DefaultLogger.Warn("Error parsing file", "file_path", p, "error", err)
            continue
        }

        for label, sourceId := range data.Sources {
            err := queryWithTx.UpdateSoundLabelBySourceId(
                ctx,
                database.UpdateSoundLabelBySourceIdParams{
                    Label: label,
                    WwiseShortID: sourceId,
                },
            )
            if err != nil {
                tx.Rollback()
                return err
            }
        }
    }

    err = tx.Commit()
    if err != nil {
        tx.Rollback()
        return err
    }

    return nil
}

func updateSoundLabelsFromFolderNoStruct(
    pathArg string, ctx context.Context,
) error {
    jsonFiles, err := os.ReadDir(pathArg)
    if err != nil {
        return err
    }

    db, err := getDBConn()
    if err != nil {
        return err
    }
    defer db.Close()

    dbQueries := database.New(db)

    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }

    queryWithTx := dbQueries.WithTx(tx)

    for _, f := range jsonFiles {
        var data LabelFile

        if !strings.HasSuffix(f.Name(), ".json") {
            continue
        }

        filename := path.Join(pathArg, f.Name())

        buf, err := os.ReadFile(filename)
        if err != nil {
            DefaultLogger.Warn(
                "Error opening file", 
                "file_path", filename, 
                "error", err,
            )
            continue
        }

        err = json.Unmarshal(buf, &data)
        if err != nil {
            DefaultLogger.Warn(
                "Error parsing file", 
                "file_path", filename, 
                "error", err,
            )
            continue
        }

        for label, sourceId := range data.Sources {
            err := queryWithTx.UpdateSoundLabelBySourceId(
                ctx,
                database.UpdateSoundLabelBySourceIdParams{
                    Label: label,
                    WwiseShortID: sourceId,
                },
            )
            if err != nil {
                tx.Rollback()
                return err
            }
        }
    }

    err = tx.Commit()
    if err != nil {
        tx.Rollback()
        return err
    }

    return nil
}
