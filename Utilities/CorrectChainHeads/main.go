package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
)

var UsingAPI bool

const level string = "level"
const bolt string = "bolt"

func main() {
	var (
		useApi = flag.Bool("api", false, "Use API instead")
	)

	flag.Parse()
	UsingAPI = *useApi

	fmt.Println("Usage:")
	fmt.Println("CorrectChainHeads level/bolt DBFileLocation")
	fmt.Println("Program will fix chainheads")

	if len(flag.Args()) < 2 {
		fmt.Println("\nNot enough arguments passed")
		os.Exit(1)
	}
	if len(flag.Args()) > 2 {
		fmt.Println("\nToo many arguments passed")
		os.Exit(1)
	}

	var reader Fetcher

	if UsingAPI {

	} else {
		levelBolt := os.Args[1]

		if levelBolt != level && levelBolt != bolt {
			fmt.Println("\nFirst argument should be `level` or `bolt`")
			os.Exit(1)
		}
		path := os.Args[2]
		reader = NewDBReader(levelBolt, path)
	}

	// dblock, err := reader.FetchDBlockHead()

	FindHeads(reader)
}

func FindHeads(f Fetcher) {
	chainHeads := make(map[string]interfaces.IHash)

	var err error
	var dblock interfaces.IDirectoryBlock

	head, err := f.FetchDBlockHead()
	if err != nil {
		panic(fmt.Sprintf("Error fetching head"))
	}

	height := head.GetDatabaseHeight()
	dblock = head
	top := height
	fmt.Println("Starting at", height)

	for ; height > 0; height-- {
		dblock, err = f.FetchDBlockByHeight(height)
		if err != nil {
			fmt.Println("Error fetching height %d", height)
		}

		eblockEnts := dblock.GetEBlockDBEntries()
		for _, eblk := range eblockEnts {
			if _, ok := chainHeads[eblk.GetChainID().String()]; ok {
				// Chainhead already exists
				continue
			}
			chainHeads[eblk.GetChainID().String()] = eblk.GetKeyMR()
			ch, err := f.FetchHeadIndexByChainID(eblk.GetChainID())
			if err != nil {
				fmt.Printf("Error getting chainhead for %s\n", eblk.GetChainID().String())
			} else {
				if !ch.IsSameAs(eblk.GetKeyMR()) {
					fmt.Println("ERROR: Chainhead found: %s, Expected %s :: For Chain: %s at height %d",
						ch.String(), eblk.GetKeyMR().String(), eblk.GetChainID().String(), height)
				}
			}
		}
		if height%1000 == 0 {
			fmt.Printf("Currently on %d our of %d. %d ChainHeads so far\n", height, top, len(chainHeads))
		}

		var _ = dblock
	}

	fmt.Printf("%d Chains found", len(chainHeads))

}

type Fetcher interface {
	FetchDBlockHead() (interfaces.IDirectoryBlock, error)
	FetchDBlockByHeight(dBlockHeight uint32) (interfaces.IDirectoryBlock, error)
	FetchDBlock(hash interfaces.IHash) (interfaces.IDirectoryBlock, error)
	FetchHeadIndexByChainID(chainID interfaces.IHash) (interfaces.IHash, error)
}

func NewDBReader(levelBolt string, path string) *databaseOverlay.Overlay {
	var dbase *hybridDB.HybridDB
	var err error
	if levelBolt == bolt {
		dbase = hybridDB.NewBoltMapHybridDB(nil, path)
	} else {
		dbase, err = hybridDB.NewLevelMapHybridDB(path, false)
		if err != nil {
			panic(err)
		}
	}

	dbo := databaseOverlay.NewOverlay(dbase)

	return dbo
}
