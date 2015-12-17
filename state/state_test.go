// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"github.com/FactomProject/factomd/log"
	. "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
	"testing"
)

var _ = log.Print
var _ = util.ReadConfig

func TestInit(t *testing.T) {
	state := new(State)
	var _ = state
}
