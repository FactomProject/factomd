// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	"fmt"

	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

//"bytes"
//"encoding/binary"

//"time"

//"github.com/FactomProject/factomd/common/messages"

//. "github.com/FactomProject/factomd/state"

func TestLoadIdentity(t *testing.T) {
	s := testHelper.CreateAndPopulateTestState()
	s.LdbPath = "/home/steven/.factom/m2/custom-" + s.LdbPath
	s.Network = "CUSTOM"
	s.DB = nil
	err := s.InitLevelDB()
	if err != nil {
		t.Error(err)
	}

	h, _ := primitives.HexToHash("8888884e225a7abcf7ddd831f7274ee3df4d8f97b2db7fee42e51a83a22a6426")

	s.AddIdentityFromChainID(h)

	id := s.IdentityControl.GetIdentity(h)
	fmt.Println(id.IdentityChainSync)
	fmt.Println(id)
	var _ = id
}
