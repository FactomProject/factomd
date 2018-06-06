package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/Utilities/CorrectChainHeads/correctChainHeads"
	"github.com/FactomProject/factomd/Utilities/tools"
	log "github.com/sirupsen/logrus"
)

var CheckFloating bool
var UsingAPI bool
var FixIt bool

const level string = "level"
const bolt string = "bolt"

func main() {
	var (
		useApi        = flag.Bool("api", false, "Use API instead")
		checkFloating = flag.Bool("floating", false, "Check Floating")
		fix           = flag.Bool("fix", false, "Actually fix head")
	)

	flag.Parse()
	UsingAPI = *useApi
	log.SetLevel(log.InfoLevel)

	fmt.Println("Usage:")
	fmt.Println("CorrectChainHeads level/bolt/api DBFileLocation")
	fmt.Println("Program will fix chainheads")

	if len(flag.Args()) < 2 {
		fmt.Println("\nNot enough arguments passed")
		os.Exit(1)
	}
	if len(flag.Args()) > 2 {
		fmt.Println("\nToo many arguments passed")
		os.Exit(1)
	}

	if flag.Args()[0] == "api" {
		UsingAPI = true
	}

	var reader tools.Fetcher

	if UsingAPI {
		reader = tools.NewAPIReader(flag.Args()[1])
	} else {
		levelBolt := flag.Args()[0]

		if levelBolt != level && levelBolt != bolt {
			fmt.Println("\nFirst argument should be `level` or `bolt`")
			os.Exit(1)
		}
		path := flag.Args()[1]
		reader = tools.NewDBReader(levelBolt, path)
	}

	// dblock, err := reader.FetchDBlockHead()

	correctChainHeads.FindHeads(reader, correctChainHeads.CorrectChainHeadConfig{
		CheckFloating: *checkFloating,
		Fix:           *fix,
	})
}
