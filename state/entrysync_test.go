package state_test

import (
	"testing"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

func TestEntrySync(t *testing.T) {

	baseBlocks := uint32(testHelper.BlockCount)

	s := testHelper.CreateAndPopulateTestState()

	es := state.NewEntrySync(s)
	s.EntrySync = es

	go es.SyncHeight()
	defer es.Stop()

	// testHelper creates 10 blocks by default
	// copy these from database to DBStates
	for i := uint32(0); i < baseBlocks; i++ {
		db, err := s.DB.FetchDBlockByHeight(i)
		if err != nil || db == nil {
			t.Fatal(err)
		}

		ablock, err := s.DB.FetchABlockByHeight(i)
		if err != nil || ablock == nil {
			t.Fatal(err)
		}

		fblock, err := s.DB.FetchFBlockByHeight(i)
		if err != nil || fblock == nil {
			t.Fatal(err)
		}

		ecblock, err := s.DB.FetchECBlockByHeight(i)
		if err != nil || ecblock == nil {
			t.Fatal(err)
		}

		var eblocks []interfaces.IEntryBlock
		for _, entry := range db.GetEBlockDBEntries() {
			eblock, err := s.DB.FetchEBlock(entry.GetKeyMR())
			if err != nil || eblock == nil {
				t.Fatal(err)
			}
			eblocks = append(eblocks, eblock)
		}

		var entries []interfaces.IEBEntry
		for _, eb := range eblocks {
			for _, h := range eb.GetEntryHashes() {
				if h.IsMinuteMarker() {
					continue
				}
				entry, err := s.DB.FetchEntry(h)
				if err != nil {
					t.Fatal(err)
				}
				if entry == nil {
					t.Fatalf("missing entry with hash %s", h.String())
				}
				entries = append(entries, entry)
			}
		}

		s.DBStates.NewDBState(true, db, ablock, fblock, ecblock, eblocks, entries)
		s.DBStates.DBStates[i].Saved = true
	}

	s.DBStates.ProcessHeight = baseBlocks // manually move this

	time.Sleep(time.Second)

	if s.GetDBHeightComplete() != baseBlocks-1 { // starts at 0
		t.Fatalf("EntrySync was unable to sync the %d existing database blocks. got = %d, want = %d", baseBlocks, s.GetDBHeightComplete(), baseBlocks-1)
	}

	if pos := es.Position(); pos != baseBlocks {
		t.Fatalf("EntrySync position wrong. got = %d, want = %d", pos, baseBlocks)
	}
}
