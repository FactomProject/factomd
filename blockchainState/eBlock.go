// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
)

func (bs *BlockchainState) ProcessEBlocks(eBlocks []interfaces.IEntryBlock, entries []interfaces.IEBEntry) error {
	bs.Init()
	entryMap := map[string]interfaces.IEBEntry{}
	for _, v := range entries {
		entryMap[v.GetHash().String()] = v
	}
	for _, v := range eBlocks {
		err := bs.ProcessEBlock(v, entryMap)
		if err != nil {
			return err
		}
	}
	err := bs.IdentityManager.ProcessOldEntries()
	if err != nil {
		return err
	}
	return bs.ClearExpiredCommits()
}

func (bs *BlockchainState) ProcessEBlock(eBlock interfaces.IEntryBlock, entryMap map[string]interfaces.IEBEntry) error {
	bs.Init()

	err := CheckEBlockMinuteNumbers(eBlock)
	if err != nil {
		return err
	}

	eHashes := eBlock.GetEntryHashes()
	for _, v := range eHashes {
		err := bs.ProcessEntryHash(v, eBlock.GetHash())
		if err != nil {
			return err
		}
	}

	if IsSpecialBlock(eBlock.GetChainID()) {
		err = bs.ProcessSpecialBlock(eBlock, entryMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func IsSpecialBlock(chainID interfaces.IHash) bool {
	switch chainID.String() {
	//Identity chain
	case "888888001750ede0eff4b05f0c3f557890b256450cabbb84cada937f9c258327":
		return true
	}
	if chainID.String()[:6] == "888888" {
		return true
	}
	return false
}

func (bs *BlockchainState) ProcessSpecialBlock(eBlock interfaces.IEntryBlock, entryMap map[string]interfaces.IEBEntry) error {
	if IsSpecialBlock(eBlock.GetChainID()) == false {
		return fmt.Errorf("Non-special block passed to ProcessSpecialBlock - %v", eBlock.GetHash().String())
	}
	if eBlock.GetChainID().String()[:6] == "888888" {
		//Identity Chain
		for _, v := range eBlock.GetEntryHashes() {
			if v.IsMinuteMarker() {
				continue
			}
			entry := entryMap[v.String()]
			fmt.Printf("Processing entry %v\n", entry.String())

			err := bs.IdentityManager.ProcessIdentityEntry(entry, bs.DBlockHeight, bs.DBlockTimestamp, true)
			if err != nil {
				fmt.Printf("Err - %v\n", err)
				continue
				return err
			}
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

func CheckEBlockMinuteNumbers(eBlock interfaces.IEntryBlock) error {
	//Check whether MinuteNumbers are increasing
	entries := eBlock.GetEntryHashes()

	var lastMinute byte = 0
	for i, v := range entries {
		if v.IsMinuteMarker() {
			minute := v.ToMinute()
			if minute < 1 || minute > 10 {
				return fmt.Errorf("EBlock Invalid minute number at position %v", i)
			}
			if minute <= lastMinute {
				return fmt.Errorf("EBlock Invalid minute number at position %v", i)
			}
			lastMinute = minute
		}
	}

	return nil
}
