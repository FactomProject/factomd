// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func (bs *BlockchainState) ProcessDBlock(dBlock interfaces.IDirectoryBlock) error {
	bs.Init()

	if dBlock.GetDatabaseHeight() != bs.DBlockHeight+1 {
		if dBlock.GetDatabaseHeight() > 0 {
			return fmt.Errorf("Invalid DBlock height - expected %v, got %v", bs.DBlockHeight+1, dBlock.GetDatabaseHeight())
		}
	}
	if bs.DBlockHeadKeyMR.String() != dBlock.GetHeader().GetPrevKeyMR().String() {
		return fmt.Errorf("Invalid DBlock %v previous KeyMR - expected %v, got %v", dBlock.DatabasePrimaryIndex().String(), bs.DBlockHeadKeyMR.String(), dBlock.GetHeader().GetPrevKeyMR().String())
	}
	if bs.DBlockHeadHash.String() != dBlock.GetHeader().GetPrevFullHash().String() {
		return fmt.Errorf("Invalid DBlock %v previous hash - expected %v, got %v", dBlock.DatabasePrimaryIndex().String(), bs.DBlockHeadHash.String(), dBlock.GetHeader().GetPrevFullHash().String())
	}
	if bs.NetworkID != dBlock.GetHeader().GetNetworkID() {
		return fmt.Errorf("Invalid network ID - expected %v, got %v", bs.NetworkID, dBlock.GetHeader().GetNetworkID())
	}
	checkpoint := constants.CheckPoints[dBlock.GetHeader().GetDBHeight()]
	if checkpoint != "" {
		if dBlock.DatabasePrimaryIndex().String() != checkpoint {
			return fmt.Errorf("Invalid KeyMR for checkpoint - expected %v, got %v", checkpoint, dBlock.DatabasePrimaryIndex().String())
		}
	}

	err := bs.CheckDBlockEntries(dBlock)
	if err != nil {
		return err
	}

	bs.DBlockHeadKeyMR = dBlock.DatabasePrimaryIndex().(*primitives.Hash)
	bs.DBlockHeadHash = dBlock.DatabaseSecondaryIndex().(*primitives.Hash)
	bs.DBlockHeight = dBlock.GetDatabaseHeight()
	bs.DBlockTimestamp = dBlock.GetTimestamp().(*primitives.Timestamp)

	dbEntries := dBlock.GetDBEntries()
	for _, v := range dbEntries {
		bs.BlockHeads[v.GetChainID().String()] = v.GetKeyMR().(*primitives.Hash)
	}

	return nil
}

func (bs *BlockchainState) CheckDBlockEntries(dBlock interfaces.IDirectoryBlock) error {
	entries := dBlock.GetDBEntries()
	if len(entries) < 3 {
		return fmt.Errorf("Invalid number of DBentries - %v", len(entries))
	}

	//Checking whether the first 3 entries are the special blocks
	if entries[0].GetChainID().String() != "000000000000000000000000000000000000000000000000000000000000000a" {
		return fmt.Errorf("Invalid ABlock ChainID - %v", entries[0].GetChainID().String())
	}
	if entries[1].GetChainID().String() != "000000000000000000000000000000000000000000000000000000000000000c" {
		return fmt.Errorf("Invalid ECBlock ChainID - %v", entries[1].GetChainID().String())
	}
	if entries[2].GetChainID().String() != "000000000000000000000000000000000000000000000000000000000000000f" {
		return fmt.Errorf("Invalid FBlock ChainID - %v", entries[2].GetChainID().String())
	}

	//Checking whether all entries have unique ChainIDs
	found := map[string]bool{}
	for _, v := range entries {
		if found[v.GetChainID().String()] == true {
			return fmt.Errorf("Duplicate ChainID found in DBlock %v - %v", dBlock.DatabasePrimaryIndex().String(), v.GetChainID().String())
		}
		found[v.GetChainID().String()] = true
	}

	return nil
}
