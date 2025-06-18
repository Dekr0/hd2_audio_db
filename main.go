package main

import (
	"context"
	"dekr0/hd2_audio_db/db"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	insertArchive := flag.Bool(
		"insert_archive",
		false, 
		"Insert records of all archives in the `data` folder into `archive` " + 
		"table.",
	)
	generate := flag.Bool(
		"generate",
		false,
		"Populate records for `asset`, `soundbank`, `hierarchy`, and `sound`" +
		" table.",
	)
	extractAllSoundbank := flag.Bool(
		"extract_all_soundbank",
		false,
		"Extract sound banks from all archives in `data` folder",
	)
	extractSoundbank := flag.Bool(
		"extract_soundbank",
		false,
		"Extract sound banks from selective archives." + 
		"Multi-selective is enable. Use `Tab` to select / deselect",
	)
	extractSoundbankDB := flag.Bool(
		"extract_soundbank_db",
		false,
		"Extract sound banks from selective archives. Each archive is tag by " + 
		"a sound bank name. Multi-selective is enable. Use `Tab` to select / " +
		"deselect",
	)
	data := flag.String("data", "", "")
	dest := flag.String("dest", "", "")

	flag.Parse()

	if *data != "" {
		slog.Info("Using data path from argument.")
		stat, err := os.Lstat(*data)
		if err != nil {
			slog.Error("Failed to query provided data path", "error", err)
			fmt.Println(err)
			os.Exit(1)
		}
		if stat.IsDir() {
			if err := os.Setenv("DATA", *data); err != nil {
				slog.Error("Failed to use provided data path", "error", err)
			}
		} else {
			slog.Error("Provided data path is a file")
		}
	}

	*data = os.Getenv("DATA")
	stat, err := os.Lstat(*data)
	if err != nil {
		slog.Error("Failed to query provided data path", "error", err)
		os.Exit(1)
	}
	if !stat.IsDir() {
		slog.Error("Provided data path is a file")
		os.Exit(1)
	}

	if *insertArchive {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second * 12)
		defer cancel()
		if err := db.WriteArchives(ctx, *data); err != nil {
			slog.Error("Failed to insert records into archive table", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *generate {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second * 360)
		defer cancel()
		if err := db.Generate(ctx, *data); err != nil {
			slog.Error(
				"Populate records for `asset`, `soundbank`, `hierarchy`, and " + 
				"`sound` table.",
				"error", err,
			)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *extractAllSoundbank  {
		if *dest == "" {
			slog.Error("Destination for output sound bank is not provided")
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second * 16)
		defer cancel()
		if err := db.ExportAllSoundbank(ctx, *data, *dest); err != nil {
			slog.Error("Failed to extract all sound banks", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *extractSoundbank {
		if *dest == "" {
			slog.Error("Destination for output sound bank is not provided")
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second * 16)
		defer cancel()
		if err := db.ExportSoundbanksTUI(ctx, *data, *dest); err != nil {
			slog.Error("Failed to extract sound banks", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *extractSoundbankDB {
		if *dest == "" {
			slog.Error("Destination for output sound bank is not provided")
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second * 16)
		defer cancel()
		if err := db.ExportSoundbanksTUIDB(ctx, *data, *dest); err != nil {
			slog.Error("Failed to extract sound banks", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	flag.Usage()
}

