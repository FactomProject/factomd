// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/database/blockExtractor"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
)

const level string = "level"
const bolt string = "bolt"

func main() {
	fmt.Println("Usage:")
	fmt.Println("BlockExtractor level/bolt [ChainID-To-Extract]")
	fmt.Println("Leave out the last one to export basic chains (A, D, EC, F)")
	if len(os.Args) < 1 {
		fmt.Println("\nNot enough arguments passed")
		os.Exit(1)
	}
	if len(os.Args) > 2 {
		fmt.Println("\nToo many arguments passed")
		os.Exit(1)
	}

	levelBolt := os.Args[1]

	if levelBolt != level && levelBolt != bolt {
		fmt.Println("\nFirst argument should be `level` or `bolt`")
		os.Exit(1)
	}

	chainID := ""
	if len(os.Args) == 3 {
		chainID = os.Args[2]
	}

	be := new(BlockExtractor)

	state := new(state.State)
	state.Cfg = util.ReadConfig("")
	if levelBolt == level {
		err := state.InitLevelDB()
		if err != nil {
			panic(err)
		}
	}
	if levelBolt == bolt {
		err := state.InitBoltDB()
		if err != nil {
			panic(err)
		}
	}
	dbo := state.GetDB().(interfaces.DBOverlay)

	if chainID != "" {
		err := be.ExportEChain(chainID, dbo)
		if err != nil {
			panic(err)
		}
	} else {
		err := be.ExportDChain(dbo)
		if err != nil {
			panic(err)
		}
		err = be.ExportECChain(dbo)
		if err != nil {
			panic(err)
		}
		err = be.ExportAChain(dbo)
		if err != nil {
			panic(err)
		}
		err = be.ExportFctChain(dbo)
		if err != nil {
			panic(err)
		}
		err = be.ExportDirBlockInfo(dbo)
		if err != nil {
			panic(err)
		}
	}
}
