// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	//"github.com/FactomProject/factomd/common/constants"
	//"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/state"
)

func TestDisplay(t *testing.T) {
	s := new(state.State)
	s.EFactory = new(electionMsgs.ElectionsFactory)
	s.LoadConfig("", "LOCAL")
	s.NodeMode = "SERVER"
	s.DBType = "Map"
	s.Init()

	s.LeaderPL = s.ProcessLists.Get(s.LLeaderHeight)
	if s.CurrentMinute > 9 {
		s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(9, s.IdentityChainID)
	} else {
		s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)
	}

	s.ControlPanelDataRequest = false
	err := s.CopyStateToControlPanel()
	if err != nil {
		t.Error("CopyState failed")
	}
	s.ControlPanelDataRequest = true
	err = s.CopyStateToControlPanel()
	if err != nil {
		t.Error("CopyState failed when requested")
	}

}
