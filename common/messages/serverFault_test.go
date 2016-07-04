// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"testing"
)

func TestMarshalUnmarshalServerFault(t *testing.T) {
	ts := primitives.NewTimestampNow()
	vmIndex := int(*ts) % 10
	sf := NewServerFault(*ts, primitives.NewHash([]byte("a test")), vmIndex, 10, 100)
	hex, err := sf.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	sf2, err := UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str := sf2.String()
	t.Logf("str - %v", str)

	if sf2.Type() != constants.FED_SERVER_FAULT_MSG {
		t.Errorf("Invalid message type unmarshalled - got %v, expected %v", sf2.Type(), constants.FED_SERVER_FAULT_MSG)
	}

	hex2, err := sf2.(*ServerFault).MarshalBinary()
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
