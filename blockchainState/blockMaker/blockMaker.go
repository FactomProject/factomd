// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker

import (
	"sync"

	"github.com/FactomProject/factomd/blockchainState"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

type BlockMaker struct {
	Mutex sync.RWMutex

	ProcessedEBEntries  []*EBlockEntry
	ProcessedFBEntries  []interfaces.ITransaction
	ProcessedABEntries  []interfaces.IABEntry
	ProcessedECBEntries []*ECBlockEntry

	PendingMessages map[string]*PendingMessages

	BState *blockchainState.BlockchainState

	ABlockHeaderExpansionArea []byte

	CurrentMinute int
}

func NewBlockMaker() *BlockMaker {
	bm := new(BlockMaker)
	bm.BState = blockchainState.NewBSLocalNet()
	return bm
}

func (bm *BlockMaker) SetCurrentMinute(m int) {
	bm.CurrentMinute = m
}

type MsgAckPair struct {
	Message interfaces.IMessageWithEntry
	Ack     *messages.Ack
}

type PendingMessages struct {
	Mutex sync.RWMutex

	DBHeight uint32

	LatestHeight uint32
	LatestAck    interfaces.IMsg

	PendingPairs []*MsgAckPair
}

func (bm *BlockMaker) GetPendingMessages(chainID interfaces.IHash) *PendingMessages {
	bm.Mutex.Lock()
	defer bm.Mutex.Unlock()

	pm := bm.PendingMessages[chainID.String()]
	if pm == nil {
		pm = new(PendingMessages)
		bm.PendingMessages[chainID.String()] = pm
	}

	return pm
}

func (bm *BlockMaker) ProcessAckedMessage(msg interfaces.IMessageWithEntry, ack *messages.Ack) {
	chainID := msg.GetEntryChainID()
	pm := bm.GetPendingMessages(chainID)

	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()

	if ack.Height < pm.LatestHeight {
		//We already processed this message, nothing to do
		return
	}
	if ack.Height == pm.LatestHeight {
		if pm.LatestAck != nil {
			//We already processed this message as well
			//AND it's not the first message!
			//Nothing to do
			return
		}
	}

	//Insert message into the slice, then process off of slice one by one
	//This is to reduce complexity of the code
	pair := new(MsgAckPair)
	pair.Ack = ack
	pair.Message = msg

	inserted := false
	for i := 0; i < len(pm.PendingPairs); i++ {
		//Looking for first pair that is higher than the current Height, so we can insert our pair before the other one
		if pm.PendingPairs[i].Ack.Height > pair.Ack.Height {
			index := i - 1
			if index < 0 {
				//Inserting as the first entry
				pm.PendingPairs = append([]*MsgAckPair{pair}, pm.PendingPairs...)
			} else {
				//Inserting somewhere in the middle
				pm.PendingPairs = append(pm.PendingPairs[:index], append([]*MsgAckPair{pair}, pm.PendingPairs[index:]...))
			}
			break
		}
	}
	if inserted == false {
		pm.PendingPairs = append(pm.PendingPairs, pair)
	}

	//Iterate over pending pairs and process them one by one until we're stuck
	for {
		if len(pm.PendingPairs) == 0 {
			break
		}
		if pm.LatestAck == nil {
			if pm.PendingPairs[0].Ack.Height != 0 {
				//We're expecting first message and we didn't find one
				break
			}
		} else {
			if pm.LatestHeight != pm.PendingPairs[0].Ack.Height-1 {
				//We didn't find the next pair
				break
			}
		}

		//Actually processing the message
		//TODO: do
	}
}
