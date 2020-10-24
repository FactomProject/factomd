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

// TestEntrySync builds an environment to test various features of EntrySync at once
// 1. See if we can verify the existence of the 10 test blocks that are added by the testhelper
// 2. Create ten more blocks but don't save the entries in the database (one block has no entries)
// 3. Check that the code makes requests for the right entries
// 4. Add missing entries to the "WriteEntries" system manually
// 5. Check that EntrySync will now have synced to the end of 20 blocks
func TestEntrySync(t *testing.T) {

	baseBlocks := uint32(testHelper.BlockCount) // default 10
	s := testHelper.CreateAndPopulateTestState()

	es := state.NewEntrySync(s)
	s.EntrySync = es
	go es.SyncHeight()
	defer es.Stop()
	go s.WriteEntries()
	defer func() {
		close(s.WriteEntry)
	}()

	blockset := testHelper.CreateFullTestBlockSet() // deterministic so we get the same as in CreateAndPopulateTestState
	// CreateAndPopulateTestState doesn't update state.DBStates, do it manually
	for i, bs := range blockset {
		var entries []interfaces.IEBEntry
		for _, e := range bs.Entries {
			entries = append(entries, e)
		}

		s.DBStates.NewDBState(true, bs.DBlock, bs.ABlock, bs.FBlock, bs.ECBlock, []interfaces.IEntryBlock{bs.EBlock}, entries)
		s.DBStates.DBStates[i].Saved = true
	}
	s.DBStates.ProcessHeight = baseBlocks // manually move this

	// test times out if something goes wrong here
	for s.GetEntryBlockDBHeightComplete() < baseBlocks-1 {
		time.Sleep(time.Millisecond * 100)
	}

	// 1.
	if pos := es.Position(); pos != baseBlocks {
		t.Fatalf("EntrySync position wrong. got = %d, want = %d", pos, baseBlocks)
	}

	// 2.
	allentries := make(map[[32]byte]*entryBlock.Entry)
	last := blockset[len(blockset)-1]
	for i := uint32(0); i < baseBlocks; i++ {
		// block 5 has no eblocks
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
	time.Sleep(time.Second * 2) // internal entrysync wait timer = 1s

	// 3.
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

		// 4.
		if !es.AskedFor(msg.RequestHash) {
			t.Errorf("a request sent for a hash that wasn't asked for: %x", msg.RequestHash)
		}
		s.WriteEntry <- response
	}

	// test times out if something goes wrong here
	for s.GetEntryBlockDBHeightComplete() < baseBlocks*2-1 {
		time.Sleep(time.Millisecond * 100)
	}

	// 5.
	if pos := es.Position(); pos != baseBlocks*2 {
		t.Fatalf("EntrySync position wrong. got = %d, want = %d", pos, baseBlocks*2)
	}
}
