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
)

func RewriteAllSoundBank(ctx context.Context) error {
	if log.DefaultLogger == nil {
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
func RewriteAllSoundAssets(ctx context.Context) error {
	if log.DefaultLogger == nil {
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
    if err = queriesWithTx.DeleteAllArchiveSoundbankRelation(ctx); err != nil {
        return err
    }
	if err = queriesWithTx.DeleteAllSoundbank(ctx); err != nil {
		return err
	}

	totalSoundbankRows := 0
    totalArchiveSoundbankRows := 0
	for _, b := range banks {
		dbId, err := genDBID()
		if err != nil {
			log.DefaultLogger.Error(
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
			},
		); err != nil {
			return err
		}
		totalSoundbankRows++

        b.dbID = dbId

        for gameArchiveDbID := range b.gameArchiveDbIDs {
            if err = queriesWithTx.CreateArchiveSounbankRelation(ctx, 
                database.CreateArchiveSounbankRelationParams {
                    GameArchiveDbID: gameArchiveDbID,
                    SoundbankDbID: b.dbID, 
                },
            ); err != nil {
                return err
            }
            totalArchiveSoundbankRows++
        }

	}
	if err = tx.Commit(); err != nil {
		return err
	}
    // End
	
	log.DefaultLogger.Info("Finished rewriting Wwise Sounbank records",
		"totalRows", totalSoundbankRows,
        "totalArchiveSoundbankRows", totalArchiveSoundbankRows,
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

            gameArchiveDBID := payload.gameArchiveDbID
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
		    	existBank, in := uniqueBanks[key]
		    	if !in {
		    		uniqueBanks[key] = &DBToCWwiseSoundbank{
                        "",
                        bank,
                        make(map[string]Empty),
                    }
                    uniqueBanks[key].gameArchiveDbIDs[gameArchiveDBID] = Empty{} 
		    	} else {
		    		if _, in := existBank.gameArchiveDbIDs[gameArchiveDBID]; !in {
		    			existBank.gameArchiveDbIDs[gameArchiveDBID] = Empty{} 
		    		}
		    	}
		    }
        default:
            for activeWorkers < 4 && dispatchTasks < len(gameArchives) {
                go extractSoundbankTask(
                    gameArchives[dispatchTasks].DbID,
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
    log.DefaultLogger.Info("Wwise Soundbank extraction complete",
		"totalArchiveVisits", finishedTasks,
        "totalBankVisits", totalBankVisits,
		"totalUniqueBanks", len(uniqueBanks),
    )

    log.DefaultLogger.Error("Errored game archives")
    for gameArchiveID, err := range erroredGameArchives {
        log.DefaultLogger.Error("Game archive with extraction error",
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
