package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"flag"
	"io"
	"os"
	"path"
	"strings"

	"dekr0/hd-audio-archive-db/internal/database"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	_ "github.com/mattn/go-sqlite3"
)

func updateHelldiverGameArchives(dir string) error {
	godotenv.Load()

	if logger == nil {
		return errors.New("Logger cannot be nil")
	}

	csvFiles, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	dbPath := os.Getenv("DB_PATH")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	dbQueries := database.New(db)
	ctx := context.Background()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queriesWithTx := dbQueries.WithTx(tx)
	if err = queriesWithTx.DeleteAllHelldiverGameArchive(ctx); err != nil {
		return err
	}

	totalLineCount := 0
	totalFileParse := 0
	perFileLineCount := 0

	type GameArchive struct {
		ID string
		GameArchiveID string
		UniqueTags map[string]struct{}
		UniqueCategories map[string]struct{}
	}

	uniqueGameArchiveIds := make(map[string]GameArchive)

	for _, csvFile := range csvFiles {
		filename := csvFile.Name()

		splits := strings.Split(path.Base(filename), ".")
		if len(splits) <= 1 {
			logger.Warn("Invalid input file: missing file extension",
				"file", filename,
				"total_line_count", totalLineCount,
				"total_file_process", totalFileParse,
			)
			continue
		}
		if splits[1] != "csv" {
			logger.Warn("Invalid input file: not a csv file",
				"file", filename,
				"total_line_count", totalLineCount,
				"total_file_parse", totalFileParse,
			)
			continue
		}

		category := splits[0]

		csvFileHandle, err := os.Open(path.Join(dir, filename))
		if err != nil {
			return err
		}
		defer csvFileHandle.Close()

		perFileLineCount = 0

		reader := csv.NewReader(csvFileHandle)
		for {
			record, err := reader.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			totalLineCount += 1
			if len(record) <= 1 {
				logger.Warn("An entry with invalid format. Skipping",
					"entry", record,
					"file", filename,
					"total_line_count", totalLineCount,
					"total_file_parse", totalFileParse,
				)
				continue
			}
			perFileLineCount += 1

			tag, gameArchiveIDs := record[0], record[1:]
			for _, gameArchiveID := range gameArchiveIDs {
				if gameArchiveID == "" {
					continue
				}

				if gameArchive, in := uniqueGameArchiveIds[gameArchiveID]; in {
					if _, in = gameArchive.UniqueCategories[category]; !in {
						gameArchive.UniqueCategories[category] = struct{}{}
					}
					if _, in = gameArchive.UniqueTags[tag]; !in {
						gameArchive.UniqueTags[tag] = struct{}{}
					}
					continue
				}

				gameArchiveDBId, err := uuid.NewUUID()
				if err != nil {
					errMsg := "Failed to generate UUID4 for a new game archive " +
				             gameArchiveID
					return errors.Join(err, errors.New(errMsg))
				}

				gameArchiveDBIdS := gameArchiveDBId.String()
				if gameArchiveDBIdS == "" {
					errMsg := "Empty string when generate UUID4 string for a new " + 
					         "archive " + gameArchiveID
					return errors.New(errMsg)
				}

				gameArchive := GameArchive{
					ID: gameArchiveDBIdS,
					GameArchiveID: gameArchiveID,
					UniqueTags: make(map[string]struct{}),
					UniqueCategories: make(map[string]struct{}),
				}

				gameArchive.UniqueCategories[category] = struct{}{}
				gameArchive.UniqueTags[tag] = struct{}{}

				uniqueGameArchiveIds[gameArchiveID] = gameArchive
			}
		}
		csvFileHandle.Close()

		totalFileParse += 1

		logger.Info("Finished parsing one file",
			"file", filename,
			"total_line_count", totalLineCount,
			"per_file_line_count", perFileLineCount,
			"totalFileParse", totalFileParse,
		)
	}

	logger.Info("Finished parsing all files", 
		"total_line_count", totalLineCount,
		"totalFileParse", totalFileParse,
	)

	totalRecord := 0

	for _, gameArchive := range uniqueGameArchiveIds {
		categories := []string{} 
		for category := range gameArchive.UniqueCategories {
			categories = append(categories, category)
		}
		tags := []string{}
		for tag := range gameArchive.UniqueTags {
			tags = append(tags, tag)
		}
		if err = queriesWithTx.CreateHelldiverGameArchive(
			ctx,
			database.CreateHelldiverGameArchiveParams{
				ID: gameArchive.ID,
				GameArchiveID: gameArchive.GameArchiveID,
				Tags: strings.Join(tags, ";"),
				Categories: strings.Join(categories, ";"),
			},
		); err != nil {
			return err
		}
		totalRecord += 1
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	logger.Info("Finished update Helldivers 2 game archives ID",
		"total_record", totalRecord,
	)

	return nil
}

func Run() error {
	xmlArg := flag.String("xmls", "", "A list of xmls describe Wwise Soundbank")
	updateArchiveArg := flag.Bool("update-audio-archives", false, 
	"Generate a SQL table of Helldivers 2 game archive")

	flag.Parse()

	if xmlArg != nil && *updateArchiveArg {
		return errors.New("Game Archive update and XML analysis cannot be " +
		"run at the same time")
	}
	
	if xmlArg != nil {
		return WwiserOuputParsing(xmlArg)
	}

	return updateHelldiverGameArchives("./csv/archives")
}

func main() {
	if err := Run(); err != nil {
		logger.Error("Error", "error_detail", err)
		os.Exit(1)
	}
	os.Exit(0)
}
