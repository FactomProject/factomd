// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker

import (
	"github.com/FactomProject/factomd/blockchainState"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

type BlockMaker struct {
	ProcessedEBEntries  []*EBlockEntry
	ProcessedFBEntries  []interfaces.ITransaction
	ProcessedABEntries  []interfaces.IABEntry
	ProcessedECBEntries []*ECBlockEntry

	PendingMessages map[string]PendingMessages

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
	Message interfaces.IMsg
	Ack     interfaces.IMsg
}

type PendingMessages struct {
	DBHeight uint32

	LatestHeight uint32
	LatestAck    interfaces.IMsg

	PendingPairs []MsgAckPair
}

func (bm *BlockMaker) ProcessAckedMessage(msg interfaces.IMessageWithEntry, ack *messages.Ack) {

}
