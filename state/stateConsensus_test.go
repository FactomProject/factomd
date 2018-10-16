// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

var _ = NewProcessList

func TestIsHighestCommit(t *testing.T) {
	s := testHelper.CreateAndPopulateTestStateAndStartValidator()

	commit := newCom()
	eh := commit.CommitEntry.EntryHash

	if !s.IsHighestCommit(eh, commit) {
		t.Error("Should be the highest, as it does not exist")
	}

	s.Commits.Put(eh.Fixed(), commit)
	if s.IsHighestCommit(eh, commit) {
		t.Error("Should not be the highest")
	}

	commit2 := newCom()
	commit2.CommitEntry.Credits++

	if !s.IsHighestCommit(eh, commit2) {
		t.Error("Should be the highest, as it is greater than the previous")
	}
}

func newCom() *messages.CommitEntryMsg {
	commit := messages.NewCommitEntryMsg()
	commit.CommitEntry = entryCreditBlock.NewCommitEntry()
	commit.CommitEntry.Credits = 2
	commit.CommitEntry.Init()
	eh := commit.CommitEntry.Hash()
	commit.CommitEntry.EntryHash = eh

	return commit
}
