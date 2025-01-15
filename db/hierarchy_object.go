package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"strconv"

	"dekr0/hd2_audio_db/internal/database"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	"dekr0/hd2_audio_db/log"
	"dekr0/hd2_audio_db/wwise"
)

/*
Rewriting all database records about all possible hierarchy object type in 
Wwise soundbank. It will completely erase all records previously written.
*/
func RewriteAllHierarchyObjectTypes(ctx context.Context) error {
	if log.DefaultLogger == nil {
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
    for _, name := range wwise.HIERARCHY_TYPE_NAME {
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

    log.DefaultLogger.Info("Update Hierarchy Object Type Stat",
        "numHierarchyType", len(wwise.HIERARCHY_TYPE_NAME),
        "totalRows", totalRows,
    )

    return nil
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
    if err = queriesWithTx.DeletionAllHierarchyObjectRelation(ctx); err != nil {
        return err
    }
    if err = queriesWithTx.DeleteAllHierarchyObject(ctx); err != nil {
        return err
    }

    totalObjRow := 0
    totalObjSoundbankRow := 0
    totalSoundRow := 0
    totalRandomSeqCntrRow := 0
    for ulid, obj := range hierarchy.uniqueObjs {
        dbid, err := genDBID()
        if err != nil {
            return errors.Join(
                errors.New("Failed to generate UUID4 for a given hierarchy object"),
                err,
            )
        }

        typeID, err := getHierarchyObjectTypePrimaryKey(
            objTypes, obj.ref.GetType(),
        )
        if err != nil {
            return err
        }

        parentULID := int(obj.ref.GetDirectParentID())

        if err = queriesWithTx.CreateHierarchyObject(
            ctx,
            database.CreateHierarchyObjectParams{
                DbID: dbid,
                WwiseObjectID: strconv.Itoa(int(ulid)),
                TypeDbID: typeID,
                ParentWwiseObjectID: strconv.Itoa(parentULID),
            },
        ); err != nil {
            return err
        }
        totalObjRow++

        obj.dbid = dbid

        for soundbankDbID := range obj.soundbankDbIDs {
            if err = queriesWithTx.CreateSoundbankHierarchyObjectRelation(
                ctx,
                database.CreateSoundbankHierarchyObjectRelationParams{
                    SoundbankDbID: soundbankDbID,
                    HierarchyObjectDbID: obj.dbid,
                },
            ); err != nil {
                return err
            }
        }
        totalObjSoundbankRow++

        switch obj.ref.GetType() {
        case wwise.HIERARCHY_TYPE_NAME[0]: // Sound
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
                        Description: "",
                    },
                ); err != nil {
                    return err
                }
            }
            totalSoundRow++
        case wwise.HIERARCHY_TYPE_NAME[1]: // Random / Sequence Container
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
                    Description: "",
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

    log.DefaultLogger.Info("Finished Rewrite all Hierarchy objects in all Wwise " +
    "Soundbanks Hierarchy objects", 
    "totalObjRow", totalObjRow,
    "totalObjSoundbankRow", totalObjSoundbankRow,
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
    uniqueSounds := make(map[uint32]*wwise.CAkSound)
    uniqueCntrs := make(map[uint32]*wwise.CAkRanSeqCntr)

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
                    dbObj.soundbankDbIDs[payload.soundbankDbID] = Empty{}

                    // Invariant check
                    if dbObj.ref.GetDirectParentID() != obj.GetDirectParentID() {
                        errMsg := fmt.Sprintf(
                            "Parent ID in map: %d. Parent ID in received: %d.",
                            dbObj.ref.GetDirectParentID(),
                            obj.GetDirectParentID(),
                        )
                        erroredObj[ulid] = errors.New(errMsg)
                    }

                } else {
                    uniqueObjs[ulid] = &DBCAkObject{
                        ref: obj,
                        soundbankDbIDs: make(map[string]Empty),
                    }
                    uniqueObjs[ulid].soundbankDbIDs[payload.soundbankDbID] = Empty{}
                }
            }

            newSounds := 0
            for ulid, sound := range payload.result.Hierarchy.Sounds {
                if existSound, in := uniqueSounds[ulid]; !in {
                    uniqueSounds[ulid] = sound
                    newSounds++
                } else {

                    // invariant check
                    if existSound.GetDirectParentID() != sound.GetDirectParentID() {
                        errMsg := fmt.Sprintf("Two sound with same object ULID" +
                        "mismatch parent ULID. ULID: %d", ulid)
                        return nil, errors.New(errMsg)
                    }

                    for shortID := range sound.ShortIDs {
                        if _, in := existSound.ShortIDs[shortID]; !in {
                            existSound.ShortIDs[shortID] = wwise.Empty{}
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
                    if existCntr.GetDirectParentID() != existCntr.GetDirectParentID() {
                        errMsg := fmt.Sprintf(
                            "Two container with same object ULID mismatch " +
                            "parent ULID. ULID: %d.",
                            ulid, 
                        )
                        return nil, errors.New(errMsg)
                    }

                }
            }
		    log.DefaultLogger.Info("Parsing Result", 
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
    log.DefaultLogger.Info("Wwise Soundbank extraction complete")
    close(parsedXMLChannel)

    log.DefaultLogger.Error("Errored Soundbanks", 
        "numErroredBanks", len(erroredBanks),
    )
    for pathName, err := range erroredBanks {
        log.DefaultLogger.Error(pathName, "error", err)
    }

    log.DefaultLogger.Error("Errored Objects",
        "numErroredObjects", len(erroredObj),
    )
    for ulid, err := range erroredObj {
        log.DefaultLogger.Error(strconv.Itoa(int(ulid)), "error", err)
    }

    result := &HierarchyResult{ uniqueObjs, uniqueSounds, uniqueCntrs }

    return result, nil
}

