package main

import (
	"fmt"
	"io/ioutil"
	//"os"

	. "github.com/FactomProject/factomd/blockchainState"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
)

const level string = "level"
const bolt string = "bolt"

func main() {
	fmt.Println("Usage:")
	fmt.Println("Test level/bolt DBFileLocation")
	/*
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
	*/
	levelBolt := "level"
	path := "C:/Users/ThePiachu/.factom/m2/main-database/ldb/MAIN/factoid_level.db"
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
	start := 0
	bs, err := LoadBS()
	if err != nil {
		bs = NewBSMainNet()
	} else {
		start = int(bs.DBlockHeight) + 1
	}
	bl, err := LoadBL()
	if err != nil {
		//panic(err)
	}
	if bl == nil {
		bl = new(BalanceLedger)
		bl.Init()
	}

	dBlock, err := dbo.FetchDBlockHead()
	if err != nil {
		panic(err)
	}
	if dBlock == nil {
		panic("DBlock head not found")
	}

	fmt.Printf("\tStarting\n")

	max := int(dBlock.GetDatabaseHeight())
	if max > 40000 {
		//max = 40000
	}

	for i := start; i < max; i++ {
		set := FetchBlockSet(dbo, i)
		if i%1000 == 0 {
			fmt.Printf("\"%v\", //%v\n", set.DBlock.DatabasePrimaryIndex(), set.DBlock.GetDatabaseHeight())
		}

		err := bs.ProcessBlockSet(set.DBlock, set.ABlock, set.FBlock, set.ECBlock, set.EBlocks, set.Entries)
		if err != nil {
			fmt.Printf("Error processing block set #%v\n", i)
			panic(err)
		}

		err = bl.ProcessFBlock(set.FBlock)
		if err != nil {
			panic(err)
		}

		if i%5000 == 0 {
			//periodically save and reload BS
			bin, err := bs.MarshalBinaryData()
			if err != nil {
				panic(err)
			}
			bs = new(BlockchainState)
			err = bs.UnmarshalBinaryData(bin)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Successfully saved and loaded BS\n")
			//if true {
			if i <= 85000 {
				err = SaveBS(bs)
				if err != nil {
					panic(err)
				}
				err = SaveBL(bl)
				if err != nil {
					panic(err)
				}
			}
		}
	}
	fmt.Printf("\tFinished!\n")

	b, err := bs.MarshalBinaryData()
	if err != nil {
		panic(err)
	}
	fmt.Printf("BS size - %v\n", len(b))

	b, err = bl.MarshalBinaryData()
	if err != nil {
		panic(err)
	}
	fmt.Printf("BL size - %v\n", len(b))

	fmt.Printf("Expired - %v\n", Expired)
	fmt.Printf("LatestReveal - %v\n", LatestReveal)
	fmt.Printf("TotalEntries - %v\n", TotalEntries)

	MES.Print()

	/*
		fmt.Printf("Balances\n")
		for _, v := range Balances {
			fmt.Printf("%v\t%v\n", v.TxID, v.Delta)
		}
	*/
}

type BlockSet struct {
	ABlock  interfaces.IAdminBlock
	ECBlock interfaces.IEntryCreditBlock
	FBlock  interfaces.IFBlock
	DBlock  interfaces.IDirectoryBlock
	EBlocks []interfaces.IEntryBlock
	Entries []interfaces.IEBEntry
}

func FetchBlockSet(dbo interfaces.DBOverlay, index int) *BlockSet {
	bs := new(BlockSet)

	dBlock, err := dbo.FetchDBlockByHeight(uint32(index))
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
			eBlock, err := dbo.FetchEBlock(entry.GetKeyMR())
			if err != nil {
				panic(err)
			}
			bs.EBlocks = append(bs.EBlocks, eBlock)
			/*
				for _, v := range eBlock.GetEntryHashes() {
					if v.IsMinuteMarker() {
						continue
					}
					e, err := dbo.FetchEntry(v)
					if err != nil {
						panic(err)
					}
					if e == nil {
						panic("Couldn't find entry " + v.String())
					}
					bs.Entries = append(bs.Entries, e)
				}
			*/
			break
		}
	}

	return bs
}

func SaveBS(bs *BlockchainState) error {
	b, err := bs.MarshalBinaryData()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("bs.test", b, 0644)
	if err != nil {
		return err
	}
	fmt.Printf("Saved BS\n")

	return nil
}

func LoadBS() (*BlockchainState, error) {
	b, err := ioutil.ReadFile("bs.test")
	if err != nil {
		return nil, err
	}
	bs := NewBSMainNet()
	err = bs.UnmarshalBinaryData(b)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Loaded BS\n")
	return bs, nil
}

func SaveBL(bl *BalanceLedger) error {
	b, err := bl.MarshalBinaryData()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("bl.test", b, 0644)
	if err != nil {
		return err
	}
	fmt.Printf("Saved BL\n")

	return nil
}

func LoadBL() (*BalanceLedger, error) {
	b, err := ioutil.ReadFile("bl.test")
	if err != nil {
		return nil, err
	}
	bl := new(BalanceLedger)
	bl.Init()
	err = bl.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Loaded BL\n")
	return bl, nil
}
