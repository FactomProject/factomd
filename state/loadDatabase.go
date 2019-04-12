// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"math"
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

func humanizeDuration(duration time.Duration) string {
	hours := int64(duration.Hours())
	minutes := int64(math.Mod(duration.Minutes(), 60))
	seconds := int64(math.Mod(duration.Seconds(), 60))

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

func LoadDatabase(s *State) {
	var blkCnt uint32

	head, err := s.DB.FetchDBlockHead()
	if err == nil && head != nil {
		blkCnt = head.GetHeader().GetDBHeight()
	}
	// prevent MMR processing from happening for blocks being loaded from the database
	s.DBHeightAtBoot = blkCnt

	first := time.Now()
	last := first
	time.Sleep(time.Second)

	//msg, err := s.LoadDBState(blkCnt)
	start := s.GetDBHeightComplete()

	if start > 0 {
		start++
	}

	for i := int(start); i <= int(blkCnt); i++ {
		if i > int(start)+500 && i%1000 == 0 {
			seconds := time.Since(last).Seconds()
			bps := float64(1000) / seconds
			f := time.Since(first).Seconds()
			abps := float64(i-int(start)) / f
			timeUsed := time.Since(first)

			blocksRemaining := float64(blkCnt) - float64(i)
			timeRemaining := time.Duration(blocksRemaining/abps) * time.Second

			fmt.Fprintf(os.Stderr, "%20s Loading Block %7d / %v. Blocks per second %8.2f average bps %8.2f Progress %v remaining %v Estimated Total Time: %v \n", s.FactomNodeName, i, blkCnt, bps, abps,
				humanizeDuration(timeUsed), humanizeDuration(timeRemaining), humanizeDuration(timeUsed+timeRemaining))
			last = time.Now()
			// height := s.GetLLeaderHeight()
			// fmt.Fprintf(os.Stderr, "%20s Federated: DBH: %8d, Feds %d, audits: %d \n", s.FactomNodeName, height, len(s.GetFedServers(height)), len(s.GetAuditServers(height)))
		}

		msg, err := s.LoadDBState(uint32(i))
		if err != nil {
			s.Println(err.Error())
			os.Stderr.WriteString(fmt.Sprintf("%20s Error reading database at block %d: %s\n", s.FactomNodeName, i, err.Error()))
			break
		}
		if msg != nil {
			// We hold off EOM and other processing (s.Runleader) till the last DBStateMsg is executed.
			if i == int(blkCnt) {
				// last block, flag it.
				dbstate, _ := msg.(*messages.DBStateMsg)
				dbstate.IsLast = true // this is the last DBState in this load
				// this will cause s.DBFinished to go true
			}

			s.LogMessage("InMsgQueue", "enqueue", msg)
			msg.SetLocal(true)
			s.InMsgQueue().Enqueue(msg)
			if s.InMsgQueue().Length() > 200 || len(s.DBStatesReceived) > 50 {
				for s.InMsgQueue().Length() > 50 || len(s.DBStatesReceived) > 50 {
					time.Sleep(100 * time.Millisecond)
				}
			}
		} else {
			// os.Stderr.WriteString(fmt.Sprintf("%20s Last Block in database: %d\n", s.FactomNodeName, i))
			break
		}

		s.Print("\r", "\\|/-"[i%4:i%4+1])
	}

	if blkCnt == 0 {
		s.Println("\n***********************************")
		s.Println("******* New Database **************")
		s.Println("***********************************\n")

		var customIdentity interfaces.IHash
		if s.Network == "CUSTOM" {
			customIdentity, err = primitives.HexToHash(s.CustomBootstrapIdentity)
			if err != nil {
				panic(fmt.Sprintf("Could not decode Custom Bootstrap Identity (likely in config file) found: %s\n", s.CustomBootstrapIdentity))
			}
		}
		dblk, ablk, fblk, ecblk := GenerateGenesisBlocks(s.GetNetworkID(), customIdentity)

		messages.LogPrintf("marshalsizes.txt", "FBlock unmarshaled transaction count: %d", len(fblk.GetTransactions()))

		msg := messages.NewDBStateMsg(s.GetTimestamp(), dblk, ablk, fblk, ecblk, nil, nil, nil)
		// last block, flag it.
		dbstate, _ := msg.(*messages.DBStateMsg)
		dbstate.IsLast = true // this is the last DBState in this load
		// this will cause s.DBFinished to go true
		s.InMsgQueue().Enqueue(msg)
	}
	s.Println(fmt.Sprintf("Loaded %d directory blocks on %s", blkCnt, s.FactomNodeName))
	fmt.Fprintf(os.Stderr, "%20s Loading complete %v.\n", s.FactomNodeName, blkCnt)
}

func GenerateGenesisBlocks(networkID uint32, bootstrapIdentity interfaces.IHash) (interfaces.IDirectoryBlock, interfaces.IAdminBlock, interfaces.IFBlock, interfaces.IEntryCreditBlock) {
	dblk := directoryBlock.NewDirectoryBlock(nil)
	ablk := adminBlock.NewAdminBlock(nil)
	fblk := factoid.GetGenesisFBlock(networkID)
	ecblk := entryCreditBlock.NewECBlock()

	//decide if a default server identity needs to be added into the genesis block
n:
	switch networkID {
	case constants.MAIN_NETWORK_ID:
		//no identities were added in the M1 genesis block
		break n
	case constants.TEST_NETWORK_ID:
		ablk.AddFedServer(primitives.NewZeroHash())
	case constants.LOCAL_NETWORK_ID:
		ablk.AddFedServer(primitives.Sha([]byte("FNode0")))
	default: //we must be using a custom network, where the network ID is based on the customnet Name, not one of the three predetermined values above
		// add the config file identity to the genesis block
		ablk.AddFedServer(bootstrapIdentity)
	}

	//for the mainnet genesis block, support the original minute markers and server index markers from M1
	//don't add these items to a genesis block generated with M2 code
	if networkID == constants.MAIN_NETWORK_ID {
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
