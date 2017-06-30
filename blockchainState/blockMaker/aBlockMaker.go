// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker

import (
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/interfaces"
)

func (bm *BlockMaker) BuildABlock() (interfaces.IAdminBlock, error) {
	ab := adminBlock.NewAdminBlock(nil)
	ab.GetHeader().SetPrevBackRefHash(bm.BState.ABlockHeadRefHash)
	ab.GetHeader().SetDBHeight(bm.BState.DBlockHeight + 1)

	for _, v := range bm.ProcessedABEntries {
		err := ab.AddEntry(v)
		if err != nil {
			return nil, err
		}
	}

	return ab, nil
}

func (bm *BlockMaker) BuildABlockWithHeaderExpansionArea(hea []byte) (interfaces.IAdminBlock, error) {
	ab, err := bm.BuildABlock()
	if err != nil {
		return nil, err
	}
	ab.GetHeader().SetHeaderExpansionArea(hea)
	return ab, nil
}

func (bm *BlockMaker) ProcessABEntry(e interfaces.IABEntry) error {
	err := bm.BState.ProcessABlockEntry(e)
	if err != nil {
		bm.PendingABEntries = append(bm.PendingABEntries, e)
	} else {
		bm.ProcessedABEntries = append(bm.ProcessedABEntries, e)
	}
	return nil
}
