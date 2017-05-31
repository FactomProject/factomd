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
	bs := NewBSLocalNet()
	blocks := testHelper.CreateFullTestBlockSet()
	for _, v := range blocks {
		entries := []interfaces.IEBEntry{}
		for _, w := range v.Entries {
			entries = append(entries, w)
		}
		err := bs.ProcessBlockSet(v.DBlock, v.ABlock, v.FBlock, v.ECBlock, []interfaces.IEntryBlock{v.EBlock, v.AnchorEBlock}, entries)
		if err != nil {
			t.Errorf("%v", err)
		}
	}
	t.Logf("%v", bs.String())
}
