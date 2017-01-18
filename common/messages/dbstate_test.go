// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

func TestUnmarshalNilDBStateMsg(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(DBStateMsg)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalDBStateMsg(t *testing.T) {
	msg := newDBStateMsg()

	hex, err := msg.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	msg2, err := UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str, _ := msg.JSONString()
	t.Logf("str1 - %v", str)
	str, _ = msg2.JSONString()
	t.Logf("str2 - %v", str)

	if msg2.Type() != constants.DBSTATE_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	hex2, err := msg2.(*DBStateMsg).MarshalBinary()
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

	if msg.IsSameAs(msg2.(*DBStateMsg)) != true {
		t.Errorf("DBStateMsg messages are not identical")
	}
}

func TestDBStateMsgValidate(t *testing.T) {
	return
	state := testHelper.CreateAndPopulateTestState()

	msg := new(DBStateMsg)
	if msg.Validate(state) >= 0 {
		t.Errorf("Empty DBState validated")
	}

	msg = newDBStateMsg()
	msg.DirectoryBlock.GetHeader().SetNetworkID(0x00)
	if msg.Validate(state) >= 0 {
		t.Errorf("Wrong network ID validated")
	}

	msg = newDBStateMsg()
	msg.DirectoryBlock.GetHeader().SetDBHeight(state.GetHighestSavedBlk() + 1)
	constants.CheckPoints[state.GetHighestSavedBlk()+1] = "123"
	if msg.Validate(state) >= 0 {
		t.Errorf("Wrong checkpoint validated")
	}

	delete(constants.CheckPoints, state.GetHighestSavedBlk()+1)

	msg = newDBStateMsg()
	msg.DirectoryBlock.GetHeader().SetDBHeight(state.GetHighestSavedBlk() + 1)
	if msg.Validate(state) <= 0 {
		t.Errorf("Proper block not validated!")
	}
}

func newDBStateMsg() *DBStateMsg {
	msg := new(DBStateMsg)
	msg.Timestamp = primitives.NewTimestampNow()

	set := testHelper.CreateTestBlockSet(nil)
	set = testHelper.CreateTestBlockSet(set)

	msg.DirectoryBlock = set.DBlock
	msg.AdminBlock = set.ABlock
	msg.FactoidBlock = set.FBlock
	msg.EntryCreditBlock = set.ECBlock
	msg.EBlocks = []interfaces.IEntryBlock{set.EBlock, set.AnchorEBlock}
	for _, e := range set.Entries {
		msg.Entries = append(msg.Entries, e)
	}

	return msg
}
