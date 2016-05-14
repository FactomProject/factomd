// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"time"
)

var _ = fmt.Print

func LoadDatabase(s *State) {

	var blkCnt uint32

	head, err := s.GetDB().FetchDirectoryBlockHead()

	if err == nil && head != nil {
		blkCnt = head.GetHeader().GetDBHeight()
	}

	s.Println("Loading ", blkCnt, " Directory Blocks")

	msg, err := s.LoadDBState(blkCnt)

	for i := 0; true; i++ {
		if err != nil {
			s.Println(err.Error())
			break
		} else {
			if msg != nil {
				if len(s.InMsgQueue()) > 100 {
					for len(s.InMsgQueue()) > 30 {
						time.Sleep(100 * time.Millisecond)
					}
				}
				s.InMsgQueue() <- msg
			} else {
				break
			}
		}
		msg, err = s.LoadDBState(uint32(i))

		s.Print("\r", "\\|/-"[i%4:i%4+1])
	}

	if blkCnt == 0 && s.NetworkNumber == constants.NETWORK_LOCAL {
		s.Println("\n***********************************")
		s.Println("******* New Database **************")
		s.Println("***********************************\n")

		dblk := directoryBlock.NewDirectoryBlock(0, nil)
		ablk := s.NewAdminBlock(0)
		fblk := factoid.GetGenesisFBlock()
		ecblk := entryCreditBlock.NewECBlock()

		ablk.AddFedServer(primitives.Sha([]byte("FNode0")))

		msg := messages.NewDBStateMsg(s.GetTimestamp(), dblk, ablk, fblk, ecblk)
		s.InMsgQueue() <- msg
	}
	s.Println(fmt.Sprintf("Loaded %d directory blocks on %s", blkCnt, s.FactomNodeName))

}
