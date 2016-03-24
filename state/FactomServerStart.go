// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid/block"
	"github.com/FactomProject/factomd/common/messages"
)

var _ = fmt.Print

func LoadDatabase(s *State) {

	var blkCnt uint32

	head, err := s.GetDB().FetchDirectoryBlockHead()

	if err == nil && head != nil {
		blkCnt = head.GetHeader().GetDBHeight()
	}

	fmt.Println("Loading ",blkCnt," Directory Blocks") 
	
	xmsg, err := s.LoadDBState(blkCnt)

	for i := 0; true; i++ {
		if err != nil {
			s.Println(err.Error())
			break
		} else {
			if msg != nil {
				s.InMsgQueue() <- msg
			} else {
				break
			}
		}
		msg, err = s.LoadDBState(uint32(i))
	}

	if blkCnt == 0 && s.NetworkNumber == constants.NETWORK_LOCAL {
		s.Println("\n***********************************")
		s.Println("******* New Database **************")
		s.Println("***********************************\n")

		dblk := directoryBlock.NewDirectoryBlock(0, nil)
		ablk := s.NewAdminBlock(0)
		fblk := block.GetGenesisFBlock()
		ecblk := entryCreditBlock.NewECBlock()

		msg := messages.NewDBStateMsg(s, dblk, ablk, fblk, ecblk)
		s.InMsgQueue() <- msg
	}
	s.Println(fmt.Sprintf("Loaded %d directory blocks on %s", blkCnt, s.FactomNodeName))

}
