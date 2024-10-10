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

	"dekr0/fuzzy/internal/database"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	_ "github.com/mattn/go-sqlite3"
)

func updateHelldiverFourVO(file string, db *sql.DB, logger *slog.Logger) error {
	if db == nil {
		return errors.New("Database connection cannot be nil")
	}
	if logger == nil {
		return errors.New("Logger cannot be nil")
	}
	csvData, err := os.Open(file)
	if err != nil {
		return err
	}
	defer csvData.Close()

	reader := csv.NewReader(csvData)

	recordCount := 0
	lineCount := 0

	dbQueries := database.New(db)

	bgCtx := context.Background()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryWithTx := dbQueries.WithTx(tx)

	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if len(record) != 3 {
			logger.Warn("The current line does not have exactly 3 columns. Skipping the current line", 
				"line_number", lineCount, "record", record)
			continue
		}

		lineCount += 1
		id, err := uuid.NewUUID()
		if err != nil {
			return err
		}

		if len(record[0]) == 0 {
			logger.Warn("A transcription without an file id. Skipping.")
			continue
		}

		if len(record[1]) == 1 {
			logger.Warn("A empty transcription???", "line_count", lineCount)
		}

		params := &database.CreateHelldiverFourTranscriptionParams{
			ID: id.String(),
			FileID: record[0],
			Transcription: record[1],
		}
		err = queryWithTx.CreateHelldiverFourTranscription(bgCtx, *params)
		if err != nil {
			return err
		}

		ftsParams := &database.CreateHelldiverFourVOFTSEntryParams{
			ID: id.String(),
			FileID: record[0],
			Transcription: record[1],
		}
		err = queryWithTx.CreateHelldiverFourVOFTSEntry(bgCtx, *ftsParams)
		if err != nil {
			return err
		}
		recordCount += 1
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	logger.Info("Helldiver Four Voice lines update finished", 
		"line_parsed", lineCount,
		"record_write", recordCount)

	return nil
}

func updateHelldiverAudioArchives(dir string, db *sql.DB, logger *slog.Logger) error {
	if db == nil {
		return errors.New("Database connection cannot be nil")
	}
	if logger == nil {
		return errors.New("Logger cannot be nil")
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	dbQueries := database.New(db)

	bgCtx := context.Background()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queriesTx := dbQueries.WithTx(tx)

	totalRecordCount := 0
	perFileRecordCount := 0
	lineCount := 0
	param := &database.CreateHelldiverAudioArchiveParams{}
	for _, filename := range files {
		csvData, err := os.Open(path.Join(dir, filename.Name()))
		if err != nil {
			return err
		}
		category := strings.Split(path.Base(csvData.Name()), ".")[0] 
		lineCount = 0
		perFileRecordCount = 0
		reader := csv.NewReader(csvData)
		for {
			record, err := reader.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			lineCount += 1
			if len(record) <= 1 {
				logger.Warn("An invalid archive entry. Skipping")
				continue
			}
			assetPath, archiveIds := record[0], record[1:]
			for _, archiveId := range archiveIds {
				id, err := uuid.NewUUID()
				if err != nil {
					return err
				}
				if len(archiveId) == 0 {
					continue
				}
				param.ID = id.String()
				param.ArchiveID = archiveId 
				param.Basename = path.Base(assetPath)
				param.Path = assetPath
				param.Category = sql.NullString{ String: category, Valid: true }
				err = queriesTx.CreateHelldiverAudioArchive(bgCtx, *param)
				if err != nil {
					return err
				}
				totalRecordCount += 1
				perFileRecordCount += 1
			}
		}
		logger.Info("Finished import one data source", 
			"file", filename.Name(),
			"total_count", totalRecordCount,
			"local_count", perFileRecordCount,
			"line_count", lineCount,
		)
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	logger.Info("Finished import all data source", "total_count", totalRecordCount,)
	return nil
}

func main() {
	godotenv.Load()

	/** Group this to setup once get too large */
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level: slog.LevelDebug,
	})
	logger := getLogger()(handler)

	dbPath := os.Getenv("DB_PATH")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.Error("Error occur when opening SQLite database", 
			"error", err.Error())
	}
	defer db.Close()

	/**
	if err := updateHelldiverFourVO("./csv/helldiver_four_vo.csv", db, logger); 
	err != nil {
		logger.Error("Error occur when updating Helldiver Four VO into database", 
			"error", err.Error())
		db.Close()
		os.Exit(1)
	}
	*/

	if err := updateHelldiverAudioArchives("./csv/archives", db, logger); 
	err != nil {
		logger.Error("Error occur when update Helldiver audio archives",
			"error", err.Error())
		db.Close()
		os.Exit(1)
	}
}
