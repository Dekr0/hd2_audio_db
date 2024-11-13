package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"io"
	"log/slog"
	"os"
	"path"
	"strings"

	"dekr0/hd-audio-archive-db/internal/database"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	_ "github.com/mattn/go-sqlite3"
)

func updateHelldiverAudioArchives(dir string, logger *slog.Logger) error {
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
		Tags map[string]struct{}
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
					Tags: make(map[string]struct{}),
					UniqueCategories: make(map[string]struct{}),
				}

				gameArchive.UniqueCategories[category] = struct{}{}
				gameArchive.Tags[tag] = struct{}{}

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
		for tag := range gameArchive.Tags {
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

func main() {
	godotenv.Load()

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level: slog.LevelDebug,
	})
	logger := getLogger()(handler)

	if err := updateHelldiverAudioArchives("./csv/archives", logger); err != nil {
		logger.Error("Failed to update Helldivers 2 game archives ID", "error", err)
		os.Exit(1)
	}
	os.Exit(0)
}
