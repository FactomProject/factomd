// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"os"
	"testing"

	"github.com/PaulSnow/factom2d/common/messages"
	"github.com/PaulSnow/factom2d/common/primitives"
	. "github.com/PaulSnow/factom2d/testHelper"
)

// TODO - very incompleate journal test. needs to be expanded with different
// message types and more messages and bad messages and so on.
func TestJournal(t *testing.T) {
	s := CreateAndPopulateTestStateAndStartValidator()
	filename := "journaltest.log"
	s.JournalFile = filename
	_, err := os.Create(s.JournalFile)
	if err != nil {
		t.Errorf("%v", err)
	}
	s.Journaling = true

	msg := new(messages.Ack)
	msg.MsgHash = primitives.NewZeroHash()
	msg.MessageHash = primitives.NewZeroHash()
	msg.SerialHash = primitives.NewZeroHash()

	s.JournalMessage(msg)

	msgs := s.GetJournalMessages()
	if msgs == nil {
		t.Error("No messages returned from journal")
	}
}
