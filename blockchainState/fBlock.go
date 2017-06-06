// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type PendingECBalanceIncrease struct {
	ECPubKey    string
	FactoidTxID string
	Index       uint64
	NumEC       uint64
}

func (bs *BlockchainState) ProcessFBlock(fBlock interfaces.IFBlock) error {
	bs.Init()

	if bs.FBlockHead.KeyMR.String() != fBlock.GetPrevKeyMR().String() {
		return fmt.Errorf("Invalid FBlock %v previous KeyMR - expected %v, got %v\n", fBlock.DatabasePrimaryIndex().String(), bs.FBlockHead.KeyMR.String(), fBlock.GetPrevKeyMR().String())
	}
	if bs.FBlockHead.Hash.String() != fBlock.GetPrevLedgerKeyMR().String() {
		return fmt.Errorf("Invalid FBlock %v previous hash - expected %v, got %v\n", fBlock.DatabasePrimaryIndex().String(), bs.FBlockHead.Hash.String(), fBlock.GetPrevLedgerKeyMR().String())
	}

	if bs.DBlockHeight != fBlock.GetDatabaseHeight() {
		return fmt.Errorf("Invalid FBlock height - expected %v, got %v", bs.DBlockHeight, fBlock.GetDatabaseHeight())
	}

	bs.FBlockHead.KeyMR = fBlock.DatabasePrimaryIndex().(*primitives.Hash)
	bs.FBlockHead.Hash = fBlock.DatabaseSecondaryIndex().(*primitives.Hash)

	transactions := fBlock.GetTransactions()
	for _, v := range transactions {
		err := bs.ProcessFactoidTransaction(v, fBlock.GetExchRate())
		if err != nil {
			return err
		}
	}
	bs.ExchangeRate = fBlock.GetExchRate()
	return nil
}

func (bs *BlockchainState) ProcessFactoidTransaction(tx interfaces.ITransaction, exchangeRate uint64) error {
	bs.Init()
	ins := tx.GetInputs()
	for _, w := range ins {
		if bs.FBalances[w.GetAddress().String()] < int64(w.GetAmount()) {
			return fmt.Errorf("Not enough factoids")
		}
		bs.FBalances[w.GetAddress().String()] = bs.FBalances[w.GetAddress().String()] - int64(w.GetAmount())
	}
	outs := tx.GetOutputs()
	for _, w := range outs {
		bs.FBalances[w.GetAddress().String()] = bs.FBalances[w.GetAddress().String()] + int64(w.GetAmount())
	}
	ecOut := tx.GetECOutputs()
	for i, w := range ecOut {

		pb := new(PendingECBalanceIncrease)
		pb.ECPubKey = w.GetAddress().String()
		pb.FactoidTxID = tx.GetHash().String()
		pb.Index = uint64(i)
		pb.NumEC = w.GetAmount() / exchangeRate

		if pb.ECPubKey == LookingFor {
			Balances = append(Balances, Balance{Delta: int64(pb.NumEC), TxID: pb.FactoidTxID})
			fmt.Printf("%v\t%v\t%v\t%v\n", bs.DBlockHeight, pb.FactoidTxID, pb.NumEC, bs.ECBalances[w.GetAddress().String()]+int64(pb.NumEC))
			if pb.FactoidTxID == "81cc0fc493395808c85bb6536d9c366f7dd20c4781644929c90954e24c1cc990" {
				panic("end")
			}
		}

		bs.ECBalances[w.GetAddress().String()] = bs.ECBalances[w.GetAddress().String()] + int64(pb.NumEC)

		bs.PendingECBalanceIncreases[fmt.Sprintf("%v:%v", pb.FactoidTxID, pb.Index)] = pb
	}
	return nil
}

func (bs *BlockchainState) CanProcessFactoidTransaction(tx interfaces.ITransaction) bool {
	bs.Init()
	ins := tx.GetInputs()
	for _, w := range ins {
		if bs.FBalances[w.GetAddress().String()] < int64(w.GetAmount()) {
			return false
		}
	}
	return true
}
