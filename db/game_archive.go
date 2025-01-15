package db

import (
	"context"
	"errors"
	"path"

	"os"
	"strings"

	"dekr0/hd2_audio_db/internal/database"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

    "dekr0/hd2_audio_db/log"
    "dekr0/hd2_audio_db/toc"
)

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
func RewriteAllGameArchivesSpreadsheet(dir string, ctx context.Context) error {
	if log.DefaultLogger == nil {
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
    if err = queriesWithTx.DeleteAllArchiveSoundbankRelation(ctx); err != nil {
        return err
    }
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
                    log.DefaultLogger.Error("Visit the same file more than once", 
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

            log.DefaultLogger.Info("Finished parsing one file",
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
	log.DefaultLogger.Info("Finished parsing all files",
        "totalLineParsed", totalFileParsed,
		"totalLineRead", totalLineRead,
		"totalFileParse", totalFileParsed,
	)

    log.DefaultLogger.Error("Error Report", "numErrorCSVFiles", len(errorCSVs))
    for filename, err := range errorCSVs {
        log.DefaultLogger.Error(filename, "error", err)
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

	log.DefaultLogger.Info("Finished rewriting all game archives ID",
		"totalRows", totalRows,
	)
    // end

	return nil
}

func RewriteAllGameArchives(ctx context.Context) error {
    if log.DefaultLogger == nil {
        return errors.New("Logger cannot be nil")
    }

    db, err := getDBConn()
    if err != nil {
        return err
    }
    defer db.Close()

    files, err := os.ReadDir(toc.DATA)
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
    if err = queriesWithTx.DeleteAllArchiveSoundbankRelation(ctx); err != nil {
        return err
    }
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

        gameArchivePath := path.Join(toc.DATA, splits[0])
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

    log.DefaultLogger.Info("Finished rewriting all game archives ID")
    log.DefaultLogger.Info("Info status", "totalRows", totalRows)
    log.DefaultLogger.Error("Error status")
    for _, f := range errorGameArchive {
        log.DefaultLogger.Error("Malformed filename", "filename", f)
    }

    return nil
}
