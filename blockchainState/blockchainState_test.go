// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState_test

import (
	"testing"

	. "github.com/FactomProject/factomd/blockchainState"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/testHelper"
)

func TestBlockchainState(t *testing.T) {
	bs := new(BlockchainState)
	blocks := testHelper.CreateFullTestBlockSet()
	for _, v := range blocks {
		err := bs.ProcessBlockSet(v.DBlock, v.FBlock, v.ECBlock, []interfaces.IEntryBlock{v.EBlock, v.AnchorEBlock})
		if err != nil {
			t.Errorf("%v", err)
		}
	}
	t.Errorf("%v", bs.String())
}
