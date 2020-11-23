// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	"time"

	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

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

func TestFactomSecond(t *testing.T) {
	s := testHelper.CreateEmptyTestState()
	// Test the 10min
	testFactomSecond(t, s, 600, time.Second)

	// Test every half
	blktime := 600
	d := time.Second
	for i := 0; i < 9; i++ {
		testFactomSecond(t, s, blktime/2, d/2)
		blktime = blktime / 2
		d = d / 2
	}

	// Test different common vectors
	//		2min blocks == 1/5s seconds
	testFactomSecond(t, s, 120, time.Second/5)
	//		1min blocks == 1/10s seconds
	testFactomSecond(t, s, 60, time.Second/10)
	//		30s blocks == 1/20s seconds
	testFactomSecond(t, s, 30, time.Second/20)
	//		6s blocks == 1/100s seconds
	testFactomSecond(t, s, 6, time.Second/100)

}

func testFactomSecond(t *testing.T, s *State, blktime int, second time.Duration) {
	s.DirectoryBlockInSeconds = blktime
	fs := s.FactomSecond()
	// We set the floor on FactomSeconds at 250ms because as blocks speed up the network and processing delays don't and
	// we can have storms of messages because the repeat frequency gets too high.
	if second.Nanoseconds()/1e6 < 250 {
		second = 250 * time.Millisecond
	}
	if fs != second {
		//avg := (fs + second) / 2
		diff := fs - second
		if diff < 0 {
			diff = diff * -1
		}
		if diff < 2*time.Millisecond {
			// This is close enough to be correct
		} else {
			t.Errorf("Blktime=%ds, Expect second of %s, found %s. Difference %s", blktime, second, fs, diff)

		}
	}
}
