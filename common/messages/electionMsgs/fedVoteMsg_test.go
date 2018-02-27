// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestUnmarshalfolunteerAudit_test(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	messages.General = new(msgsupport.GeneralFactory)
	primitives.General = messages.General

	a := new(FedVoteMsg)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Error("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Error("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalAck(t *testing.T) {
	test := func(va *FedVoteVolunteerMsg, num string) {
		_, err := va.JSONString()
		if err != nil {
			t.Error(err)
		}
		hex, err := va.MarshalBinary()
		if err != nil {
			t.Error(err)
		}

		va2, err := msgsupport.UnmarshalMessage(hex)
		if err != nil {
			t.Error(err)
		}
		_, err = va2.JSONString()
		if err != nil {
			t.Error(err)
		}

		if va2.Type() != constants.VOLUNTEERAUDIT {
			t.Error(num + " Invalid message type unmarshalled")
		}

		if va.IsSameAs(va2) == false {
			t.Error(num + " Acks are not the same")
		}
	}

	va := new(FedVoteVolunteerMsg)
	va.Minute = 5
	va.Name = "bob"
	va.DBHeight = 10
	va.ServerID = primitives.Sha([]byte("leader"))
	//va.Weight = primitives.Sha([]byte("Weight"))
	va.ServerIdx = 3
	va.Missing = new(messages.EOM)
	eom := va.Missing.(*messages.EOM)
	eom.ChainID = primitives.NewHash([]byte("id"))
	eom.LeaderChainID = primitives.NewHash([]byte("leader"))
	eom.Timestamp = primitives.NewTimestampNow()

	va.Ack = new(messages.Ack)
	ack := va.Ack.(*messages.Ack)
	ack.Timestamp = primitives.NewTimestampNow()
	ack.LeaderChainID = primitives.NewHash([]byte("leader"))
	ack.MessageHash = primitives.NewHash([]byte("msg"))
	ack.SerialHash = primitives.NewHash([]byte("serial"))
	va.TS = primitives.NewTimestampNow()

	va.FedID = primitives.RandomHash()
	test(va, "1")
}
