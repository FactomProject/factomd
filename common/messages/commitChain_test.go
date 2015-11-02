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

func TestMarshalUnmarshalCommitChain(t *testing.T) {
	cc := newCommitChain()
	hex, err := cc.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	cc2, err := UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str := cc2.String()
	t.Logf("str - %v", str)

	if cc2.Type() != constants.COMMIT_CHAIN_MSG {
		t.Error("Invalid message type unmarshalled")
	}
}

func newCommitChain() *CommitChainMsg {
	cc := new(CommitChainMsg)

	return cc
}
