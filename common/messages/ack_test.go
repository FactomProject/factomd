// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"testing"
)

func Test(t *testing.T) {
	ack := new(Ack)
	t.Log(ack.String())
}

func TestAck(t *testing.T) {
	ack := new(Ack)
	//ack.Timestamp.SetTimeNow()
	hash, err := primitives.NewShaHashFromStr("cbd3d09db6defdc25dfc7d57f3479b339a077183cd67022e6d1ef6c041522b40")
	if err != nil {
		t.Error(err)
	}
	ack.OriginalHash = hash
	hex, err := MarshalAck(ack)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	ack2, err := UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str := ack2.String()
	t.Logf("str - %v", str)

	if ack2.Type() != constants.ACK_MSG {
		t.Error("Invalid message type unmarshalled")
	}
}
