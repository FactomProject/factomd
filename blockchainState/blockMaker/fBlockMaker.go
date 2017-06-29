// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker

import (
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
)

func (bm *BlockMaker) BuildFBlock() (interfaces.IFBlock, error) {
	fb := factoid.NewFBlock(nil)

	fb.SetDBHeight(bm.BState.DBlockHeight + 1)
	fb.SetPrevKeyMR(bm.BState.FBlockHead.KeyMR)
	fb.SetPrevLedgerKeyMR(bm.BState.FBlockHead.Hash)
	fb.SetExchRate(bm.BState.ExchangeRate)

	for i, v := range bm.ProcessedFBEntries {
		if i == 0 {
			err := fb.AddCoinbase(v)
			if err != nil {
				return nil, err
			}
		} else {
			err := fb.AddTransaction(v)
			if err != nil {
				return nil, err
			}
		}
	}

	return fb, nil
}

func (bm *BlockMaker) ProcessFactoidTransaction(e interfaces.ITransaction) error {
	if bm.BState.ProcessFactoidTransaction(e, bm.BState.ExchangeRate) != nil {
		bm.PendingFBEntries = append(bm.PendingFBEntries, e)
	} else {
		bm.ProcessedFBEntries = append(bm.ProcessedFBEntries, e)
	}
	return nil
}
