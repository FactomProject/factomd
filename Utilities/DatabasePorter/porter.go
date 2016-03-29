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
		dbo.StartMultiBatch()

		err = dbo.ProcessDBlockMultiBatch(v)
		if err != nil {
			panic(err)
		}

		entries := v.GetDBEntries()
		for _, e := range entries {
			switch e.GetChainID().String() {
			case "000000000000000000000000000000000000000000000000000000000000000a":
				ablock, err := GetABlock(e.GetKeyMR().String())
				if err != nil {
					panic(err)
				}
				err = dbo.ProcessABlockMultiBatch(ablock)
				if err != nil {
					panic(err)
				}
				break
			case "000000000000000000000000000000000000000000000000000000000000000f":
				fblock, err := GetFBlock(e.GetKeyMR().String())
				if err != nil {
					panic(err)
				}
				err = dbo.ProcessFBlockMultiBatch(fblock)
				if err != nil {
					panic(err)
				}
				break
			case "000000000000000000000000000000000000000000000000000000000000000c":
				ecblock, err := GetECBlock(e.GetKeyMR().String())
				if err != nil {
					panic(err)
				}
				err = dbo.ProcessECBlockMultiBatch(ecblock)
				if err != nil {
					panic(err)
				}
				break
				//handle anchor block separately?
			default:
				eblock, err := GetEBlock(e.GetKeyMR().String())
				if err != nil {
					panic(err)
				}
				err = dbo.ProcessEBlockMultiBatch(eblock)
				if err != nil {
					panic(err)
				}
				entries := eblock.GetEntryHashes()
				for _, eHash := range entries {
					if eHash.IsMinuteMarker() == true {
						continue
					}
					entry, err := GetEntry(eHash.String())
					if err != nil {
						fmt.Printf("Problem getting entry %v from block %v\n", eHash.String(), e.GetKeyMR().String())
						panic(err)
					}
					err = dbo.InsertEntry(entry)
					if err != nil {
						panic(err)
					}
				}
				break
			}
		}

		if err := dbo.ExecuteMultiBatch(); err != nil {
			panic(err)
		}
		fmt.Printf("Saved block height %v\n", v.GetDatabaseHeight())
	}
}
