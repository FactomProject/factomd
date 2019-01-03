package main

import (
	"flag"
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

var usage = "dbck [-b] DATABASE"

func main() {
	// parse the command line flags
	bflag := flag.Bool("b", false, "analize a bolt database")
	flag.Parse()
	os.Args = flag.Args()

	// open the database and create the database overlay
	db := func() *databaseOverlay.Overlay {
		if len(os.Args) < 1 {
			fmt.Println(usage)
			os.Exit(1)
		}
		path := os.Args[0]

		fmt.Println("Opening the database ", path)

		if *bflag {
			return databaseOverlay.NewOverlay(
				hybridDB.NewBoltMapHybridDB(nil, path),
			)
		}

		d, err := hybridDB.NewLevelMapHybridDB(path, false)
		if err != nil {
			fmt.Println("ERROR:", err)
		}

		return databaseOverlay.NewOverlay(d)
	}()

	next := func() *BlockSet {
		fmt.Println("Getting the inital block set")

		dbhead, err := db.FetchDBlockHead()
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		if dbhead == nil {
			fmt.Println("ERROR: Directory Block Head was nil")
		}
		return FetchBlockSet(db, dbhead.DatabasePrimaryIndex())
	}()

	// cycle through all the blocks
	for {
		height := next.DBlock.GetHeader().GetDBHeight()
		if height%1000 == 0 {
			fmt.Println("DBHeight: ", height)
			if height == 0 {
				fmt.Println("Finished")
				os.Exit(0)
			}
		}

		prev := FetchBlockSet(db, next.DBlock.GetHeader().GetPrevKeyMR())

		if err := directoryBlock.CheckBlockPairIntegrity(
			next.DBlock, prev.DBlock,
		); err != nil {
			fmt.Println("ERROR:", err)
		}

		if err := adminBlock.CheckBlockPairIntegrity(
			next.ABlock, prev.ABlock,
		); err != nil {
			fmt.Println("ERROR:", err)
		}

		if err := entryCreditBlock.CheckBlockPairIntegrity(
			next.ECBlock, prev.ECBlock,
		); err != nil {
			fmt.Println("ERROR:", err)
		}

		// TODO: check for duplicate fct transactions in the database
		if err := factoid.CheckBlockPairIntegrity(
			next.FBlock, prev.FBlock,
		); err != nil {
			fmt.Println("ERROR:", err)
		}

		if prev.DBlock == nil {
			break
		}
		next = prev
	}
}

// BlockSet is a set of the index blocks at a given database height
type BlockSet struct {
	ABlock  interfaces.IAdminBlock
	DBlock  interfaces.IDirectoryBlock
	ECBlock interfaces.IEntryCreditBlock
	FBlock  interfaces.IFBlock
	//EBlocks
}

// FetchBlockSet gets the set of index blocks at a given height from the
// database
func FetchBlockSet(dbo interfaces.DBOverlay, dBlockHash interfaces.IHash) *BlockSet {
	bs := new(BlockSet)

	dblock, err := dbo.FetchDBlock(dBlockHash)
	if err != nil {
		fmt.Println("ERROR:", err)
	} else if dblock == nil {
		fmt.Printf("ERROR: dblock %s was nil\n", dBlockHash)
	}
	bs.DBlock = dblock

	height := dblock.GetDatabaseHeight()

	// Get the various index blocks at the current database height.

	if ablock, err := dbo.FetchABlockByHeight(height); err != nil {
		fmt.Println("ERROR:", err)
	} else {
		bs.ABlock = ablock
	}

	if ecblock, err := dbo.FetchECBlockByHeight(height); err != nil {
		fmt.Println("ERROR:", err)
	} else {
		bs.ECBlock = ecblock
	}

	if fblock, err := dbo.FetchFBlockByHeight(height); err != nil {
		fmt.Println("ERROR:", err)
	} else {
		bs.FBlock = fblock
	}

	return bs
}
