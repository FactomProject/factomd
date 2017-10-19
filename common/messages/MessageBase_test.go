// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

func TestMessageBase(t *testing.T) {
	for i := 0; i < 1000; i++ {
		mb := new(MessageBase)

		h := mb.GetLeaderChainID()
		if h.IsZero() == false {
			t.Errorf("Hash is not zero")
		}

		hash := primitives.RandomHash()
		mb.SetLeaderChainID(hash)

		h = mb.GetLeaderChainID()
		if h.IsSameAs(hash) == false {
			t.Errorf("Hashes are not equal")
		}

		mb.SetFullMsgHash(hash)
		fmh := mb.GetFullMsgHash()
		if fmh.IsSameAs(hash) == false {
			t.Errorf("FullMsgHashes are not equal")
		}

		b := random.RandByteSlice()
		mb.SetVMHash(b)
		if primitives.AreBytesEqual(mb.GetVMHash(), b) == false {
			t.Errorf("VMHashes are not equal")
		}

		mb.SetMinute(byte(i % 256))
		if mb.GetMinute() != byte(i%256) {
			t.Errorf("Minutes are not equal")
		}
	}
}

func TestOrigin(t *testing.T) {
	eom := newEOM()
	eom.SetOrigin(123)
	if eom.GetOrigin() != 123 {
		t.Error("SetOrigin/GetOrigin mismatch")
	}
	eom.SetOrigin(321)
	if eom.GetOrigin() != 321 {
		t.Error("SetOrigin/GetOrigin mismatch")
	}
}

func TestNetworkOrigin(t *testing.T) {
	eom := newEOM()
	eom.SetNetworkOrigin("FNode00")
	if eom.GetNetworkOrigin() != "FNode00" {
		t.Error("SetNetworkOrigin/GetNetworkOrigin mismatch")
	}
	eom.SetNetworkOrigin("FNode123")
	if eom.GetNetworkOrigin() != "FNode123" {
		t.Error("SetNetworkOrigin/GetNetworkOrigin mismatch")
	}
}
