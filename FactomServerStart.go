// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/factoid/block"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/state"
)

var _ = fmt.Print

func FactomServerStart(state *state.State) {

	state.Init()
	go loadDatabase(state)
	// Timer runs periodically, and inserts eom messages into the stream
	go Timer(state)
	// Validator is the gateway.
	go Validator(state)

}

func loadDatabase(s *state.State) {
	
	var blkCnt uint32
	
	dblks, err := s.DB.FetchAllDBlocks()
	if err != nil {
		panic(err)
	}
	
	s.Println("Directory Blocks Found:", len(dblks))
	
	blkCnt = uint32(len(dblks))
	for i := int(blkCnt)-1; i >=0; i-- {
				
		dblk := dblks[i]
	
		ablk, err := s.DB.FetchABlockByKeyMR(dblk.GetDBEntries()[0].GetKeyMR())
		if err != nil {
			panic(err)
		}
		if ablk == nil {
			panic("ablk is nil" + dblk.GetDBEntries()[0].GetKeyMR().String())
		}
		ecblk, err := s.DB.FetchECBlockByHash(dblk.GetDBEntries()[1].GetKeyMR())
		if err != nil {
			panic(err)
		}
		if ecblk == nil {
			panic("ecblk is nil - " + dblk.GetDBEntries()[1].GetKeyMR().String())
		}
		fblk, err := s.DB.FetchFBlockByKeyMR(dblk.GetDBEntries()[2].GetKeyMR())
		if err != nil {
			panic(err)
		}
		if fblk == nil {
			panic("fblk is nil" + dblk.GetDBEntries()[2].GetKeyMR().String())
		}
		
		msg := messages.NewDBStateMsg(s,dblk,ablk,fblk, ecblk)
		
		s.InMsgQueue() <- msg
	}
	
	if blkCnt == 0 && s.NetworkNumber == constants.NETWORK_LOCAL {
		s.Println("\n***********************************")
		s.Println("******* New Database **************")
		s.Println("***********************************\n")
		
		dblk := directoryBlock.NewDirectoryBlock(0, nil)
		ablk := s.NewAdminBlock(0)
		fblk := block.GetGenesisFBlock()
		ecblk := entryCreditBlock.NewECBlock()
		
		msg := messages.NewDBStateMsg(s, dblk,ablk,fblk, ecblk)
		
		s.InMsgQueue() <- msg
	}
	s.Println(fmt.Sprintf("Loaded %d directory blocks", blkCnt))

}
