// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

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
		err := bs.ProcessFactoidTransaction(v)
		if err != nil {
			return err
		}
	}
	bs.ExchangeRate = fBlock.GetExchRate()
	return nil
}

func (bs *BlockchainState) ProcessFactoidTransaction(tx interfaces.ITransaction) error {
	bs.Init()
	ins := tx.GetInputs()
	for _, w := range ins {
		if bs.FBalances[w.GetAddress().String()] < w.GetAmount() {
			return fmt.Errorf("Not enough factoids")
		}
		bs.FBalances[w.GetAddress().String()] = bs.FBalances[w.GetAddress().String()] - w.GetAmount()
	}
	outs := tx.GetOutputs()
	for _, w := range outs {
		bs.FBalances[w.GetAddress().String()] = bs.FBalances[w.GetAddress().String()] + w.GetAmount()
	}
	ecOut := tx.GetECOutputs()
	for _, w := range ecOut {
		bs.ECBalances[w.GetAddress().String()] = bs.ECBalances[w.GetAddress().String()] + w.GetAmount()
	}
	return nil
}
