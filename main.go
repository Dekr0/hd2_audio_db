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
	"Completely rewrite basic information of all game archives (based on spreadsheets) into the DB" + 
	" (Destructive). Used by DB maintainers")

    allArchiveTableFlag := flag.Bool("table-archive-all", false,
    "Completely rewrite basic information of all game archives (not based on " +
    "spreadsheets, based on content in the Helldivers 2 data directory) into the" +
    " DB (Destructive). Used by DB maintainers")

    bankTableFlag := flag.Bool("table-bank", false,
    "Completely rewrite basic information about Wwise Soundbank into the DB" +
    "(Destructive). Used by DB maintainers.")

	soundAssetsTableFlag := flag.Bool("table-sound-asset", false,
	"Completely rewrite information about Wwise Soundbank and its hierarchy " +
	" objects used in game into the DB (Destructive). Used by DB maintainers.")

	streamTableFlag := flag.Bool("table-stream", false, 
	"Completely rewrite information about all Wwise streams used in game into" +
	" the DB (Destructive). Used by DB maintainers.")

	extractBankFlag := flag.String("extract-bank", "", 
	"Extract basic information of Wwise Soundbanks in (a) game archive(s), Wwise " +
	"Soundbank binary content, and its Wwiser XML output")

    labelNoStructFlag := flag.String("table-label", "",
    "Import labels for audio source into the database (Overwrite the existing" +
    "one). Accept a list of `.json` file path separated by `;`")

    labelNoStructFolderFlag := flag.String("table-label-folder", "",
    "Import labels for audio source into the database (Overwrite the existing" +
    "one). Accept a folder path that contains a collection of `.json` file")

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
        if err := rewriteAllHierarchyObjectTypes(ctx); err != nil {
            return err
        }
		return rewriteAllGameArchivesSpreadsheet("./csv/archives", ctx)
	}

    if *allArchiveTableFlag {
        if err := rewriteAllHierarchyObjectTypes(ctx); err != nil {
            return err
        }
        return rewriteAllGameArchives(ctx)
    }
    
    if *bankTableFlag {
        return rewriteAllSoundBank(ctx)
    }

	if *soundAssetsTableFlag {
		return rewriteAllSoundAssets(ctx)
	}

	if *streamTableFlag {
        return nil
	}

	if *extractBankFlag != "" {
		extractBank(*extractBankFlag)
		return nil
	}

	if *xmlFlag != "" {
		return parseWwiserXML(*xmlFlag)
	}

    if *labelNoStructFlag != "" {
        return updateSoundLabelsFromFileNoStruct(*labelNoStructFlag, ctx)
    }

    if *labelNoStructFolderFlag != "" {
        return updateSoundLabelsFromFolderNoStruct(*labelNoStructFolderFlag, ctx)
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
