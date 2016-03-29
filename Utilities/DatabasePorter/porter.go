// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/util"
)

func main() {
	fmt.Println("DatabasePorter")

	cfg := util.ReadConfig("")

	var dbo interfaces.DBOverlay

	switch cfg.App.DBType {
	case "Bolt":
		dbo = InitBolt(cfg)
		break
	case "LDB":
		dbo = InitLevelDB(cfg)
		break
	default:
		dbo = InitMapDB(cfg)
		break
	}

	fmt.Printf("dbo - %v\n", dbo)

	keymr, err := GetDBlockHead()
	if err != nil {
		panic(err)
	}
	dBlock, err := GetDBlock(keymr)
	if err != nil {
		panic(err)
	}
	if dBlock == nil {
		panic("dblock head not found")
	}
	dBlockList := make([]interfaces.IDirectoryBlock, int(dBlock.GetDatabaseHeight())+1)
	dBlockList[int(dBlock.GetDatabaseHeight())] = dBlock

	for {
		keymr = dBlock.GetHeader().GetPrevKeyMR().String()
		if keymr == "0000000000000000000000000000000000000000000000000000000000000000" {
			break
		}
		dBlock, err = GetDBlock(keymr)
		if err != nil {
			panic(err)
		}
		if dBlock == nil {
			panic("dblock " + keymr + " not found")
		}
		dBlockList[int(dBlock.GetDatabaseHeight())] = dBlock
		fmt.Printf("Fetched dblock %v\n", dBlock.GetDatabaseHeight())
	}

	for _, v := range dBlockList {
		entries := v.GetDBEntries()
		for _, e := range entries {
			switch e.GetChainID().String() {
			case "000000000000000000000000000000000000000000000000000000000000000a":
				break
			case "000000000000000000000000000000000000000000000000000000000000000f":
				break
			case "000000000000000000000000000000000000000000000000000000000000000c":
				break
				//handle anchor block separately?
			default:
				break
			}
		}
		fmt.Printf("Saved block height %v", dBlock.GetDatabaseHeight())
	}
}
