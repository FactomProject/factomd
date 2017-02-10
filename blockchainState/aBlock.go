// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func (bs *BlockchainState) ProcessABlock(aBlock interfaces.IAdminBlock) error {
	bs.Init()

	if bs.ABlockHeadRefHash.String() != aBlock.GetHeader().GetPrevBackRefHash().String() {
		fmt.Printf("Invalid ABlock %v previous KeyMR - expected %v, got %v\n", aBlock.GetHash(), bs.ABlockHeadRefHash.String(), aBlock.GetHeader().GetPrevBackRefHash().String())
	}
	bs.ABlockHeadRefHash = aBlock.DatabaseSecondaryIndex().(*primitives.Hash)

	return nil
}
