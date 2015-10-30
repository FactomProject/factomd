// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/messages"
	//"github.com/FactomProject/factomd/common/primitives"
	"testing"
)

func TestMarshalUnmarshalCommitEntry(t *testing.T) {
	ce := newCommitEntry()
	hex, err := ce.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	ce2, err := UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str := ce2.String()
	t.Logf("str - %v", str)

	if ce2.Type() != constants.COMMIT_ENTRY_MSG {
		t.Error("Invalid message type unmarshalled")
	}
}

func newCommitEntry() *CommitEntryMsg {
	ce := new(CommitEntryMsg)

	return ce
}
