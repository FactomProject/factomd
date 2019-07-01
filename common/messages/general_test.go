// +build all 

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

func TestUnmarshalNil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	_, _, err := UnmarshalMessageData(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	_, _, err = UnmarshalMessageData([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	_, _, err = UnmarshalMessageData([]byte{0xFF})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestUnmarshalMsgTypes(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	s := testHelper.CreateEmptyTestState()

	err := testUnmarshalMsg(NewMissingMsg(s, 0, 0, 0))
	if err != nil {
		t.Error(err)
	}

	b := new(Bounce)
	b.Stamps = []interfaces.Timestamp{}
	b.Timestamp = primitives.NewTimestampNow()
	err = testUnmarshalMsg(b)
	if err != nil {
		t.Error(err)
	}

	br := new(BounceReply)
	br.Stamps = []interfaces.Timestamp{}
	br.Timestamp = primitives.NewTimestampNow()
	err = testUnmarshalMsg(br)
	if err != nil {
		t.Error(err)
	}
}

func testUnmarshalMsg(m interfaces.IMsg) error {
	d, err := m.MarshalBinary()
	if err != nil {
		return err
	}

	nd, nm, err := UnmarshalMessageData(d)
	if err != nil {
		return err
	}

	if len(nd) > 0 {
		return fmt.Errorf("%d bytes remain after unmarshal", len(nd))
	}

	d2, err := nm.MarshalBinary()
	if err != nil {
		return err
	}

	if bytes.Compare(d, d2) != 0 {
		return fmt.Errorf("New message differs from original")
	}

	return nil
}

func TestMessageNames(t *testing.T) {
	if constants.MessageName(constants.EOM_MSG) != "EOM" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.ACK_MSG) != "Ack" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.FULL_SERVER_FAULT_MSG) != "Full Server Fault" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.COMMIT_CHAIN_MSG) != "Commit Chain" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.COMMIT_ENTRY_MSG) != "Commit Entry" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.DIRECTORY_BLOCK_SIGNATURE_MSG) != "Directory Block Signature" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.EOM_TIMEOUT_MSG) != "EOM Timeout" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.FACTOID_TRANSACTION_MSG) != "Factoid Transaction" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.HEARTBEAT_MSG) != "HeartBeat" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.INVALID_ACK_MSG) != "Invalid Ack" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.INVALID_DIRECTORY_BLOCK_MSG) != "Invalid Directory Block" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.MISSING_MSG) != "Missing Msg" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.MISSING_MSG_RESPONSE) != "Missing Msg Response" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.MISSING_DATA) != "Missing Data" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.DATA_RESPONSE) != "Data Response" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.REVEAL_ENTRY_MSG) != "Reveal Entry" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.REQUEST_BLOCK_MSG) != "Request Block" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.SIGNATURE_TIMEOUT_MSG) != "Signature Timeout" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.DBSTATE_MISSING_MSG) != "DBState Missing" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.DBSTATE_MSG) != "DBState" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.BOUNCE_MSG) != "Bounce Message" {
		t.Error("EOM MessageName incorrect")
	}
	if constants.MessageName(constants.BOUNCEREPLY_MSG) != "Bounce Reply Message" {
		t.Error("EOM MessageName incorrect")
	}
}
