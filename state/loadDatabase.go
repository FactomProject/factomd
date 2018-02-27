// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"time"

	"os"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = fmt.Print

func LoadDatabase(s *State) {

	var blkCnt uint32

	head, err := s.DB.FetchDBlockHead()
	if err == nil && head != nil {
		blkCnt = head.GetHeader().GetDBHeight()
	}

	last := time.Now()

	//msg, err := s.LoadDBState(blkCnt)
	start := s.GetDBHeightComplete()
	if start > 10 {
		start = start - 10
	}

	for i := int(start); i <= int(blkCnt); i++ {
		if i > 0 && i%1000 == 0 {
			bps := float64(1000) / time.Since(last).Seconds()
			os.Stderr.WriteString(fmt.Sprintf("%20s Loading Block %7d / %v. Blocks per second %8.2f\n", s.FactomNodeName, i, blkCnt, bps))
			last = time.Now()
		}

		msg, err := s.LoadDBState(uint32(i))
		if err != nil {
			s.Println(err.Error())
			os.Stderr.WriteString(fmt.Sprintf("%20s Error reading database at block %d: %s\n", s.FactomNodeName, i, err.Error()))
			break
		} else {
			if msg != nil {
				// We hold off EOM and other processing (s.Runleader) till the last DBStateMsg is executed.
				if i == int(blkCnt) {
					// last block, flag it.
					dbstate, _ := msg.(*messages.DBStateMsg)
					dbstate.IsLast = true // this is the last DBState in this load
					// this will cause s.DBFinished to go true
				}
				s.InMsgQueue().Enqueue(msg)
				msg.SetLocal(true)
				if s.InMsgQueue().Length() > 500 {
					for s.InMsgQueue().Length() > 100 {
						time.Sleep(10 * time.Millisecond)
					}
				}
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

		dblk, ablk, fblk, ecblk := GenerateGenesisBlocks(s.GetNetworkID())

		msg := messages.NewDBStateMsg(s.GetTimestamp(), dblk, ablk, fblk, ecblk, nil, nil, nil)
		// last block, flag it.
		dbstate, _ := msg.(*messages.DBStateMsg)
		dbstate.IsLast = true // this is the last DBState in this load
		// this will cause s.DBFinished to go true
		s.InMsgQueue().Enqueue(msg)
	}
	s.Println(fmt.Sprintf("Loaded %d directory blocks on %s", blkCnt, s.FactomNodeName))
}

func GenerateGenesisBlocks(networkID uint32) (interfaces.IDirectoryBlock, interfaces.IAdminBlock, interfaces.IFBlock, interfaces.IEntryCreditBlock) {
	dblk := directoryBlock.NewDirectoryBlock(nil)
	ablk := adminBlock.NewAdminBlock(nil)
	fblk := factoid.GetGenesisFBlock(networkID)
	ecblk := entryCreditBlock.NewECBlock()

	if networkID != constants.MAIN_NETWORK_ID {
		if networkID == constants.TEST_NETWORK_ID {
			ablk.AddFedServer(primitives.NewZeroHash())
		} else {
			ablk.AddFedServer(primitives.Sha([]byte("FNode0")))
		}
	} else {
		ecblk.GetBody().AddEntry(entryCreditBlock.NewServerIndexNumber())
		for i := 1; i < 11; i++ {
			minute := entryCreditBlock.NewMinuteNumber(uint8(i))
			ecblk.GetBody().AddEntry(minute)
		}
	}

	dblk.SetABlockHash(ablk)
	dblk.SetECBlockHash(ecblk)
	dblk.SetFBlockHash(fblk)
	dblk.GetHeader().SetNetworkID(networkID)

	dblk.GetHeader().SetTimestamp(primitives.NewTimestampFromMinutes(24018960))
	dblk.BuildBodyMR()

	return dblk, ablk, fblk, ecblk
}
