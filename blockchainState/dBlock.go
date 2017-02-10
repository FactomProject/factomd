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

	bs.DBlockHeadKeyMR = dBlock.DatabasePrimaryIndex().(*primitives.Hash)
	bs.DBlockHeadHash = dBlock.DatabaseSecondaryIndex().(*primitives.Hash)
	bs.DBlockHeight = dBlock.GetDatabaseHeight()

	dbEntries := dBlock.GetDBEntries()
	for _, v := range dbEntries {
		bs.BlockHeads[v.GetChainID().String()] = v.GetKeyMR().(*primitives.Hash)
	}

	return nil
}
