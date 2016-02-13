// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/util"
	"testing"
)

var _ = log.Print
var _ = util.ReadConfig

var state *State

func GetState() *State {
	if state == nil {
		state = new(State)
		state.Init("")
	}
	return state
}

func TestInit(t *testing.T) {
	GetState()
}
