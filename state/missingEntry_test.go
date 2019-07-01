// +build all 

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	. "github.com/FactomProject/factomd/state"
)

func TestMissingEntryBlockMarshalUnmarshal(t *testing.T) {
	for i := 0; i < 1000; i++ {
		s := RandomMissingEntryBlock()
		b, err := s.MarshalBinary()
		if err != nil {
			t.Errorf("%v", err)
		}
		s2 := new(MissingEntryBlock)
		rest, err := s2.UnmarshalBinaryData(b)
		if err != nil {
			t.Errorf("%v", err)
		}

		if len(rest) > 0 {
			t.Errorf("Returned too much data")
		}
		if s.IsSameAs(s2) == false {
			t.Errorf("MissingEntryBlocks are not the same")
		}

		if i == 0 {
			err := s2.UnmarshalBinary(b)
			if err != nil {
				t.Errorf("%v", err)
			}
			if s.IsSameAs(s2) == false {
				t.Errorf("MissingEntryBlocks are not the same")
			}
		}
	}
}

func TestMissingEntryMarshalUnmarshal(t *testing.T) {
	for i := 0; i < 1000; i++ {
		s := RandomMissingEntry()
		b, err := s.MarshalBinary()
		if err != nil {
			t.Errorf("%v", err)
		}
		s2 := new(MissingEntry)
		rest, err := s2.UnmarshalBinaryData(b)
		if err != nil {
			t.Errorf("%v", err)
		}
		if len(rest) > 0 {
			t.Errorf("Returned too much data")
		}
		if s.IsSameAs(s2) == false {
			t.Errorf("MissingEntries are not the same")
		}
	}
}
