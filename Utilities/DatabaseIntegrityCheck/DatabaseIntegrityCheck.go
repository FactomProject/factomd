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

	fmt.Printf("\tStarting consecutive block analysis\n")

	hashMap := map[string]string{}

	var i int
	for {
		/*
			if next.DBlock.GetDatabaseHeight()%1000 == 0 {
				fmt.Printf("\"%v\", //%v\n", next.DBlock.DatabasePrimaryIndex(), next.DBlock.GetDatabaseHeight())
			}
		*/
		prev := FetchBlockSet(dbo, next.DBlock.GetHeader().GetPrevKeyMR())

		hashMap[next.DBlock.DatabasePrimaryIndex().String()] = "OK"
		err = directoryBlock.CheckBlockPairIntegrity(next.DBlock, prev.DBlock)
		if err != nil {
			fmt.Printf("Error for DBlock %v %v - %v\n", next.DBlock.GetHeader().GetDBHeight(), next.DBlock.DatabasePrimaryIndex(), err)
		}

		hashMap[next.ABlock.DatabasePrimaryIndex().String()] = "OK"
		err = adminBlock.CheckBlockPairIntegrity(next.ABlock, prev.ABlock)
		if err != nil {
			fmt.Printf("Error for ABlock %v %v - %v\n", next.ABlock.GetDatabaseHeight(), next.ABlock.DatabasePrimaryIndex(), err)
		}

		hashMap[next.ECBlock.DatabasePrimaryIndex().String()] = "OK"
		err = entryCreditBlock.CheckBlockPairIntegrity(next.ECBlock, prev.ECBlock)
		if err != nil {
			fmt.Printf("Error for ECBlock %v %v - %v\n", next.ECBlock.GetDatabaseHeight(), next.ECBlock.DatabasePrimaryIndex(), err)
		}

		hashMap[next.FBlock.DatabasePrimaryIndex().String()] = "OK"
		err = factoid.CheckBlockPairIntegrity(next.FBlock, prev.FBlock)
		if err != nil {
			fmt.Printf("Error for FBlock %v %v - %v\n", next.FBlock.GetDatabaseHeight(), next.FBlock.DatabasePrimaryIndex(), err)
		}

		i++
		if prev.DBlock == nil {
			break
		}
		next = prev
	}

	fmt.Printf("\tFinished analysing %v sets of blocks\n", i)

	fmt.Printf("\tLooking for free-floating blocks\n")

	dBlocks, err := dbo.FetchAllDBlocks()
	if err != nil {
		panic(err)
	}
	if len(dBlocks) != i {
		fmt.Printf("Found %v dBlocks, expected %v\n", len(dBlocks), i)
	}
	for _, block := range dBlocks {
		if hashMap[block.DatabasePrimaryIndex().String()] == "" {
			fmt.Printf("Free-floating DBlock - %v, %v\n", block.DatabasePrimaryIndex().String(), block.String())
		}
	}

	aBlocks, err := dbo.FetchAllABlocks()
	if err != nil {
		panic(err)
	}
	if len(aBlocks) != i {
		fmt.Printf("Found %v aBlocks, expected %v\n", len(aBlocks), i)
	}
	for _, block := range aBlocks {
		if hashMap[block.DatabasePrimaryIndex().String()] == "" {
			fmt.Printf("Free-floating ABlock - %v, %v\n", block.DatabasePrimaryIndex().String(), block.String())
		}
	}

	fBlocks, err := dbo.FetchAllFBlocks()
	if err != nil {
		panic(err)
	}
	if len(fBlocks) != i {
		fmt.Printf("Found %v fBlocks, expected %v\n", len(fBlocks), i)
	}
	for _, block := range fBlocks {
		if hashMap[block.DatabasePrimaryIndex().String()] == "" {
			fmt.Printf("Free-floating FBlock - %v, %v\n", block.DatabasePrimaryIndex().String(), block.String())
		}
	}

	ecBlocks, err := dbo.FetchAllECBlocks()
	if err != nil {
		panic(err)
	}
	if len(ecBlocks) != i {
		fmt.Printf("Found %v ecBlocks, expected %v\n", len(ecBlocks), i)
	}
	for _, block := range ecBlocks {
		if hashMap[block.DatabasePrimaryIndex().String()] == "" {
			fmt.Printf("Free-floating ECBlock - %v, %v\n", block.DatabasePrimaryIndex().String(), block.String())
		}
	}

	fmt.Printf("\tFinished looking for free-floating blocks\n")
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
