// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/receipts"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
)

const level string = "level"
const bolt string = "bolt"

func main() {
	fmt.Println("Usage:")
	fmt.Println("ReceiptGenerator level/bolt [EntryID-To-Extract]")
	fmt.Println("Leave out the last one to export all entries")
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

	entryID := ""
	if len(os.Args) == 3 {
		entryID = os.Args[2]
	}

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

	if entryID != "" {
		err := ExportEntryReceipt(entryID, dbo)
		if err != nil {
			panic(err)
		}
	} else {
		err := ExportAllEntryReceipts(dbo)
		if err != nil {
			panic(err)
		}
	}
}
