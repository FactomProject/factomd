// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
)

const level string = "level"
const bolt string = "bolt"

func main() {
	fmt.Println("Usage:")
	fmt.Println("FixBlockHeads level/bolt DBFileLocation")
	fmt.Println("Program will reset the block heads to the highest valid DBlock")

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

	dbo := databaseOverlay.NewOverlay(dbase, nil)
	err = FixBlockHeads(dbo)
	if err != nil {
		fmt.Errorf("ERROR: %v", err)
	}

	head, err := dbo.FetchDirectoryBlockHead()
	if err != nil {
		panic(err)
	}
	if head == nil {
		fmt.Printf("Head not found!\n")
	} else {
		fmt.Printf("Head - %v\n", head.String())
	}
}

func FixBlockHeads(dbo *databaseOverlay.Overlay) error {
	var prevs []*databaseOverlay.BlockSet
	var prev *databaseOverlay.BlockSet
	prev = nil
	for i := 0; ; i++ {
		if i%1000 == 0 {
			fmt.Printf("Processing block %v\n", i)
		}
		bs, err := dbo.FetchBlockSetByHeight(uint32(i))
		if err != nil {
			return err
		}
		if bs == nil {
			return nil
		}
		if prev != nil {
			prevKeyMR := bs.DBlock.GetHeader().GetPrevKeyMR()
			keyMR := prev.DBlock.GetKeyMR()
			if prevKeyMR.IsSameAs(keyMR) == false {
				return fmt.Errorf("KeyMR mismatch at height %v", i)
			}
		}

		prev = bs
		prevs = append(prevs, prev)

		//Ensuring we stay far away from the corrupted block head
		if len(prevs) > 50 {
			bs = prevs[0]
			chainIDs := []interfaces.IHash{}
			keyMRs := []interfaces.IHash{}

			chainIDs = append(chainIDs, bs.DBlock.GetChainID())
			keyMRs = append(keyMRs, bs.DBlock.DatabasePrimaryIndex())

			chainIDs = append(chainIDs, bs.ABlock.GetChainID())
			keyMRs = append(keyMRs, bs.ABlock.DatabasePrimaryIndex())

			chainIDs = append(chainIDs, bs.FBlock.GetChainID())
			keyMRs = append(keyMRs, bs.FBlock.DatabasePrimaryIndex())

			chainIDs = append(chainIDs, bs.ECBlock.GetChainID())
			keyMRs = append(keyMRs, bs.ECBlock.DatabasePrimaryIndex())

			for _, v := range bs.EBlocks {
				chainIDs = append(chainIDs, v.GetChainID())
				keyMRs = append(keyMRs, v.DatabasePrimaryIndex())
			}
			err = dbo.SetChainHeads(keyMRs, chainIDs)
			if err != nil {
				return err
			}
			prevs = prevs[1:]
		}
	}

	return nil
}
