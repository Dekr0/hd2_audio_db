package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"dekr0/hd2_audio_db/internal/database"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	_ "github.com/mattn/go-sqlite3"
)

func genDBID() (*string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	
	idS := id.String()
	if idS == "" {
		return nil, errors.New("Empty string when generate UUID4 string")
	}

	return &idS, nil
}

func updateHelldiverGameArchives(dir string, ctx context.Context) error {
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

				dbId, err := genDBID()
				if err != nil {
					logger.Error(
						"Failed to generate UUID4 for a new game archive",
						"GameArchiveID", gameArchiveID,
						"error", err,
					)
					continue
				}

				gameArchive := GameArchive{
					ID: *dbId,
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

func updateHelldiverSoundbanks(dataPath string, ctx context.Context) error {
	godotenv.Load()

	if logger == nil {
		return errors.New("Logger cannot be nil")
	}

	dbPath := os.Getenv("DB_PATH")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	dbQueries := database.New(db)

	gameArchives, err := dbQueries.GetAllGameArchives(ctx)
	if err != nil {
		return err
	}

	logger.Info("", "len", len(gameArchives))

	totalBanksVisit := 0

	uniqueBanks := make(map[uint64]*WwiseSoundbank)
	for _, gameArchive := range gameArchives {
		tocFilePath := path.Join(dataPath, gameArchive.GameArchiveID)

		gameArchiveID := gameArchive.GameArchiveID

		_, err := os.Stat(tocFilePath); 
		if err != nil {
			if os.IsNotExist(err) {
				logger.Error("Game archive ID " + tocFilePath + " does not exist") 
				continue
			} else {
				logger.Error("OS error", "error", err)
				continue
			}
		}

		tocFile, err := os.Open(tocFilePath)
		if err != nil {
			logger.Error("File open error", "error", err)
			continue
		}
		defer tocFile.Close()

		banks, err := ExtractWwiseSoundbank(tocFile, false)
		if err != nil {
			logger.Error("Failed to parse ToC File", "error", err)
		}
		totalBanksVisit += len(banks)

		for id, bank := range banks {
			existBank, in := uniqueBanks[id]
			if !in {
				uniqueBanks[id] = bank
			} else {
				if _, in := existBank. LinkedGameArchiveIds[gameArchiveID]; !in {
					existBank.LinkedGameArchiveIds[gameArchiveID] = struct{}{}
				}
			}
		}
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queriesWithTx := dbQueries.WithTx(tx)
	if err = queriesWithTx.DeleteAllHelldiverSoundbank(ctx); err != nil {
		return err
	}

	totalRecord := 0
	for bid, b := range uniqueBanks {
		linkedGameArchiveIds := []string{}
		for gid := range b.LinkedGameArchiveIds {
			linkedGameArchiveIds = append(linkedGameArchiveIds, gid)
		}
		dbId, err := genDBID()
		if err != nil {
			logger.Error(
				"Failed to generate UUID4 for a new game archive",
				"ToCFileID", bid,
				"error", err,
			)
			continue
		}
		if err = queriesWithTx.CreateHelldiverSoundbank(ctx,
			database.CreateHelldiverSoundbankParams{
				ID: *dbId,
				TocFileID: strconv.Itoa(int(b.ToCFileId)),
				SonndbankPathName: b.PathName,
				SoundbankReadableName: "",
				Categories: "",
				LinkedGameArchiveIds: strings.Join(linkedGameArchiveIds, ";"),
			},
		); err != nil {
			return err
		}
		totalRecord++
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	
	logger.Info("Finished update Helldiver 2 Wwise Sounbank records",
		"totalRecord", totalRecord,
		"totalBanksVisit", totalBanksVisit,
		"totalUniqueBanks", len(uniqueBanks),
	)

	return nil
}

func updateHelldiverStreams(dataPath string, ctx context.Context) error {
	return nil
}
