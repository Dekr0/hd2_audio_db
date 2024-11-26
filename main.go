package main

import (
	"context"
	"errors"
	"flag"
	"os"
)

func Run() error {
	if len(os.Args) > 3 {
		logger.Error("Only one option can run at the same time")
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

	flag.Parse()

	ctx := context.Background()

	if *archiveTableFlag {
		return updateHelldiverGameArchives("./csv/archives", ctx)
	}

	if *bankTableFlag {
		return updateHelldiverSoundbanks(*dataPathFlag, ctx)
	}

	if *streamTableFlag {
		return updateHelldiverStreams(*dataPathFlag, ctx)
	}

	if *extractBankFlag != "" {
		logger.Debug("Option selected: extract-bank")
		ExtractBank(extractBankFlag)
		return nil
	}

	if *xmlFlag != "" {
		logger.Debug("Option selected: Wwiser XML parsing")
		return ParseWwiseXML(xmlFlag)
	}

	flag.Usage()

	return errors.New("Invalid argument for the provided option")
}

func main() {
	if err := Run(); err != nil {
		logger.Error("Error", "error_detail", err)
		os.Exit(1)
	}
	os.Exit(0)
}
