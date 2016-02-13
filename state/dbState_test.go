// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid/block"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/util"
	"testing"
)

var _ = fmt.Print
var _ = log.Print
var _ = util.ReadConfig

func Test_DBState1(t *testing.T) {
	fmt.Println("DBState Test")
	defer fmt.Println("DBState Test Done")

	off := true			// We need to create states with a testing database!
	if off {
		return
	}
	
	
	state := GetState()

	var prev interfaces.IDirectoryBlock	// First call gets a nil, rest the previous DirectoryBlock
	
	var i uint32
	for i = 0; i < 2; i++ {

		dblk := directoryBlock.NewDirectoryBlock(i,prev.(*directoryBlock.DirectoryBlock))
		prev = dblk
		ablk := state.NewAdminBlock(i)
		eblk := entryCreditBlock.NewECBlock()
		fblk := block.GetGenesisFBlock()

		state.DBStates.NewDBState(true, dblk, ablk, fblk, eblk)

		h := dblk.GetHeader().GetDBHeight()
		if i != h {
			t.Errorf("Height error.  Expecting %d and got %d", i, h)
		}
		if state.DBHeight != i {
			t.Errorf("DBHeight error.  Expecting %d and got %d", i, state.DBHeight)
		}
	}

	fmt.Println("Testing all blocks")

	dblks := make([]interfaces.IDirectoryBlock, 0)

	for j := uint32(0); j < i; j++ {
		fmt.Print("\r At Block: ", j)
		dblk, _ := state.DB.FetchDBlockByHeight(j)
		fmt.Println(dblk.String())
		if dblk == nil {
			fmt.Println("last dblk found:", j)
			break
		}
		dblks = append(dblks, dblk)
	}

	/*
		 * ecblkHash := dblks[len(dblks)-1].DBEntries[1].KeyMR

			i := 0
			for i = 0; ecblkHash != nil; i++ {
				fmt.Printf(" %x\n",ecblkHash.Bytes())
				ecblk, _ := db.FetchECBlockByHash(ecblkHash)
				if ecblk == nil {
					break
				}
				ecblkHash = ecblk.Header.PrevHeaderHash
			}
			fmt.Println ("End found after",i,"ec blocks")
		}
	*/
}
