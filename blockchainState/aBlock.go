// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"fmt"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func (bs *BlockchainState) ProcessABlock(aBlock interfaces.IAdminBlock, dBlock interfaces.IDirectoryBlock) error {
	bs.Init()

	if bs.ABlockHeadRefHash.String() != aBlock.GetHeader().GetPrevBackRefHash().String() {
		return fmt.Errorf("Invalid ABlock %v previous KeyMR - expected %v, got %v\n", aBlock.GetHash(), bs.ABlockHeadRefHash.String(), aBlock.GetHeader().GetPrevBackRefHash().String())
	}
	bs.ABlockHeadRefHash = aBlock.DatabaseSecondaryIndex().(*primitives.Hash)

	if bs.DBlockHeight != aBlock.GetDatabaseHeight() {
		return fmt.Errorf("Invalid ABlock height - expected %v, got %v", bs.DBlockHeight, aBlock.GetDatabaseHeight())
	}

	err := CheckABlockMinuteNumbers(aBlock)
	if err != nil {
		return err
	}

	err = bs.CheckDBSignatureEntries(aBlock, dBlock)
	if err != nil {
		return err
	}

	return nil
}

func CheckABlockMinuteNumbers(aBlock interfaces.IAdminBlock) error {
	//Check whether MinuteNumbers are increasing
	entries := aBlock.GetABEntries()

	var lastMinute uint8 = 0
	for i, v := range entries {
		if v.Type() == constants.TYPE_MINUTE_NUM {
			minute := v.(*adminBlock.EndOfMinuteEntry).MinuteNumber
			if minute < 1 || minute > 10 {
				return fmt.Errorf("ABlock Invalid minute number at position %v", i)
			}
			if minute <= lastMinute {
				return fmt.Errorf("ABlock Invalid minute number at position %v", i)
			}
			lastMinute = minute
		}
	}

	return nil
}

func (bs *BlockchainState) CheckDBSignatureEntries(aBlock interfaces.IAdminBlock, dBlock interfaces.IDirectoryBlock) error {
	if dBlock.GetDatabaseHeight() == 0 {
		return nil
	}

	entries := aBlock.GetABEntries()

	foundSigs := map[string]string{}

	for _, v := range entries {
		if v.Type() == constants.TYPE_DB_SIGNATURE {
			dbs := v.(*adminBlock.DBSignatureEntry)
			if foundSigs[dbs.IdentityAdminChainID.String()] != "" {
				return fmt.Errorf("Found duplicate entry for ChainID %v", dbs.IdentityAdminChainID.String())
			}
			foundSigs[dbs.IdentityAdminChainID.String()] = "ok"
			pub := dbs.PrevDBSig.Pub
			sig := dbs.PrevDBSig.GetSignature()

			if bs.IdentityChains[dbs.IdentityAdminChainID.String()] != pub.String() {
				return fmt.Errorf("Invalid Public Key in DBSignatureEntry %v - expected %v, got %v", v.Hash().String(), bs.IdentityChains[dbs.IdentityAdminChainID.String()], pub.String())
			}

			if pub.Verify(dBlock.GetHeader().GetPrevKeyMR().Bytes(), sig) == false {
				return fmt.Errorf("Invalid signature in DBSignatureEntry %v", v.Hash().String())
			}
		}
	}
	if len(foundSigs) != len(bs.IdentityChains) {
		return fmt.Errorf("Invalid number of DBSignatureEntries found in aBlock %v", aBlock.DatabasePrimaryIndex().String())
	}
	return nil
}
