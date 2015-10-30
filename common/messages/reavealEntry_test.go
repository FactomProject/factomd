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

func TestMarshalUnmarshalRevealEntry(t *testing.T) {
	re := newDirectoryBlockSignature()
	hex, err := re.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	re2, err := UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str := re2.String()
	t.Logf("str - %v", str)

	if re2.Type() != constants.REVEAL_ENTRY_MSG {
		t.Error("Invalid message type unmarshalled")
	}
}

func newRevealEntry() *RevealEntry {
	re := new(RevealEntry)

	return re
}
