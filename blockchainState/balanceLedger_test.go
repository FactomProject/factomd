// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState_test

import (
	"testing"

	. "github.com/FactomProject/factomd/blockchainState"
	"github.com/FactomProject/factomd/testHelper"
)

func TestBalanceLedger(t *testing.T) {
	bs := new(BalanceLedger)
	blocks := testHelper.CreateFullTestBlockSet()
	for _, v := range blocks {
		err := bs.ProcessFBlock(v.FBlock)
		if err != nil {
			t.Errorf("%v", err)
		}
	}
	t.Logf("%v", bs.String())
}
