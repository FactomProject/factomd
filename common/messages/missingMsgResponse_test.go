// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

func TestBadUnmarshal(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	m := new(MissingMsgResponse)
	err := m.UnmarshalBinary(nil)
	if err == nil {
		t.Error("Should error")
	}
}

func TestMissingMessageResponseMarshaling(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	s := testHelper.CreateEmptyTestState()

	for i := 0; i < 1; i++ {
		b := new(Bounce)
		b.Timestamp = primitives.NewTimestampNow()
		m := NewMissingMsgResponse(s, b, newSignedAck())
		m.GetHash()
		m.GetMsgHash()
		d, err := m.MarshalBinary()
		if err != nil {
			t.Error(err)
		}

		m2 := new(MissingMsgResponse)
		nd, err := m2.UnmarshalBinaryData(d)
		if err != nil {
			t.Error(err)
		}

		if len(nd) > 0 {
			t.Errorf("Should not have leftover bytes, found %d", len(nd))
		}

		m1 := m.(*MissingMsgResponse)
		if !m1.IsSameAs(m2) {
			t.Error("Unmarshal gave back a different message")
		}
	}

}
