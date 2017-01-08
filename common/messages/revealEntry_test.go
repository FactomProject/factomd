// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"testing"
)

func TestMarshalUnmarshalRevealEntry(t *testing.T) {
	re := newRevealEntry()
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
		t.Errorf("Invalid message type unmarshalled - got %v, expected %v", re2.Type(), constants.REVEAL_ENTRY_MSG)
	}

	hex2, err := re2.(*RevealEntryMsg).MarshalBinary()
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
}

func newRevealEntry() *RevealEntryMsg {
	re := new(RevealEntryMsg)

	entry := new(entryBlock.Entry)

	entry.ExtIDs = make([]primitives.ByteSlice, 0, 5)
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("1asdfadfasdf")})
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("")})
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("3")})
	entry.ChainID = new(primitives.Hash)
	entry.ChainID.SetBytes(constants.EC_CHAINID)

	entry.Content = primitives.ByteSlice{Bytes: []byte("1asdf asfas dfsg\"08908098(*)*^*&%&%&$^#%##%$$@$@#$!$#!$#@!~@!#@!%#@^$#^&$*%())_+_*^*&^&\"\"?>?<<>/./,")}

	re.Entry = entry

	return re
}
