// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/util"
	"testing"
)

func TestInit(t *testing.T) {
	state := new(State)

	cfg := util.ReadConfig("")
	cfg.App.DBType = "Map"

	state.Cfg = cfg

	state.Init("")

	if state.Cfg.(*util.FactomdConfig).App.DBType != "Map" {
		t.Error("CFG has been overwritten")
	}
	log.Println(state.String())
}
