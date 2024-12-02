package main

import (
	"context"
	"flag"
	"os"
)

type Empty struct {}

func run() error {
	if len(os.Args) > 3 {
		DefaultLogger.Error("Only one option can run at the same time")
	}

	dataPathFlag := flag.String("data", "", 
	"Absolute path to Helldivers 2 data directory")

	archiveTableFlag := flag.Bool("table-archive", false, 
	"Completely rewrite basic information of all game archives into the DB" + 
	" (Destructive)")

	bankTableFlag := flag.Bool("table-bank", false,
	"Completely rewrite information about Wwise Soundbank and its hierarchy " +
	" objects used in game into the DB (Destructive)")

	streamTableFlag := flag.Bool("table-stream", false, 
	"Completely rewrite information about all Wwise streams used in game into" +
	" the DB (Destructive)")

	extractBankFlag := flag.String("extract-bank", "", 
	"Extract basic information of Wwise Soundbanks in (a) game archive(s), Wwise " +
	"Soundbank binary content, and its Wwiser XML output")

	xmlFlag := flag.String("xmls", "", "")

    if *dataPathFlag != "" {
        DATA = *dataPathFlag
    } else {
        DATA = os.Getenv("HELLDIVER2_DATA")
    }

    CSV_DIR = os.Getenv("CSV_DIR")
    if CSV_DIR == "" {
        CSV_DIR = "./csv/archives/"
    }

	flag.Parse()

	ctx := context.Background()

	if *archiveTableFlag {
        if err := updateHelldiverHirearchyObjectType(ctx); err != nil {
            return err
        }
		return updateHelldiverGameArchives("./csv/archives", ctx)
	}

	if *bankTableFlag {
		return updateHelldiverSoundAssets(ctx)
	}

	if *streamTableFlag {
		return updateHelldiverStreams(ctx)
	}

	if *extractBankFlag != "" {
		extractBank(*extractBankFlag)
		return nil
	}

	if *xmlFlag != "" {
		return parseWwiserXML(*xmlFlag)
	}

	flag.Usage()

	return nil
}

func main() {
	if err := run(); err != nil {
		DefaultLogger.Error("Error", "error_detail", err)
		os.Exit(1)
	}
	os.Exit(0)
}
