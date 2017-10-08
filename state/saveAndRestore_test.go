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

func TestPushPopBalanceMap(t *testing.T) {
	for i := 0; i < 1000; i++ {
		m := map[[32]byte]int64{}
		l := random.RandIntBetween(0, 1000)
		for j := 0; j < l; j++ {
			h := primitives.RandomHash()
			m[h.Fixed()] = random.RandInt64()
		}
		b := primitives.NewBuffer(nil)

		err := PushBalanceMap(b, m)
		if err != nil {
			t.Errorf("%v", err)
		}

		m2, err := PopBalanceMap(b)
		if err != nil {
			t.Errorf("%v", err)
		}
		if len(m) != len(m2) {
			t.Errorf("Map lengths are not identical - %v vs %v", len(m), len(m2))
		}

		for k := range m {
			if m[k] != m2[k] {
				t.Errorf("Invalid balances - %v vs %v", m[k], m2[k])
			}
		}
	}
}

func TestSaveRestore(t *testing.T) {

	ss := new(SaveState)
	ss.LeaderTimestamp = primitives.NewTimestampNow()
	ss.Init()
	snil := (*SaveState)(nil)
	snil2 := snil
	if !snil.IsSameAs(snil2) {
		t.Error("Should be able to compare nils")
	}
	if snil.IsSameAs(ss) {
		t.Error("Should be able to compare a nil with a state")
	}
	if ss.IsSameAs(snil) {
		t.Error("Should be able to compare a state with a nil")
	}
	if !ss.IsSameAs(ss) {
		t.Error("One should be the same as one's self")
	}
}
