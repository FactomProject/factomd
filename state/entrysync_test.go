package state_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
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
	go s.WriteEntries()
	defer func() {
		close(s.WriteEntry)
	}()

	// testHelper creates 10 blocks by default
	// copy these from database to DBStates
	// they all have entries that exist in db
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

	// test times out if something goes wrong here
	for s.GetEntryBlockDBHeightComplete() < baseBlocks-1 {
		time.Sleep(time.Millisecond * 100)
	}

	if pos := es.Position(); pos != baseBlocks {
		t.Fatalf("EntrySync position wrong. got = %d, want = %d", pos, baseBlocks)
	}

	// create blocks with missing entries that need to be synced
	allentries := make(map[[32]byte]*entryBlock.Entry)
	oldblockset := testHelper.CreateFullTestBlockSet()
	last := oldblockset[len(oldblockset)-1]
	for i := uint32(0); i < baseBlocks; i++ {
		last = testHelper.CreateTestBlockSetWithNetworkIDAndEBlocks(last, constants.LOCAL_NETWORK_ID, true, i != 5)

		s.DB.StartMultiBatch()
		err := s.DB.ProcessABlockMultiBatch(last.ABlock)
		if err != nil {
			t.Fatal(err)
		}

		if last.EBlock != nil {
			err = s.DB.ProcessEBlockMultiBatch(last.EBlock, true)
			if err != nil {
				t.Fatal(err)
			}
		}

		if last.AnchorEBlock != nil {
			err = s.DB.ProcessEBlockMultiBatch(last.AnchorEBlock, true)
			if err != nil {
				t.Fatal(err)
			}
		}

		err = s.DB.ProcessECBlockMultiBatch(last.ECBlock, false)
		if err != nil {
			t.Fatal(err)
		}

		err = s.DB.ProcessFBlockMultiBatch(last.FBlock)
		if err != nil {
			t.Fatal(err)
		}

		err = s.DB.ProcessDBlockMultiBatch(last.DBlock)
		if err != nil {
			t.Fatal(err)
		}
		err = s.DB.ExecuteMultiBatch()
		if err != nil {
			t.Fatal(err)
		}

		// don't save entries
		var eblocks []interfaces.IEntryBlock
		if last.EBlock != nil {
			eblocks = append(eblocks, last.EBlock)
		}

		if last.AnchorEBlock != nil {
			eblocks = append(eblocks, last.AnchorEBlock)
		}

		var entries []interfaces.IEBEntry
		for _, e := range last.Entries {
			allentries[e.GetHash().Fixed()] = e
			entries = append(entries, e)
		}

		s.DBStates.NewDBState(true, last.DBlock, last.ABlock, last.FBlock, last.ECBlock, eblocks, entries)
		s.DBStates.DBStates[baseBlocks+i].Saved = true
	}

	s.DBStates.ProcessHeight = baseBlocks * 2
	time.Sleep(time.Second * 2)

	if s.NetworkOutMsgQueue().Length() == 0 {
		t.Fatalf("expected message requests. got = %d, want = %d", s.NetworkOutMsgQueue().Length(), len(allentries))
	}
	for s.NetworkOutMsgQueue().Length() > 0 {
		imsg := s.NetworkOutMsgQueue().Dequeue()
		msg, ok := imsg.(*messages.MissingData)
		if !ok {
			t.Errorf("unexpected message in network queue. got = %v, want = *messages.MissingData", reflect.TypeOf(imsg))
			continue
		}

		response, ok := allentries[msg.RequestHash.Fixed()]
		if !ok {
			t.Errorf("response to %x not in map", msg.RequestHash)
		}

		if !es.AskedFor(msg.RequestHash) {
			t.Errorf("a request sent for a hash that wasn't asked for: %x", msg.RequestHash)
		}

		s.WriteEntry <- response
	}

	// test times out if something goes wrong here
	for s.GetEntryBlockDBHeightComplete() < baseBlocks*2-1 {
		time.Sleep(time.Millisecond * 100)
	}

	if pos := es.Position(); pos != baseBlocks*2 {
		t.Fatalf("EntrySync position wrong. got = %d, want = %d", pos, baseBlocks*2)
	}
}
