// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"os"
)

var _ = fmt.Print

func LoadDatabase(s *State) {
	var blkCnt uint32

	head, err := s.DB.FetchDirectoryBlockHead()
	if err == nil && head != nil {
		blkCnt = head.GetHeader().GetDBHeight()
	}

	t := time.Now()

	//msg, err := s.LoadDBState(blkCnt)

	for i := 0; true; i++ {
		if i > 0 && i%1000 == 0 {
			since := time.Since(t)
			ss := float64(since.Nanoseconds()) / 1000000000
			bps := float64(i) / ss
			os.Stderr.WriteString(fmt.Sprintf("%20s Loading Block %7d Blocks per second %8.2f\n", s.FactomNodeName, i, bps))
		}
		msg, err := s.LoadDBState(uint32(i))
		if err != nil {
			s.Println(err.Error())
			os.Stderr.WriteString(fmt.Sprintf("%20s Error reading database at block %d: %s\n", s.FactomNodeName, i, err.Error()))
			break
		} else {
			if msg != nil {
				msg.SetLocal(true)
				if len(s.InMsgQueue()) > 500 {
					for len(s.InMsgQueue()) > 200 {
						time.Sleep(10 * time.Millisecond)
					}
				}
				s.InMsgQueue() <- msg
			} else {
				// os.Stderr.WriteString(fmt.Sprintf("%20s Last Block in database: %d\n", s.FactomNodeName, i))
				break
			}
		}

		s.Print("\r", "\\|/-"[i%4:i%4+1])
	}

	if blkCnt == 0 {
		s.Println("\n***********************************")
		s.Println("******* New Database **************")
		s.Println("***********************************\n")

		dblk := directoryBlock.NewDirectoryBlock(nil)
		ablk := adminBlock.NewAdminBlock(nil)
		fblk := factoid.GetGenesisFBlock(s.GetNetworkID())
		ecblk := entryCreditBlock.NewECBlock()

		ablk.AddFedServer(primitives.Sha([]byte("FNode0")))

		dblk.SetABlockHash(ablk)
		dblk.SetECBlockHash(ecblk)
		dblk.SetFBlockHash(fblk)
		dblk.GetHeader().SetNetworkID(s.GetNetworkID())

		msg := messages.NewDBStateMsg(s.GetTimestamp(), dblk, ablk, fblk, ecblk)
		s.InMsgQueue() <- msg
	}
	s.Println(fmt.Sprintf("Loaded %d directory blocks on %s", blkCnt, s.FactomNodeName))
}
