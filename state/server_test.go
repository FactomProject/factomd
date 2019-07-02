// +build all

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	. "github.com/FactomProject/factomd/state"
)

func TestServerMarshalUnmarshal(t *testing.T) {
	for i := 0; i < 1000; i++ {
		s := RandomServer()
		b, err := s.MarshalBinary()
		if err != nil {
			t.Errorf("%v", err)
		}
		s2 := new(Server)
		rest, err := s2.UnmarshalBinaryData(b)
		if err != nil {
			t.Errorf("%v", err)
		}
		if len(rest) > 0 {
			t.Errorf("Returned too much data")
		}
		if s.IsSameAs(s2) == false {
			t.Errorf("Servers are not the same")
		}
	}
}

func TestServer(t *testing.T) {
	rbool := func() bool {
		if random.RandIntBetween(0, 2) == 1 {
			return true
		}
		return false
	}

	for i := 0; i < 5; i++ {
		c := primitives.RandomHash()
		n := random.RandomString()
		o := rbool()
		r := primitives.RandomHash()

		f := new(Server)
		f.ChainID = c
		f.Name = n
		f.Online = o
		f.Replace = r

		sf := new(Server)
		sf.ChainID = c
		sf.Name = n
		sf.Online = o
		sf.Replace = r
		if f.String() != sf.String() {
			t.Error("String should be same")
		}

		if !f.GetChainID().IsSameAs(c) {
			t.Error("Should be same chain")
		}

		if !(f.GetName() == n) {
			t.Error("Should be same name")
		}

		if f.IsOnline() != o {
			t.Error("Should be same online")
		}

		f.SetOnline(!o)
		if f.IsOnline() == o {
			t.Error("Should be different online")
		}

		if !f.LeaderToReplace().IsSameAs(r) {
			t.Error("Should be same replace")
		}

		f.SetReplace(primitives.RandomHash())
		if f.LeaderToReplace().IsSameAs(r) {
			t.Error("Should be different replace")
		}
	}
}
