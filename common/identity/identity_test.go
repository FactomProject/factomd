// +build all 

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity_test

import (
	"testing"

	"bytes"

	. "github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestMarshalEntryBlockSync(t *testing.T) {
	e := NewEntryBlockSync()
	e.Target = *NewEntryBlockMarker()
	e.Current = *NewEntryBlockMarker()
	e.BlocksToBeParsed = []EntryBlockMarker{*NewEntryBlockMarker(), *NewEntryBlockMarker(), *NewEntryBlockMarker()}

	data, err := e.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	e2 := NewEntryBlockSync()
	data, err = e2.UnmarshalBinaryData(data)
	if err != nil {
		t.Error(err)
	}
	if len(data) != 0 {
		t.Errorf("%d bytes left after unmarshal", len(data))
	}

	if len(e2.BlocksToBeParsed) != 3 {
		t.Errorf("Should be 3 in blocks to be parsed, found %d", len(e2.BlocksToBeParsed))
	}
}

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

func TestIdentityClone(t *testing.T) {
	for i := 0; i < 1000; i++ {
		id := RandomIdentity()
		id2 := id.Clone()
		if id.IsSameAs(id2) == false {
			t.Errorf("Identities are not the same")
		}

		// Check their marshalled values
		d1, err := id.MarshalBinary()
		if err != nil {
			t.Errorf("%v", err)
		}
		d2, err := id2.MarshalBinary()
		if err != nil {
			t.Errorf("%v", err)
		}

		if bytes.Compare(d1, d2) != 0 {
			t.Errorf("Identities are not the same when marshalled")
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
