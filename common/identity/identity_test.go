// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestIdentityMarshalUnmarshal(t *testing.T) {
	for i := 0; i < 1000; i++ {
		id := RandomIdentity()
		h, err := id.MarshalBinary()
		if err != nil {
			t.Errorf("%v", err)
		}
		id2 := new(Identity)
		err = id2.UnmarshalBinary(h)
		if err != nil {
			t.Errorf("%v", err)
		}
		if id.IsSameAs(id2) == false {
			t.Errorf("Identities are not the same")
		}
	}
}

func TestIsComplete(t *testing.T) {
	i := NewIdentity()
	if complete, _ := i.IsComplete(); complete {
		t.Errorf("(1) Identity returns as complete, but it is incomplete")
	}

	i = RandomIdentity()
	if complete, _ := i.IsComplete(); !complete {
		t.Errorf("(2) Identity returns as incomplete, but it is complete")
	}

	i = RandomIdentity()
	i.ManagementChainID = nil
	if complete, _ := i.IsComplete(); complete {
		t.Errorf("(3) Identity returns as complete, but it is incomplete")
	}

	i = RandomIdentity()
	i.IdentityChainID = primitives.NewZeroHash()
	if complete, _ := i.IsComplete(); complete {
		t.Errorf("(4) Identity returns as complete, but it is incomplete")
	}

	i = RandomIdentity()
	i.ManagementChainID = nil
	if complete, _ := i.IsComplete(); complete {
		t.Errorf("(5) Identity returns as complete, but it is incomplete")
	}

	i = RandomIdentity()
	i.Keys[0] = nil
	if complete, _ := i.IsComplete(); complete {
		t.Errorf("(6) Identity returns as complete, but it is incomplete")
	}

	i = RandomIdentity()
	i.AnchorKeys = nil
	if complete, _ := i.IsComplete(); complete {
		t.Errorf("(7) Identity returns as complete, but it is incomplete")
	}

	i = RandomIdentity()
	i.MatryoshkaHash = primitives.NewZeroHash()
	if complete, _ := i.IsComplete(); complete {
		t.Errorf("(8) Identity returns as complete, but it is incomplete")
	}
}
