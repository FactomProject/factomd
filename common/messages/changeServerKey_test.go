// +build all 

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestUnmarshalNilChangeServerKeyMsg(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(ChangeServerKeyMsg)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalChangeServerKey(t *testing.T) {
	addserv := newChangeServerKey()

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

	if addserv2.Type() != constants.CHANGESERVER_KEY_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	if addserv.IsSameAs(addserv2.(*ChangeServerKeyMsg)) != true {
		t.Errorf("AddServer messages are not identical")
	}
}

func TestMarshalUnmarshalSignedChangeServerKey(t *testing.T) {
	addserv := newSignedChangeServerKey()

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

	valid, err := addserv.VerifySignature()
	if err != nil {
		t.Error(err)
	}
	if valid == false {
		t.Error("Signature is not valid")
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

	if addserv2.Type() != constants.CHANGESERVER_KEY_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	if addserv.IsSameAs(addserv2.(*ChangeServerKeyMsg)) != true {
		t.Errorf("AddServer messages are not identical")
	}

	valid, err = addserv2.(*ChangeServerKeyMsg).VerifySignature()
	if err != nil {
		t.Error(err)
	}
	if valid == false {
		t.Error("Signature is not valid")
	}
}

func newChangeServerKey() *ChangeServerKeyMsg {
	addserv := new(ChangeServerKeyMsg)
	addserv.Timestamp = primitives.NewTimestampNow()
	addserv.IdentityChainID = primitives.Sha([]byte("FNode0"))
	addserv.AdminBlockChange = 0
	addserv.KeyPriority = 0
	addserv.KeyType = 0
	addserv.Key = primitives.Sha([]byte("A_Key"))
	return addserv
}

func newSignedChangeServerKey() *ChangeServerKeyMsg {
	addserv := newChangeServerKey()

	key, err := primitives.NewPrivateKeyFromHex("07c0d52cb74f4ca3106d80c4a70488426886bccc6ebc10c6bafb37bf8a65f4c38cee85c62a9e48039d4ac294da97943c2001be1539809ea5f54721f0c5477a0a")
	if err != nil {
		panic(err)
	}
	err = addserv.Sign(key)
	if err != nil {
		panic(err)
	}

	return addserv
}

// TODO: Add test for signed messages (See ack_test.go)
