// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestUnmarshalNilAddServerMsg(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(AddServerMsg)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalAddServer(t *testing.T) {
	addserv := newAddServer()

	str, err := addserv.JSONString()
	if err != nil {
		t.Error(err)
	}
	t.Logf("str1 - %v", str)
	hex, err := addserv.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	addserv2, err := msgsupport.UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str, err = addserv2.JSONString()
	if err != nil {
		t.Error(err)
	}
	t.Logf("str2 - %v", str)

	if addserv2.Type() != constants.ADDSERVER_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	if addserv.IsSameAs(addserv2.(*AddServerMsg)) != true {
		t.Errorf("AddServer messages are not identical")
	}
}

func TestMarshalUnmarshalSignedAddServer(t *testing.T) {
	addserv := newSignedAddServer()

	str, err := addserv.JSONString()
	if err != nil {
		t.Error(err)
	}
	t.Logf("str1 - %v", str)
	hex, err := addserv.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	valid, err := addserv.VerifySignatures()
	if err != nil {
		t.Error(err)
	}
	if len(valid) != len(addserv.GetSignatures()) {
		t.Error("Some signatures not valid")
	}

	addserv2, err := msgsupport.UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str, err = addserv2.JSONString()
	if err != nil {
		t.Error(err)
	}
	t.Logf("str2 - %v", str)

	if addserv2.Type() != constants.ADDSERVER_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	if addserv.IsSameAs(addserv2.(*AddServerMsg)) != true {
		t.Errorf("AddServer messages are not identical")
	}

	valid, err = addserv2.(*AddServerMsg).VerifySignatures()
	if err != nil {
		t.Error(err)
	}
	if len(valid) != len(addserv2.(*AddServerMsg).GetSignatures()) {
		t.Error("Signatures 2 are not valid")
	}
}

func newAddServer() *AddServerMsg {
	addserv := new(AddServerMsg)
	addserv.Timestamp = primitives.NewTimestampNow()
	addserv.ServerChainID = primitives.Sha([]byte("FNode0"))
	addserv.ServerType = 0
	addserv.Signatures = factoid.NewFullSignatureBlock()
	return addserv
}

func newSignedAddServer() *AddServerMsg {
	addserv := newAddServer()

	key, err := primitives.NewPrivateKeyFromHex("07c0d52cb74f4ca3106d80c4a70488426886bccc6ebc10c6bafb37bf8a65f4c38cee85c62a9e48039d4ac294da97943c2001be1539809ea5f54721f0c5477a0a")
	if err != nil {
		panic(err)
	}
	err = addserv.AddSignature(key)
	if err != nil {
		panic(err)
	}

	return addserv
}

// TODO: Add test for signed messages (See ack_test.go)
