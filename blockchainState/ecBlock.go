// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"fmt"

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
		err := bs.ProcessECEntries(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bs *BlockchainState) ProcessECEntries(v interfaces.IECBlockEntry) error {
	bs.Init()
	switch v.ECID() {
	case entryCreditBlock.ECIDBalanceIncrease:
		e := v.(*entryCreditBlock.IncreaseBalance)

		fTxID := e.TXID.String()
		index := e.Index
		key := fmt.Sprintf("%v:%v", fTxID, index)

		if bs.PendingECBalanceIncreases[key] == nil {
			return fmt.Errorf("EC Balance Increase exists without a proper factoid transaction")
		}
		if bs.PendingECBalanceIncreases[key].ECPubKey != e.ECPubKey.String() {
			return fmt.Errorf("Invalid ECPubKey")
		}
		if bs.PendingECBalanceIncreases[key].FactoidTxID != fTxID {
			return fmt.Errorf("Invalid FactoidTxID")
		}
		if bs.PendingECBalanceIncreases[key].Index != index {
			return fmt.Errorf("Invalid Index")
		}
		if bs.PendingECBalanceIncreases[key].NumEC != e.NumEC {
			return fmt.Errorf("Invalid NumEC - %v vs %v. FTxID: %v, ECTxID: %v", bs.PendingECBalanceIncreases[key].NumEC, e.NumEC, fTxID, e.GetHash())
		}
		delete(bs.PendingECBalanceIncreases, key)
		//Already accounted for in the FBlock entry
		//bs.ECBalances[e.ECPubKey.String()] = bs.ECBalances[e.ECPubKey.String()] + e.NumEC
		break
	case entryCreditBlock.ECIDEntryCommit:
		e := v.(*entryCreditBlock.CommitEntry)
		if bs.ECBalances[e.ECPubKey.String()] < uint64(e.Credits) {
			bs.ECBalances[e.ECPubKey.String()] = uint64(e.Credits)
			fmt.Printf("Not enough ECs - %v:%v<%v\n", e.ECPubKey.String(), bs.ECBalances[e.ECPubKey.String()], uint64(e.Credits))
			//return fmt.Errorf("Not enough ECs - %v:%v<%v", e.ECPubKey.String(), bs.ECBalances[e.ECPubKey.String()], uint64(e.Credits))
		}
		bs.ECBalances[e.ECPubKey.String()] = bs.ECBalances[e.ECPubKey.String()] - uint64(e.Credits)
		bs.PushCommit(e.GetEntryHash(), v.Hash())
		break
	case entryCreditBlock.ECIDChainCommit:
		e := v.(*entryCreditBlock.CommitChain)
		if bs.ECBalances[e.ECPubKey.String()] < uint64(e.Credits) {
			bs.ECBalances[e.ECPubKey.String()] = uint64(e.Credits)
			fmt.Printf("Not enough ECs - %v:%v<%v\n", e.ECPubKey.String(), bs.ECBalances[e.ECPubKey.String()], uint64(e.Credits))
			//return fmt.Errorf("Not enough ECs - %v:%v<%v", e.ECPubKey.String(), bs.ECBalances[e.ECPubKey.String()], uint64(e.Credits))
		}
		bs.ECBalances[e.ECPubKey.String()] = bs.ECBalances[e.ECPubKey.String()] - uint64(e.Credits)
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
