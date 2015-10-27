// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"github.com/FactomProject/factomd/log"
	"testing"
)

func TestInit(t *testing.T) {
	state := new(State)
	state.Init()
	log.Println(state.String())
}
