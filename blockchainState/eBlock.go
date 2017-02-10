// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
)

func (bs *BlockchainState) ProcessEBlocks(eBlocks []interfaces.IEntryBlock) error {
	bs.Init()
	for _, v := range eBlocks {
		err := bs.ProcessEBlock(v)
		if err != nil {
			return err
		}
	}
	return bs.ClearExpiredCommits()
}

func (bs *BlockchainState) ProcessEBlock(eBlock interfaces.IEntryBlock) error {
	bs.Init()
	eHashes := eBlock.GetEntryHashes()
	for _, v := range eHashes {
		err := bs.ProcessEntryHash(v, eBlock.GetHash())
		if err != nil {
			return err
		}
	}
	return nil
}

func (bs *BlockchainState) ProcessEntryHash(v, block interfaces.IHash) error {
	bs.Init()
	if v.IsMinuteMarker() {
		return nil
	}
	TotalEntries++
	if bs.HasFreeCommit(v) == true {

	} else {
		return fmt.Errorf("Non-committed entry found in an eBlock - %v, %v, %v, %v\n", bs.DBlockHeadKeyMR.String(), bs.DBlockHeight, block.String(), v.String())
		//MES.NewMissing(v.String(), bs.DBlockHeadKeyMR.String(), bs.DBlockHeight)
	}
	err := bs.PopCommit(v)
	if err != nil {
		return err
		//fmt.Printf("Error - %v\n", err)
		//panic("")
	}
	return nil
}
