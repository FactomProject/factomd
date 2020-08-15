// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"strconv"

	"github.com/PaulSnow/factom2d/database/databaseOverlay"
	"github.com/PaulSnow/factom2d/database/hybridDB"
)

const level string = "level"
const bolt string = "bolt"

func main() {
	fmt.Println("Usage:")
	fmt.Println("SetChainHead level/bolt NewMaxHeight DBFileLocation")
	fmt.Println("Program will reset the highest directory block to the specified height")

	if len(os.Args) < 4 {
		fmt.Println("\nNot enough arguments passed")
		os.Exit(1)
	}
	if len(os.Args) > 4 {
		fmt.Println("\nToo many arguments passed")
		os.Exit(1)
	}

	levelBolt := os.Args[1]
	if levelBolt != level && levelBolt != bolt {
		fmt.Println("\nFirst argument should be `level` or `bolt`")
		os.Exit(1)
	}

	newHeightString := os.Args[2]
	newHeight, err2 := strconv.Atoi(newHeightString)
	if err2 != nil || newHeight < 0 {
		fmt.Println("\nSecond argument should be a number greater thatn zero instead of", newHeight)
		os.Exit(1)
	}

	path := os.Args[3]

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
	err = SetDirectoryBlockHead(dbo, newHeight)
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
		fmt.Printf("Head - \n%v\n", head.String())
	}
}

func SetDirectoryBlockHead(dbo *databaseOverlay.Overlay, newHeight int) error {

	fmt.Printf("new height being set to: %v\n", newHeight)

	dBlock, err := dbo.FetchDBlockHead()
	if err != nil {
		panic(err)
	}
	if dBlock == nil {
		panic("DBlock head not found")
	}

	if dBlock.GetHeader().GetDBHeight() < uint32(newHeight) {
		fmt.Printf("highest DB is %v but specified %v\n", dBlock.GetHeader().GetDBHeight(), newHeight)
		panic("not enough DBlocks to reset")
	}
	newHeadBlock, err := dbo.FetchDBlockByHeight(uint32(newHeight))
	if err != nil {
		panic(err)
	}
	if dBlock == nil {
		panic("DBlock head not found")
	}

	dbo.SaveDirectoryBlockHead(newHeadBlock)

	return nil
}
