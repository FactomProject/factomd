// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker

import (
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
)

type BlockSet struct {
	DBlock  interfaces.IDirectoryBlock
	ABlock  interfaces.IAdminBlock
	ECBlock interfaces.IEntryCreditBlock
	FBlock  interfaces.IFBlock
	EBlocks []interfaces.IEntryBlock
	Entries []interfaces.IEBEntry
}

func (bm *BlockMaker) BuildBlockSet() (*BlockSet, error) {
	aBlock, err := bm.BuildABlock()
	if err != nil {
		return nil, err
	}
	ecBlock, err := bm.BuildECBlock()
	if err != nil {
		return nil, err
	}
	fBlock, err := bm.BuildFBlock()
	if err != nil {
		return nil, err
	}
	eBlocks, entries, err := bm.BuildEBlocks()
	if err != nil {
		return nil, err
	}

	bSet := new(BlockSet)
	bSet.ABlock = aBlock
	bSet.ECBlock = ecBlock
	bSet.FBlock = fBlock
	bSet.EBlocks = eBlocks
	bSet.Entries = entries

	dBlock := directoryBlock.NewDirectoryBlock(nil)
	dBlock.GetHeader().SetPrevKeyMR(bm.BState.DBlockHead.KeyMR)
	dBlock.GetHeader().SetPrevFullHash(bm.BState.DBlockHead.Hash)
	dBlock.GetHeader().SetDBHeight(bm.BState.DBlockHeight + 1)

	err = dBlock.SetABlockHash(aBlock)
	if err != nil {
		return nil, err
	}
	err = dBlock.SetFBlockHash(fBlock)
	if err != nil {
		return nil, err
	}
	err = dBlock.SetECBlockHash(ecBlock)
	if err != nil {
		return nil, err
	}

	for _, v := range eBlocks {
		err = dBlock.AddEntry(v.GetChainID(), v.GetHash())
		if err != nil {
			return nil, err
		}
	}

	bSet.DBlock = dBlock

	return bSet, nil
}
