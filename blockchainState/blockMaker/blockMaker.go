// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker

import (
	"github.com/FactomProject/factomd/blockchainState"
	"github.com/FactomProject/factomd/common/interfaces"
)

type BlockMaker struct {
	PendingEBEntries   []*EBlockEntry
	ProcessedEBEntries []*EBlockEntry

	PendingFBEntries   []interfaces.ITransaction
	ProcessedFBEntries []interfaces.ITransaction

	PendingABEntries   []interfaces.IABEntry
	ProcessedABEntries []interfaces.IABEntry

	PendingECBEntries   []*ECBlockEntry
	ProcessedECBEntries []*ECBlockEntry

	BState *blockchainState.BlockchainState

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
