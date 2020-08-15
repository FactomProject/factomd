// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	//"github.com/PaulSnow/factom2d/common/constants"
	//"github.com/PaulSnow/factom2d/common/primitives"
	"github.com/PaulSnow/factom2d/common/messages/electionMsgs"
	"github.com/PaulSnow/factom2d/state"
	"github.com/PaulSnow/factom2d/util"
)

var _ = util.ReadConfig

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
