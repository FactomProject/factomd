package main

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
)

const level string = "level"
const bolt string = "bolt"

func main() {
	fmt.Println("Usage:")
	fmt.Println("DBCleanCopy level/bolt DBFileLocation")
	fmt.Println("Database will be copied over block by block to remove some DB inconsistencies")

	if len(os.Args) < 3 {
		fmt.Println("\nNot enough arguments passed")
		os.Exit(1)
	}
	if len(os.Args) > 3 {
		fmt.Println("\nToo many arguments passed")
		os.Exit(1)
	}

	levelBolt := os.Args[1]

	if levelBolt != level && levelBolt != bolt {
		fmt.Println("\nFirst argument should be `level` or `bolt`")
		os.Exit(1)
	}
	path := os.Args[2]

	var dbase1 *hybridDB.HybridDB
	var dbase2 *hybridDB.HybridDB

	var err error
	if levelBolt == bolt {
		dbase1 = hybridDB.NewBoltMapHybridDB(nil, path)
		dbase2 = hybridDB.NewBoltMapHybridDB(nil, "copied.db")
	} else {
		dbase1, err = hybridDB.NewLevelMapHybridDB(path, false)
		if err != nil {
			panic(err)
		}
		dbase2, err = hybridDB.NewLevelMapHybridDB("copied.db", true)
		if err != nil {
			panic(err)
		}
	}

	dbo1 := databaseOverlay.NewOverlay(dbase1)
	dbo2 := databaseOverlay.NewOverlay(dbase2)

	CopyDB(dbo1, dbo2)

	dbo1.Close()
	dbo2.Close()
}

func CopyDB(dbase1, dbase2 interfaces.DBOverlay) {
	processing := ""
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("Error processing: %v", processing)
			panic(err)
		}
	}()

	dBlocks, err := dbase1.FetchAllDBlocks()
	if err != nil {
		panic(err)
	}
	prevECHash := primitives.NewZeroHash()
	for _, dBlock := range dBlocks {
		dbase2.StartMultiBatch()

		err := dbase2.ProcessDBlockMultiBatch(dBlock)
		if err != nil {
			panic(err)
		}

		for _, dbEntry := range dBlock.GetDBEntries() {
			switch dbEntry.GetChainID().String() {
			case "000000000000000000000000000000000000000000000000000000000000000a":
				aBlock, err := dbase1.FetchABlock(dbEntry.GetKeyMR())
				if err != nil {
					panic(err)
				}
				err = dbase2.ProcessABlockMultiBatch(aBlock)
				if err != nil {
					panic(err)
				}
				break
			case "000000000000000000000000000000000000000000000000000000000000000c":
				ecBlock, err := dbase1.FetchECBlock(dbEntry.GetKeyMR())
				if err != nil {
					panic(err)
				}

				if ecBlock.GetHeader().GetPrevHeaderHash().IsSameAs(prevECHash) == false {
					prev, err := dbase1.FetchECBlock(ecBlock.GetHeader().GetPrevHeaderHash())
					if err != nil {
						panic(err)
					}
					err = dbase2.ProcessECBlockMultiBatch(prev, true)
					if err != nil {
						panic(err)
					}
				}

				err = dbase2.ProcessECBlockMultiBatch(ecBlock, true)
				if err != nil {
					panic(err)
				}
				prevECHash = dbEntry.GetKeyMR()
				break
			case "000000000000000000000000000000000000000000000000000000000000000f":
				fBlock, err := dbase1.FetchFBlock(dbEntry.GetKeyMR())
				if err != nil {
					panic(err)
				}
				err = dbase2.ProcessFBlockMultiBatch(fBlock)
				if err != nil {
					panic(err)
				}
				break
			default:
				processing = fmt.Sprintf("%v - %v - %v", dBlock.GetKeyMR().String(), dbEntry.GetChainID().String(), dbEntry.GetKeyMR().String())
				eBlock, err := dbase1.FetchEBlock(dbEntry.GetKeyMR())
				if err != nil {
					panic(err)
				}
				err = dbase2.ProcessEBlockMultiBatch(eBlock, true)
				if err != nil {
					panic(err)
				}
				for _, h := range eBlock.GetEntryHashes() {
					entry, err := dbase1.FetchEntry(h)
					if err != nil {
						panic(err)
					}
					err = dbase2.InsertEntryMultiBatch(entry)
					if err != nil {
						panic(err)
					}
				}
				break
			}
		}
		if err := dbase2.ExecuteMultiBatch(); err != nil {
			panic(err)
		}
		if dBlock.GetDatabaseHeight()%1000 == 0 {
			fmt.Printf("Processed block #%v\n", dBlock.GetDatabaseHeight())
		}
	}
}
