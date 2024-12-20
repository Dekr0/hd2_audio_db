package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"dekr0/hd-audio-archive-db/internal/database"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	_ "github.com/mattn/go-sqlite3"
)

func updateHelldiverFourVO(csvFile string, logger *slog.Logger) error {
	if logger == nil {
		return errors.New("Logger cannot be nil")
	}
	
	dbPath := os.Getenv("DB_PATH")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	csvFileHandle, err := os.Open(csvFile)
	if err != nil {
		return err
	}
	defer csvFileHandle.Close()

	reader := csv.NewReader(csvFileHandle)

	totalRecordCount := 0
	totalLineCount := 0

	dbQueries := database.New(db)

	ctx := context.Background()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryWithTx := dbQueries.WithTx(tx)

	if err = queryWithTx.DeleteAllHelldiverFourTranscription(ctx); err != nil {
		return err
	}
	if err = queryWithTx.DeleteAllHelldiverFourVOFTSEntry(ctx); err != nil {
		return err
	}

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
				"total_line_count", totalLineCount,
				"record", record,
			)
			continue
		}

		totalLineCount += 1

		helldiverFourVoiceLineID, err := uuid.NewUUID()
		if err != nil {
			return err
		}
		helldiverFourVoiceLineIDS := helldiverFourVoiceLineID.String()
		if helldiverFourVoiceLineIDS == "" {
			logger.Warn("Failed to generate a String representation of UUID for this entry. Skipping", 
				"total_line_count", totalLineCount,
				"total_record_count", totalRecordCount,
			)
			continue
		}

		if len(record[0]) == 0 {
			logger.Warn("A transcription without an file id. Skipping.")
			continue
		}

		if len(record[1]) == 1 {
			logger.Warn("A empty transcription???", 
				"total_line_count", totalLineCount,
				"total_record_count", totalRecordCount,
				"record", record,
			)
		}

		params := &database.CreateHelldiverFourTranscriptionParams{
			HelldiverFourVoiceLineID: helldiverFourVoiceLineIDS,
			FileID: record[0],
			Transcription: record[1],
		}
		err = queryWithTx.CreateHelldiverFourTranscription(ctx, *params)
		if err != nil {
			return err
		}

		ftsParams := &database.CreateHelldiverFourVOFTSEntryParams{
			HelldiverFourVoiceLineID: helldiverFourVoiceLineID.String(),
			FileID: record[0],
			Transcription: record[1],
		}
		err = queryWithTx.CreateHelldiverFourVOFTSEntry(ctx, *ftsParams)
		if err != nil {
			return err
		}
		totalRecordCount += 1
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	logger.Info("Helldiver Four Voice lines update finished", 
		"total_line_parsed", totalLineCount,
		"total_record_written", totalRecordCount)

	return nil
}

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

	queriesTx := dbQueries.WithTx(tx)

	if err = queriesTx.DeleteAllHelldiverArchiveName(ctx); err != nil {
		return err
	}
	if err = queriesTx.DeleteAllHelldiverAudioArchive(ctx); err != nil {
		return err
	}
	if err = queriesTx.DeleteAllHelldiverAudioSource(ctx); err != nil {
		return err
	}

	totalRecordCount := 0
	perFileRecordCount := 0
	totalLineCount := 0

	audioArchiveNameSet := make(map[string]string)

	param := &database.CreateHelldiverAudioArchiveParams{}
	for _, csvFile := range csvFiles {
		csvFileHandle, err := os.Open(path.Join(dir, csvFile.Name()))
		if err != nil {
			return err
		}

		audioArchiveCategory := strings.Split(path.Base(csvFileHandle.Name()), ".")[0] 

		totalLineCount = 0
		perFileRecordCount = 0

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
				logger.Warn("An entry with invalid format audio archive data. Skipping",
					"total_record_count", totalRecordCount,
					"total_record_count_per_file", perFileRecordCount,
					"total_line_count", totalLineCount,
				)
				continue
			}

			audioArchiveName, audioArchiveIds := record[0], record[1:]
			audioArchiveNameIdS, in := audioArchiveNameSet[audioArchiveName];
			if !in {
				audioArchiveNameId, err := uuid.NewUUID()
				if err != nil {
					return err
				}
				audioArchiveNameIdS = audioArchiveNameId.String()
				if audioArchiveNameIdS == "" {
					err := fmt.Sprintf("Failed to generate a String representation of UUID for audio archive name %s", audioArchiveName)
					return errors.New(err)
				}
				audioArchiveNameSet[audioArchiveName] = audioArchiveNameIdS
				
				if err = queriesTx.CreateHelldiverAudioArchiveName(
					ctx,
					database.CreateHelldiverAudioArchiveNameParams{
						AudioArchiveNameID: audioArchiveNameIdS,
						AudioArchiveName: audioArchiveName,
					},
				); err != nil {
					return err
				}
			}
			
			for _, audioArchiveId := range audioArchiveIds {
				if len(audioArchiveId) == 0 {
					continue
				}

				audioArchiveDbId, err := uuid.NewUUID()
				if err != nil {
					return err
				}
				audioArchiveDbIdS := audioArchiveDbId.String()
				if audioArchiveDbIdS == "" {
					logger.Warn("Failed to generate a String representation of UUID for this entry. Skipping", 
						"total_record_count", totalRecordCount,
						"total_record_count_per_file", perFileRecordCount,
						"total_line_count", totalLineCount,
					)
					continue
				}

				param.AudioArchiveDbID = audioArchiveDbIdS
				param.AudioArchiveID = audioArchiveId
				param.AudioArchiveNameID = audioArchiveNameIdS
				param.AudioArchiveCategory = audioArchiveCategory 
				err = queriesTx.CreateHelldiverAudioArchive(ctx, *param)
				if err != nil {
					return err
				}
				totalRecordCount += 1
				perFileRecordCount += 1
			}
		}
		logger.Info("Finished import one data source", 
			"file", csvFile.Name(),
			"total_record_count", totalRecordCount,
			"total_record_count_per_file", perFileRecordCount,
			"total_line_count", totalLineCount,
		)
		csvFileHandle.Close()
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	logger.Info("Finished import all data source", 
		"total_record_count", totalRecordCount,)
	return nil
}

func generateDependencyByArchiveId(queriedId string) error {
	dbPath := os.Getenv("DB_PATH")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	/** Validate audio archive ID */
	dbQueries := database.New(db)
	ctx := context.Background()
	count, err := dbQueries.HasAudioArchiveID(ctx, queriedId)
	if err != nil {
		return err
	}
	if count == 0 {
		errMsg := fmt.Sprintf("Archive ID %s does not exist", queriedId)
		return errors.New(errMsg)
	}
	if count != 1 {
		return errors.New("Count should be exactly 1. Something wrong with the logic in CLI.")
	}
	/** End of validate audio archive ID */

	/** Name mapping */
	allAudioArchiveNames, err := dbQueries.QueryAllAudioArchiveName(ctx)
	if err != nil {
		return err
	}
	audioArchiveNameMap := make(map[string]string)
	for _, e := range allAudioArchiveNames {
		id, name := e.AudioArchiveNameID, e.AudioArchiveName
		if _, in := audioArchiveNameMap[id]; !in {
			audioArchiveNameMap[id] = name
		} else {
			errMsg := fmt.Sprintf("Duplicating audio archive name id %s", id)
			return errors.New(errMsg)
		}
	}
	allAudioArchiveNames = nil
	/** End of name mapping */

	sharedErr := make(chan error)
	safeErr := make(chan error)

	/** Shared audio sources */
	go func() {
		conn, err := sql.Open("sqlite3", dbPath)
		defer func() {
			if conn != nil {
				conn.Close()
			}
			sharedErr <- err
		}()
		if err != nil {
			return
		}

		dbQueries := database.New(conn)
		sharedAudioSources, err := dbQueries.QueryAllSharedAudioSourceByAudioArchiveID(
			ctx, sql.NullString{ String: queriedId, Valid: true },
		)

		filename := fmt.Sprintf("%s_shared.json", queriedId)
		dataFile, err := os.OpenFile(
			filename, 
			os.O_RDWR | os.O_CREATE, 
			0644,
		)
		if err != nil {
			return
		}
		defer func() {
			dataFile.Close()
		}()

		type SharedAudioSource struct {
			AudioSourceID string `json:"audio_source_id"`
			LinkedAudioArchiveIds []string `json:"linked_audio_archive_ids"`
			LinkedAudioArchiveNames []string `json:"linked_audio_archive_names"`
		}

		type Data struct {
			SharedAudioSourceIDs []string `json:"shared_audio_source_ids"`
			SharedAudioSources []SharedAudioSource `json:"detail"`
		}

		data := Data{ 
			make([]string, len(sharedAudioSources)),
			make([]SharedAudioSource, len(sharedAudioSources)),
		}
		for i, e := range sharedAudioSources {
			data.SharedAudioSourceIDs[i] = e.AudioSourceID
			s := SharedAudioSource{
				AudioSourceID: e.AudioSourceID,
				LinkedAudioArchiveNames: []string{},
			}
			for _, n := range strings.Split(e.LinkedAudioArchiveNameIds, ",") {
				v, in := audioArchiveNameMap[n]
				if !in {
					errMsg := fmt.Sprintf("Failed to locate audio archive name with ID %s", 
						n)
					err = errors.New(errMsg)
					return
				}
				s.LinkedAudioArchiveNames = append(s.LinkedAudioArchiveNames, v)
			}
			s.LinkedAudioArchiveIds = strings.Split(e.LinkedAudioArchiveIds, ",")
			data.SharedAudioSources[i] = s
		}
		bytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return
		}
		if _, err = dataFile.Write(bytes); err != nil {
			return
		}
	}()
	/** End of shared audio sources */

	/** Safe (Non shared audio sources) */
	go func() {
		conn, err := sql.Open("sqlite3", dbPath)
		defer func() {
			if conn != nil {
				conn.Close()
			}
			safeErr <- err
		}()

		safeAudioSources, err := dbQueries.QueryAllSafeAudioSourceByAudioArchiveID(
			ctx, queriedId,
		)
		if err != nil {
			return
		}

		filename := fmt.Sprintf("%s_safe.csv", queriedId)
		dataFile, err := os.OpenFile(
			filename, 
			os.O_RDWR | os.O_CREATE, 
			0644,
		)

		if err != nil {
			return
		}

		writer := csv.NewWriter(dataFile)
		defer func() {
			writer.Flush()
			dataFile.Close()
		}()
		for _, e := range safeAudioSources {
			record := []string{ e.AudioSourceID }
			splits := strings.Split(e.LinkedAudioArchiveNameIds, ",")
			if len(splits) != 1 {
				errMsg := fmt.Sprintf("Something wrong with the result of database querying. Multiple linked audio archive name ids detected")
				err = errors.New(errMsg)
				return
			}
			v, in := audioArchiveNameMap[splits[0]]
			if !in {
				errMsg := fmt.Sprintf(
					"Failed to locate audio archive name with ID %s", 
					splits[0],
				)
				err = errors.New(errMsg)
				return
			}
			record = append(record, v)
			err = writer.Write(record)
			if err != nil {
				return
			}
		}
	}()
	/** End of safe audio sources */

	err = <- sharedErr
	if err != nil {
		err = errors.Join(err, <- safeErr)
		return err
	} else {
		err = <- safeErr
	}
	return err 
}

func overwriteCheck(filename string, logger *slog.Logger) error {
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
	ctx := context.Background()
	
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(f)
	audioArchiveId, isPrefix, err := reader.ReadLine()
	if isPrefix {
		return errors.New("Error when reading header: line is too long")
	}
	if err != nil {
		return errors.Join(errors.New("Error when reading header"), err)
	}
	if len(audioArchiveId) == 0 {
		return errors.Join(errors.New("Missing archive id in the first line of input file"), err)
	}

	count, err := dbQueries.HasAudioArchiveID(ctx, string(audioArchiveId))
	if err != nil {
		return err
	}
	if count == 0 {
		err := fmt.Sprintf("Archive ID %s does not exist", string(audioArchiveId))
		return errors.New(err)
	}
	if count != 1 {
		return errors.New("Count should be exactly 1. Something wrong with the logic in CLI.")
	}

	patchedAudioSourceIDs := []string{}
	for {
		audioSourceID, isPrefix, err := reader.ReadLine()
		if isPrefix {
			f.Close()
			return errors.New("Error when reading header: line is too long")
		}
		if err != nil {
			if err == io.EOF {
				f.Close()
				break
			}
			f.Close()
			return err
		}
		if len(string(audioSourceID)) == 0 {
			continue
		}
		patchedAudioSourceIDs = append(patchedAudioSourceIDs, 
			string(audioSourceID))
	}
	if len(patchedAudioSourceIDs) == 0 {
		logger.Info("There are no audio sources needed to patched. Skip checking",
			"total_patched_audio_source_ids", len(patchedAudioSourceIDs),
		)
		return nil
	}

	allAudioArchiveNames, err := dbQueries.QueryAllAudioArchiveName(ctx)
	if err != nil {
		return err
	}
	/** ID: NAME */
	audioArchiveNameMap := make(map[string]string)
	for _, e := range allAudioArchiveNames {
		id, name := e.AudioArchiveNameID, e.AudioArchiveName
		if _, in := audioArchiveNameMap[id]; !in {
			audioArchiveNameMap[id] = name
		} else {
			err := fmt.Sprintf("Duplicating audio archive name id %s", id)
			return errors.New(err)
		}
	}
	allAudioArchiveNames = nil

	audioSources, err := dbQueries.QuerySharedAudioSourceByAudioSourceID(
		ctx, patchedAudioSourceIDs,
	)
	if err != nil {
		return err
	}

	if len(audioSources) == 0 {
		logger.Info("CLI suggests the provided audio sources have zero effect with other audio archives.")
		logger.Warn("Please double check there is zero conflict on release.")
		return nil
	}
	patchedAudioSourceIDs = nil

	type SharedAudioSource struct {
		AudioSourceID string `json:"audio_source_id"`
		AffectingAudioArchiveIds []string `json:"affecting_audio_archive_ids"`
		AffectingAudioArchiveNames []string `json:"affecting_audio_archive_names"`
	}

	type Result struct {
		SharedAudioSources []SharedAudioSource `json:"shared_audio_sources"`
	}

	result := &Result{ []SharedAudioSource{} }
	for _, e := range audioSources {
		linkedAudioArchiveIds := strings.Split(e.LinkedAudioArchiveIds, ",")
		linkedAudioArchiveNameIds := strings.Split(e.LinkedAudioArchiveNameIds, ",")
		if len(linkedAudioArchiveIds) != len(linkedAudioArchiveNameIds) {
			err := fmt.Sprintf("Potential data integrity errors in database. Miss match between the number of affected audio archive names and the number of affected archive ids")
			return errors.New(err)
		}
		sharedAudioSource := SharedAudioSource{ 
			AudioSourceID: "", 
			AffectingAudioArchiveNames: make([]string, len(linkedAudioArchiveNameIds)),
		}
		for i, f := range linkedAudioArchiveNameIds {
			n, in := audioArchiveNameMap[f]
			if !in {
				err := fmt.Sprintf("Audio archive name ID %s an audio archive name", f)
				return errors.New(err)
			}
			sharedAudioSource.AffectingAudioArchiveNames[i] = n
		}
		sharedAudioSource.AudioSourceID = e.AudioSourceID
		sharedAudioSource.AffectingAudioArchiveIds = linkedAudioArchiveIds
		result.SharedAudioSources = append(
			result.SharedAudioSources, 
			sharedAudioSource,
		)
	}

	blob, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		return err
	}	
	audioSources = nil
	result = nil

	filename = fmt.Sprintf("%s_result_%d", audioArchiveId, time.Now().Unix())
	blobFile, err := os.OpenFile(filename, os.O_RDWR | os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer blobFile.Close()
	if _, err = blobFile.Write(blob); err != nil {
		return err
	}

	logger.Info("CLI suggests the provided audio sources have effect with other audio archives.")
	logger.Info("Please check the output file", "output_file", filename)

	return nil
}

func cleanup(logger *slog.Logger, start time.Time, rcode int) {
	diff := time.Now().Sub(start)
	logger.Info("CLI performance status", 
		"trivial run time stats in milliseconds", diff.Milliseconds())

	os.Exit(rcode)
}

func main() {
	godotenv.Load()

	start := time.Now()

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level: slog.LevelDebug,
	})
	logger := getLogger()(handler)

	/** Initialize CLI flag definition */
	initAudioArchive := flag.Bool(
		"init_audio_archive",
		false,
		"(enable / disable flag) Re-initialize the data for the audio archive database tables",
	)
	initVO := flag.Bool(
		"init_vo",
		false,
		"(enable / disable flag) Re-initialize the data for voice line tables (Placeholder)",
	)
	genDepAll := flag.Bool(
		"gen_dep_all",
		false,
		"(enable / disable flag) Generate all shared audio sources and potential safe audio sources for each audio archive id",
	)
	genDepByArchiveID := flag.String(
		"gen_dep_archive_id",
		"",
		"(archive_id: string) Generate shared audio sources and potential safe audio sources for a provided audio archive id",
	)
	overWriteCheckFile := flag.String(
		"overwrite_check",
		"",
		"(input_file_path: string) Perform audio source overwrite check",
	)

	flag.Parse()

	rcode := 0
	if *initAudioArchive || *initVO {
		if *initVO {
			if err := updateHelldiverFourVO(
				"./csv/helldiver_four_vo.csv", 
				logger,
			); 
			err != nil {
				logger.Error("Error occur when updating Helldiver Four VO into database", 
					"error", err.Error())
				rcode = 1
				cleanup(logger, start, rcode)
			}
		}

		if *initAudioArchive {
			if err := updateHelldiverAudioArchives("./csv/archives", logger); 
			err != nil {
				logger.Error("Error occur when update Helldiver audio archives",
					"error", err.Error())
				rcode = 1
				cleanup(logger, start, rcode)
			}
		}

		cleanup(logger, start, rcode)
	}

	if len(*overWriteCheckFile) > 0 {
		err := overwriteCheck(*overWriteCheckFile, logger)
		if err != nil {
			logger.Error("Error occur when performing overwrite check")
			rcode = 1
		}
		cleanup(logger, start, rcode)
	}

	if *genDepAll {
		cleanup(logger, start, rcode)
	}

	if len(*genDepByArchiveID) > 0 {
		err := generateDependencyByArchiveId(*genDepByArchiveID)
		if err != nil {
			logger.Error("Error occur when generate dependency for the provided archive ID", 
				"error", err)
			rcode = 1
		}
		cleanup(logger, start, rcode)
	}

	flag.CommandLine.Usage()
}
