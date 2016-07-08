package main

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
)

const level string = "level"
const bolt string = "bolt"

func main() {
	fmt.Println("Usage:")
	fmt.Println("DatabaseIntegrityCheck level/bolt DBFileLocation")
	fmt.Println("Database will be analysed for integrity errors")

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

	CheckDatabase(dbase)
}

func CheckDatabase(db interfaces.IDatabase) {
	if db == nil {
		return
	}

	dbo := databaseOverlay.NewOverlay(db)

	dBlock, err := dbo.FetchDBlockHead()
	if err != nil {
		panic(err)
	}
	if dBlock == nil {
		panic("DBlock head not found")
	}

	next := FetchBlockSet(dbo, dBlock.DatabasePrimaryIndex())

	var i int
	for {
		prev := FetchBlockSet(dbo, next.DBlock.GetHeader().GetPrevKeyMR())
		err = directoryBlock.CheckBlockPairIntegrity(next.DBlock, prev.DBlock)
		if err != nil {
			fmt.Printf("Error for DBlock %v %v - %v\n", next.DBlock.GetHeader().GetDBHeight(), next.DBlock.DatabasePrimaryIndex(), err)
		}
		err = adminBlock.CheckBlockPairIntegrity(next.ABlock, prev.ABlock)
		if err != nil {
			fmt.Printf("Error for ABlock %v %v - %v\n", next.ABlock.GetDatabaseHeight(), next.ABlock.DatabasePrimaryIndex(), err)
		}
		err = entryCreditBlock.CheckBlockPairIntegrity(next.ECBlock, prev.ECBlock)
		if err != nil {
			fmt.Printf("Error for ECBlock %v %v - %v\n", next.ECBlock.GetDatabaseHeight(), next.ECBlock.DatabasePrimaryIndex(), err)
		}
		err = factoid.CheckBlockPairIntegrity(next.FBlock, prev.FBlock)
		if err != nil {
			fmt.Printf("Error for FBlock %v %v - %v\n", next.FBlock.GetDatabaseHeight(), next.FBlock.DatabasePrimaryIndex(), err)
		}
		if prev.DBlock == nil {
			break
		}
		next = prev
		i++
	}

	fmt.Printf("Finished analysing %v sets of blocks\n", i)
}

type BlockSet struct {
	ABlock  interfaces.IAdminBlock
	ECBlock interfaces.IEntryCreditBlock
	FBlock  interfaces.IFBlock
	DBlock  interfaces.IDirectoryBlock
	//EBlocks
}

func FetchBlockSet(dbo interfaces.DBOverlay, dBlockHash interfaces.IHash) *BlockSet {
	bs := new(BlockSet)

	dBlock, err := dbo.FetchDBlock(dBlockHash)
	if err != nil {
		panic(err)
	}
	bs.DBlock = dBlock

	if dBlock == nil {
		return bs
	}
	entries := dBlock.GetDBEntries()
	for _, entry := range entries {
		switch entry.GetChainID().String() {
		case "000000000000000000000000000000000000000000000000000000000000000a":
			aBlock, err := dbo.FetchABlock(entry.GetKeyMR())
			if err != nil {
				panic(err)
			}
			bs.ABlock = aBlock
			break
		case "000000000000000000000000000000000000000000000000000000000000000c":
			ecBlock, err := dbo.FetchECBlock(entry.GetKeyMR())
			if err != nil {
				panic(err)
			}
			bs.ECBlock = ecBlock
			break
		case "000000000000000000000000000000000000000000000000000000000000000f":
			fBlock, err := dbo.FetchFBlock(entry.GetKeyMR())
			if err != nil {
				panic(err)
			}
			bs.FBlock = fBlock
			break
		default:
			break
		}
	}

	return bs
}
