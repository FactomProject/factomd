// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

/*
import (
	"os"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/state"
)

func TestJournal(t *testing.T) {
	s := new(State)
	filename := "journaltest.log"
	s.JournalFile = filename
	_, err := os.Create(s.JournalFile)
	if err != nil {
		t.Errorf("%v", err)
	}

	msg := new(messages.Ack)
	msg.MsgHash = primitives.NewZeroHash()
	msg.MessageHash = primitives.NewZeroHash()
	msg.SerialHash = primitives.NewZeroHash()
	msg.Timestamp = interfaces.Timestamp(0x112233)

	s.JournalMessage(msg)
}
*/
