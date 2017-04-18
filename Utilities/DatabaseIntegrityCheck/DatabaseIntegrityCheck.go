package main

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
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

	fmt.Printf("\tChecking block indexes\n")

	hashes, keys, err := dbo.GetAll(databaseOverlay.DIRECTORYBLOCK_NUMBER, primitives.NewZeroHash())
	for i, v := range hashes {
		h := v.(*primitives.Hash)
		if hashMap[h.String()] != "OK" {
			fmt.Printf("Invalid DBlock indexed at height 0x%x - %v\n", keys[i], h)
		}
	}

	hashes, keys, err = dbo.GetAll(databaseOverlay.FACTOIDBLOCK_NUMBER, primitives.NewZeroHash())
	for i, v := range hashes {
		h := v.(*primitives.Hash)
		if hashMap[h.String()] != "OK" {
			fmt.Printf("Invalid FBlock indexed at height 0x%x - %v\n", keys[i], h)
		}
	}

	hashes, keys, err = dbo.GetAll(databaseOverlay.ADMINBLOCK_NUMBER, primitives.NewZeroHash())
	for i, v := range hashes {
		h := v.(*primitives.Hash)
		if hashMap[h.String()] != "OK" {
			fmt.Printf("Invalid ABlock indexed at height 0x%x - %v\n", keys[i], h)
		}
	}

	hashes, keys, err = dbo.GetAll(databaseOverlay.ENTRYCREDITBLOCK_NUMBER, primitives.NewZeroHash())
	for i, v := range hashes {
		h := v.(*primitives.Hash)
		if hashMap[h.String()] != "OK" {
			fmt.Printf("Invalid ECBlock indexed at height 0x%x - %v\n", keys[i], h)
		}
	}

	fmt.Printf("\tFinished checking block indexes\n")

	fmt.Printf("\tLooking for free-floating blocks\n")

	dBlocks, err := dbo.FetchAllDBlockKeys()
	if err != nil {
		panic(err)
	}
	if len(dBlocks) != i {
		fmt.Printf("Found %v dBlocks, expected %v\n", len(dBlocks), i)
	}
	for _, block := range dBlocks {
		if hashMap[block.String()] == "" {
			fmt.Printf("Free-floating DBlock - %v\n", block.String())
		}
	}

	aBlocks, err := dbo.FetchAllABlockKeys()
	if err != nil {
		panic(err)
	}
	if len(aBlocks) != i {
		fmt.Printf("Found %v aBlocks, expected %v\n", len(aBlocks), i)
	}
	for _, block := range aBlocks {
		if hashMap[block.String()] == "" {
			fmt.Printf("Free-floating ABlock - %v\n", block.String())
		}
	}

	fBlocks, err := dbo.FetchAllFBlockKeys()
	if err != nil {
		panic(err)
	}
	if len(fBlocks) != i {
		fmt.Printf("Found %v fBlocks, expected %v\n", len(fBlocks), i)
	}
	for _, block := range fBlocks {
		if hashMap[block.String()] == "" {
			fmt.Printf("Free-floating FBlock - %v\n", block.String())
		}
	}

	ecBlocks, err := dbo.FetchAllECBlockKeys()
	if err != nil {
		panic(err)
	}
	if len(ecBlocks) != i {
		fmt.Printf("Found %v ecBlocks, expected %v\n", len(ecBlocks), i)
	}
	for _, block := range ecBlocks {
		if hashMap[block.String()] == "" {
			fmt.Printf("Free-floating ECBlock - %v\n", block.String())
		}
	}

	fmt.Printf("\tFinished looking for free-floating blocks\n")

	fmt.Printf("\tLooking for missing EBlocks\n")

	foundBlocks := 0
	for _, dHash := range dBlocks {
		dBlock, err := dbo.FetchDBlock(dHash)
		if err != nil {
			panic(err)
		}
		if dBlock == nil {
			fmt.Printf("Could not find DBlock %v!", dHash.String())
			panic("")
		}
		eBlockEntries := dBlock.GetEBlockDBEntries()
		for _, v := range eBlockEntries {
			eBlock, err := dbo.FetchEBlock(v.GetKeyMR())
			if err != nil {
				panic(err)
			}
			if eBlock == nil {
				fmt.Errorf("Could not find eBlock %v!\n", v.GetKeyMR())
			} else {
				foundBlocks++
			}
		}
	}

	fmt.Printf("\tFinished looking for missing EBlocks - found %v\n", foundBlocks)

	fmt.Printf("\tLooking for missing EBlock Entries\n")

	chains, err := dbo.FetchAllEBlockChainIDs()
	if err != nil {
		panic(err)
	}
	checkCount := 0
	missingCount := 0
	for _, chain := range chains {
		blocks, err := dbo.FetchAllEBlocksByChain(chain)
		if err != nil {
			panic(err)
		}
		if len(blocks) == 0 {
			panic("Found no blocks!")
		}
		for _, block := range blocks {
			entryHashes := block.GetEntryHashes()
			if len(entryHashes) == 0 {
				panic("Found no entryHashes!")
			}
			for _, eHash := range entryHashes {
				if eHash.IsMinuteMarker() == true {
					continue
				}
				entry, err := dbo.FetchEntry(eHash)
				if err != nil {
					panic(err)
				}
				if entry == nil {
					missingCount++
					fmt.Printf("Missing entry %v!\n", eHash.String())
				} else {
					checkCount++
				}
			}
		}
	}
	fmt.Printf("Found %v entries, missing %v\n", checkCount, missingCount)
	fmt.Printf("\tFinished looking for missing EBlock Entries\n")

	//CheckMinuteNumbers(dbo)
}

func CheckMinuteNumbers(dbo interfaces.DBOverlay) {
	fmt.Printf("\tChecking Minute Numbers\n")

	ecBlocks, err := dbo.FetchAllECBlocks()
	if err != nil {
		panic(err)
	}
	for _, v := range ecBlocks {
		entries := v.GetEntries()
		found := 0
		lastNumber := 0
		for _, e := range entries {
			if e.ECID() == entryCreditBlock.ECIDMinuteNumber {
				number := int(e.(*entryCreditBlock.MinuteNumber).Number)
				if number != lastNumber+1 {
					fmt.Printf("Block #%v %v, Minute Number %v is not last minute plus 1\n", v.GetDatabaseHeight(), v.GetHash().String(), number)
				}
				lastNumber = number
				found++
			}
		}
		if found != 10 {
			fmt.Printf("Block #%v %v only contains %v minute numbers\n", v.GetDatabaseHeight(), v.GetHash().String(), found)
		}
	}
	fmt.Printf("\tFinished checking Minute Numbers\n")
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
