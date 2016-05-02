// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/messages"
)

func TestMarshalUnmarshalMissingAck(t *testing.T) {
	msg := newMissingAck()

	hex, err := msg.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	msg2, err := UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str := msg2.String()
	t.Logf("str - %v", str)

	if msg2.Type() != constants.MISSING_ACK_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	hex2, err := msg2.(*MissingAck).MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if len(hex) != len(hex2) {
		t.Error("Hexes aren't of identical length")
	}
	for i := range hex {
		if hex[i] != hex2[i] {
			t.Error("Hexes do not match")
		}
	}

	if msg.IsSameAs(msg2.(*MissingAck)) != true {
		t.Errorf("MissingAck messages are not identical")
	}
}

func newMissingAck() *MissingAck {
	msg := new(MissingAck)
	msg.Timestamp.SetTimeNow()

	return msg
}
