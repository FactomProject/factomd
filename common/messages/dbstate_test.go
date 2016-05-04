// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/testHelper"
)

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
	str := msg2.String()
	t.Logf("str - %v", str)

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

func newDBStateMsg() *DBStateMsg {
	msg := new(DBStateMsg)
	msg.Timestamp.SetTimeNow()

	set := testHelper.CreateTestBlockSet(nil)
	set = testHelper.CreateTestBlockSet(set)

	msg.DirectoryBlock = set.DBlock
	msg.AdminBlock = set.ABlock
	msg.FactoidBlock = set.FBlock
	msg.EntryCreditBlock = set.ECBlock
	msg.EntryBlocks = []interfaces.IEntryBlock{set.AnchorEBlock, set.EBlock}

	return msg
}
