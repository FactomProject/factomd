// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package systemState

import (
	"fmt"

	"github.com/FactomProject/factomd/blockchainState"
	"github.com/FactomProject/factomd/blockchainState/blockMaker"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

type BStateHandler struct {
	//Main, full BState
	MainBState *blockchainState.BlockchainState
	//BState for synching from the Genesis Block
	//BacklogBState *blockchainState.BlockchainState
	//BlockMaker for making the next set of blocks
	BlockMaker *blockMaker.BlockMaker
	//Database for storing new blocks and entries
	DB interfaces.DBOverlay

	//DBStateMsgs that have not been applied or dismissed yet
	PendingDBStateMsgs []*messages.DBStateMsg
	//Marking whether we're still synchronising with the network, or are we fully synched
	FullySynched bool
}

func (bh *BStateHandler) InitMainNet() {
	if bh.MainBState == nil {
		bh.MainBState = blockchainState.NewBSMainNet()
	}
}

func (bh *BStateHandler) HandleDBStateMsg(msg interfaces.IMsg) error {
	if msg.Type() != constants.DBSTATE_MSG {
		return fmt.Errorf("Invalid message type")
	}
	dbStateMsg := msg.(*messages.DBStateMsg)

	height := dbStateMsg.DirectoryBlock.GetDatabaseHeight()
	if bh.MainBState.DBlockHeight >= height {
		//Nothing to do - we're already ahead
		return nil
	}
	if bh.MainBState.DBlockHeight+1 < height {
		//DBStateMsg is too far ahead - ignore it for now
		bh.PendingDBStateMsgs = append(bh.PendingDBStateMsgs, dbStateMsg)
		return nil
	}

	tmpBState, err := bh.MainBState.Clone()
	if err != nil {
		return err
	}

	err = tmpBState.ProcessBlockSet(dbStateMsg.DirectoryBlock, dbStateMsg.AdminBlock, dbStateMsg.FactoidBlock, dbStateMsg.EntryCreditBlock,
		dbStateMsg.EBlocks, dbStateMsg.Entries)
	if err != nil {
		return err
	}

	err = bh.SaveBlockSetToDB(dbStateMsg.DirectoryBlock, dbStateMsg.AdminBlock, dbStateMsg.FactoidBlock, dbStateMsg.EntryCreditBlock,
		dbStateMsg.EBlocks, dbStateMsg.Entries)
	if err != nil {
		return err
	}

	bh.MainBState = tmpBState

	err = bh.SaveBState(bh.MainBState)
	if err != nil {
		return err
	}

	for i := len(bh.PendingDBStateMsgs) - 1; i >= 0; i-- {
		if bh.PendingDBStateMsgs[i].DirectoryBlock.GetDatabaseHeight() <= bh.MainBState.DBlockHeight {
			//We already dealt with this DBState, deleting the message
			bh.PendingDBStateMsgs = append(bh.PendingDBStateMsgs[:i], bh.PendingDBStateMsgs[i+1:])
		}
		if bh.PendingDBStateMsgs[i].DirectoryBlock.GetDatabaseHeight() == bh.MainBState.DBlockHeight+1 {
			//Next DBState to process - do it now
			err = bh.HandleDBStateMsg(bh.PendingDBStateMsgs[i])
			if err != nil {
				return err
			}
		}
	}

	//TODO: overwrite BlockMaker if appropriate

	return nil
}

func (bh *BStateHandler) SaveBState(bState *blockchainState.BlockchainState) error {
	//TODO: figure out how often we want to save BStates
	//TODO: do
	return nil
}

func (bh *BStateHandler) SaveBlockSetToDB(dBlock interfaces.IDirectoryBlock, aBlock interfaces.IAdminBlock, fBlock interfaces.IFBlock,
	ecBlock interfaces.IEntryCreditBlock, eBlocks []interfaces.IEntryBlock, entries []interfaces.IEBEntry) error {

	bh.DB.StartMultiBatch()

	err := bh.DB.ProcessDBlockMultiBatch(dBlock)
	if err != nil {
		bh.DB.CancelMultiBatch()
		return err
	}
	err = bh.DB.ProcessABlockMultiBatch(aBlock)
	if err != nil {
		bh.DB.CancelMultiBatch()
		return err
	}
	err = bh.DB.ProcessFBlockMultiBatch(fBlock)
	if err != nil {
		bh.DB.CancelMultiBatch()
		return err
	}
	err = bh.DB.ProcessECBlockMultiBatch(ecBlock)
	if err != nil {
		bh.DB.CancelMultiBatch()
		return err
	}
	for _, e := range eBlocks {
		err = bh.DB.ProcessEBlockMultiBatch(e)
		if err != nil {
			return err
		}
	}
	for _, e := range entries {
		err = bh.DB.InsertEntryMultiBatch(e)
		if err != nil {
			bh.DB.CancelMultiBatch()
			return err
		}
	}

	err = bh.DB.ExecuteMultiBatch()
	if err != nil {
		return err
	}
}

func (bh *BStateHandler) HandleCommitChainMsg() error {

	return nil
}

func (bh *BStateHandler) HandleCommitEntryMsg() error {

	return nil
}

func (bh *BStateHandler) HandleRevealEntryMsg() error {

	return nil
}
