// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func (bs *BlockchainState) ProcessECBlock(ecBlock interfaces.IEntryCreditBlock) error {
	bs.Init()

	if bs.ECBlockHead.KeyMR.String() != ecBlock.GetHeader().GetPrevHeaderHash().String() {
		return fmt.Errorf("Invalid ECBlock %v previous KeyMR - expected %v, got %v\n", ecBlock.DatabasePrimaryIndex().String(), bs.ECBlockHead.KeyMR.String(), ecBlock.GetHeader().GetPrevHeaderHash().String())
	}
	if bs.ECBlockHead.Hash.String() != ecBlock.GetHeader().GetPrevFullHash().String() {
		return fmt.Errorf("Invalid ECBlock %v previous hash - expected %v, got %v\n", ecBlock.DatabasePrimaryIndex().String(), bs.ECBlockHead.Hash.String(), ecBlock.GetHeader().GetPrevFullHash().String())
	}

	if bs.DBlockHeight > M2SWITCHHEIGHT {
		//Only check in M2, since that's when this error got fixed
		if bs.DBlockHeight != ecBlock.GetDatabaseHeight() {
			return fmt.Errorf("Invalid ECBlock height - expected %v, got %v", bs.DBlockHeight, ecBlock.GetDatabaseHeight())
		}
	}

	bs.ECBlockHead.KeyMR = ecBlock.DatabasePrimaryIndex().(*primitives.Hash)
	bs.ECBlockHead.Hash = ecBlock.DatabaseSecondaryIndex().(*primitives.Hash)

	err := CheckECBlockMinuteNumbers(ecBlock)
	if err != nil {
		return err
	}

	entries := ecBlock.GetEntries()
	for _, v := range entries {
		err := bs.ProcessECEntry(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bs *BlockchainState) ProcessECEntry(v interfaces.IECBlockEntry) error {
	bs.Init()
	switch v.ECID() {
	case entryCreditBlock.ECIDBalanceIncrease:
		//Deprecated
		break
	case entryCreditBlock.ECIDEntryCommit:
		e := v.(*entryCreditBlock.CommitEntry)
		if bs.ECBalances[e.ECPubKey.String()] < int64(e.Credits) {
			if bs.NetworkID != constants.MAIN_NETWORK_ID || bs.DBlockHeight > 93719 {
				return fmt.Errorf("Not enough ECs - %v:%v<%v", e.ECPubKey.String(), bs.ECBalances[e.ECPubKey.String()], uint64(e.Credits))
			}
		}
		bs.ECBalances[e.ECPubKey.String()] = bs.ECBalances[e.ECPubKey.String()] - int64(e.Credits)
		bs.PushCommit(e.GetEntryHash(), v.Hash())
		break
	case entryCreditBlock.ECIDChainCommit:
		e := v.(*entryCreditBlock.CommitChain)
		if bs.ECBalances[e.ECPubKey.String()] < int64(e.Credits) {
			if bs.NetworkID != constants.MAIN_NETWORK_ID || bs.DBlockHeight > 93719 {
				return fmt.Errorf("Not enough ECs - %v:%v<%v", e.ECPubKey.String(), bs.ECBalances[e.ECPubKey.String()], uint64(e.Credits))
			}
		}
		bs.ECBalances[e.ECPubKey.String()] = bs.ECBalances[e.ECPubKey.String()] - int64(e.Credits)
		bs.PushCommit(e.GetEntryHash(), v.Hash())
		break
	default:
		break
	}
	return nil
}

func CheckECBlockMinuteNumbers(ecBlock interfaces.IEntryCreditBlock) error {
	//Check whether MinuteNumbers are increasing
	entries := ecBlock.GetEntries()

	var lastMinute uint8 = 0
	for i, v := range entries {
		if v.ECID() == entryCreditBlock.ECIDMinuteNumber {
			minute := v.(*entryCreditBlock.MinuteNumber).Number
			if minute < 1 || minute > 10 {
				return fmt.Errorf("ECBlock Invalid minute number at position %v - %v", i, minute)
			}
			if minute <= lastMinute {
				return fmt.Errorf("ECBlock Invalid minute number at position %v - %v", i, minute)
			}
			lastMinute = minute
		}
	}

	return nil
}
