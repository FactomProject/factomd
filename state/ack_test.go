// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	"github.com/FactomProject/factomd/testHelper"
	//. "github.com/FactomProject/factomd/state"
)

func TestIsStateFullySynced(t *testing.T) {
	s1_good := testHelper.CreateAndPopulateTestState()
	t.Log("IsStateFullySynced():", s1_good.IsStateFullySynced())
	if !s1_good.IsStateFullySynced() {
		t.Error("test state is show to be not fully synced")
	}

	// we can't test the negative here because when we set the bad DBHeight the
	// state.ValidatorLoop() will panic before we call IsStateFullySynced() the
	// state.ValidatorLoop()
	//	s2_bad := testHelper.CreateAndPopulateTestState()
	//	s2_bad.ProcessLists.DBHeightBase = s2_bad.ProcessLists.LastList().DBHeight+10
	//	fmt.Println("DEBUG: IsStateFullySynced:", s2_bad.IsStateFullySynced())

}
