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
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
)

var usage = "dbck [-bf] DATABASE"

func main() {
	// parse the command line flags
	bflag := flag.Bool("b", false, "analize a bolt database")
	fflag := flag.Bool("f", false, "do not check for duplicate factoid transactions")
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

	// collect transaction hashes to detect duplicate transactions in the
	// database
	fcthashes := make(map[[32]byte]uint32)
	sighashes := make(map[[32]byte]uint32)

	// keep a list of read blocks to check against the DB index
	blkMap := make(map[[32]byte]bool)

	// cycle through all the blocks
	for {
		height := next.DBlock.GetHeader().GetDBHeight()
		if height%1000 == 0 {
			fmt.Println("DBHeight: ", height)
			if height == 0 {
				break
			}
		}

		prev := FetchBlockSet(db, next.DBlock.GetHeader().GetPrevKeyMR())

		if err := directoryBlock.CheckBlockPairIntegrity(
			next.DBlock, prev.DBlock,
		); err != nil {
			fmt.Println("ERROR: DBlock:", height, err)
		} else {
			blkMap[next.DBlock.DatabasePrimaryIndex().Fixed()] = true
		}

		if err := adminBlock.CheckBlockPairIntegrity(
			next.ABlock, prev.ABlock,
		); err != nil {
			fmt.Println("ERROR: ABlock:", height, err)
		} else {
			blkMap[next.ABlock.DatabasePrimaryIndex().Fixed()] = true
		}

		if err := entryCreditBlock.CheckBlockPairIntegrity(
			next.ECBlock, prev.ECBlock,
		); err != nil {
			fmt.Println("ERROR: ECBlock:", height, err)
		} else {
			blkMap[next.ECBlock.DatabasePrimaryIndex().Fixed()] = true
		}

		if err := factoid.CheckBlockPairIntegrity(
			next.FBlock, prev.FBlock,
		); err != nil {
			fmt.Println("ERROR: FBlock:", height, err)
		} else {
			blkMap[next.FBlock.DatabasePrimaryIndex().Fixed()] = true
		}

		if !*fflag {
			// check for duplicate factoid transactions
			for _, fct := range next.FBlock.GetEntryHashes() {
				if h, exists := fcthashes[fct.Fixed()]; exists {
					fmt.Printf(
						"ERROR: duplicate transactions found at heights %d and %d\n",
						h, height,
					)
				} else {
					// add the fcthash to the list
					fcthashes[fct.Fixed()] = height
				}
			}

			// check for duplicate transaction signatures hashes
			for _, sig := range next.FBlock.GetEntrySigHashes() {
				if h, exists := sighashes[sig.Fixed()]; exists {
					fmt.Printf(
						"ERROR: duplicate tx signatures found at height %d and %d\n",
						h, height,
					)
				} else {
					// add the sighash to the list
					sighashes[sig.Fixed()] = height
				}
			}
		}

		if prev.DBlock == nil {
			break
		}
		next = prev
	}

	if !*fflag {
		fmt.Println("total factoid transactions: ", len(fcthashes))
	}

	// check indexes for DBlocks
	fmt.Println("Checking Drirectory Block index")
	{
		hs, ks, err := db.GetAll(
			databaseOverlay.DIRECTORYBLOCK_NUMBER,
			primitives.NewZeroHash(),
		)
		if err != nil {
			fmt.Println(err)
		}
		for i, v := range hs {
			h := v.(*primitives.Hash)
			if !blkMap[h.Fixed()] {
				fmt.Println("Invalid DBlock indexed", ks[i], h)
			}
		}
	}

	fmt.Println("Finished")
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
