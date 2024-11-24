package main

import (
	"errors"
	"flag"
	"os"
)

func Run() error {
	if len(os.Args) > 2 {
		logger.Error("Only one option can run at the same time")
	}

	uflag := flag.Bool("u", false, "")
	xflag := flag.String("x", "", "")
	eflag := flag.String("e", "", "")

	flag.Parse()

	if *uflag {
		logger.Debug("Option selected: Update Helldivers 2 Game Archives Table")
		return updateHelldiverGameArchives("./csv/archives")
	}

	if *eflag != "" {
		logger.Debug("Option selected: Exporting Wwise XML")
		genWwiserXML(eflag)

		return nil
	}

	if *xflag != "" {
		logger.Debug("Option selected: Wwiser XML parsing")
		return ParseWwiseXML(xflag)
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
