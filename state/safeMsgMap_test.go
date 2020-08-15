// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	"github.com/PaulSnow/factom2d/common/messages"
	"github.com/PaulSnow/factom2d/common/primitives"
	. "github.com/PaulSnow/factom2d/state"
)

func TestSafeMsgMag(t *testing.T) {
	state := new(State)
	m := NewSafeMsgMap("test", state)
	addAndTest := func() {
		for i := 0; i < 10; i++ {
			hash := primitives.RandomHash()
			m.Put(hash.Fixed(), new(messages.Ack))

			msg := m.Get(hash.Fixed())
			if msg == nil {
				t.Error("Should not be nil")
			}

			if m.Len() != i+1 {
				t.Errorf("Length should be %d, found %d", i, m.Len())
			}

		}
	}

	addAndTest()

	if m.Len() != 10 {
		t.Errorf("Length should be 10, found %d", m.Len())
	}

	m.Reset()
	if m.Len() != 0 {
		t.Error("Reset should clear map")
	}

	addAndTest()
}
