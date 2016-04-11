// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/messages"
	"testing"

	"github.com/FactomProject/factomd/common/primitives"
)

func TestMarshalUnmarshalAddServer(t *testing.T) {
    
	addserv := new(AddServerMsg)
    ts := new(interfaces.Timestamp)
	ts.SetTimeNow()
    addserv.Timestamp = *ts
    addserv.ServerChainID = primitives.Sha([]byte("FNode0"))

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

	addserv2, err := UnmarshalMessage(hex)
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

}

// TODO: Add test for signed messages (See ack_test.go)
