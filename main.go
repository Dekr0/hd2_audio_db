package main

import (
	"context"
	"flag"
	"os"

    "dekr0/hd2_audio_db/db"
    "dekr0/hd2_audio_db/log"
    "dekr0/hd2_audio_db/toc"
    "dekr0/hd2_audio_db/wwise"
)

func run() error {
	if len(os.Args) > 3 {
		log.DefaultLogger.Error("Only one option can run at the same time")
	}

	dataPathFlag := flag.String("data", "", 
	"Absolute path to Helldivers 2 data directory")

	archiveTableFlag := flag.Bool("table-archive", false, 
	"Completely db.Rewrite basic information of all game archives (based on spreadsheets) into the DB" + 
	" (Destructive). Used by DB maintainers")

    allArchiveTableFlag := flag.Bool("table-archive-all", false,
    "Completely db.Rewrite basic information of all game archives (not based on " +
    "spreadsheets, based on content in the Helldivers 2 data directory) into the" +
    " DB (Destructive). Used by DB maintainers")

    bankTableFlag := flag.Bool("table-bank", false,
    "Completely db.Rewrite basic information about Wwise Soundbank into the DB" +
    "(Destructive). Used by DB maintainers.")

	hierarchyObjTableFlag := flag.Bool("table-hierarchy-object", false,
	"Completely db.Rewrite information about Wwise Soundbank and its hierarchy " +
	" objects used in game into the DB (Destructive). Used by DB maintainers.")

	streamTableFlag := flag.Bool("table-stream", false, 
	"Completely db.Rewrite information about all Wwise streams used in game into" +
	" the DB (Destructive). Used by DB maintainers.")

	extractBankFlag := flag.String("extract-bank", "", 
	"Extract basic information of Wwise Soundbanks in (a) game archive(s), Wwise " +
	"Soundbank binary content, and its Wwiser XML output")

    labelDirFlag := flag.String("import-label-folder", "",
    "Import label and description for hierarchy objects. Accept a abs. / relative" +
    " path of a directory that contains JSON files.",
    )

	xmlFlag := flag.String("xmls", "", "")

    if *dataPathFlag != "" {
        toc.DATA = *dataPathFlag
    } else {
        toc.DATA = os.Getenv("HELLDIVER2_DATA")
    }

    db.CSV_DIR = os.Getenv("CSV_DIR")
    if db.CSV_DIR == "" {
        db.CSV_DIR = "./csv/archives/"
    }

	flag.Parse()

	ctx := context.Background()

	if *archiveTableFlag {
        if err := db.RewriteAllHierarchyObjectTypes(ctx); err != nil {
            return err
        }
		return db.RewriteAllGameArchivesSpreadsheet("./csv/archives", ctx)
	}

    if *allArchiveTableFlag {
        if err := db.RewriteAllHierarchyObjectTypes(ctx); err != nil {
            return err
        }
        return db.RewriteAllGameArchives(ctx)
    }
    
    if *bankTableFlag {
        return db.RewriteAllSoundBank(ctx)
    }

	if *hierarchyObjTableFlag {
		return db.RewriteAllSoundAssets(ctx)
	}

	if *streamTableFlag {
        return nil
	}

	if *extractBankFlag != "" {
		toc.ExtractBank(*extractBankFlag)
		return nil
	}

	if *xmlFlag != "" {
		return wwise.ParseWwiserXML(*xmlFlag)
	}

    if *labelDirFlag != "" {
        return db.UpdateSoundLabelsFromFolder(*labelDirFlag, ctx)
    }

	flag.Usage()

	return nil
}

func main() {
	if err := run(); err != nil {
		log.DefaultLogger.Error("Error", "error_detail", err)
		os.Exit(1)
	}
	os.Exit(0)
}
