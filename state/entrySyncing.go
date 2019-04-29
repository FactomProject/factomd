// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/database/databaseOverlay"
)

type ReCheck struct {
	TimeToCheck int64            //Time in seconds to recheck
	EntryHash   interfaces.IHash //Entry Hash to check
	DBHeight    int
	NumEntries  int
	Tries       int
}

type EntrySync struct {
	MissingDBlockEntries chan []*ReCheck    // We don't have these entries.  Each list is from a directory block.
	DBHeightBase         int                // This is the highest block with entries not yet checked or are missing
	TotalEntries         int                // Total Entries in the database
	SyncingBlocks        map[int][]*ReCheck // Map of Directory blocks by height
	finishedDBlocks      chan int           // Channel of finished Directory blocks
	finishedEntries      chan int           // We get a ping every time an entry is done
	Processing           int                // Directory block we are processing
	EntriesProcessing    int                // Total of Entries being processed
	EntryRequests        int                // Requests made
	EntriesFound         int                // Entries found
}

// Maintain queues of what we want to test, and what we are currently testing.
func (es *EntrySync) Init() {
	es.MissingDBlockEntries = make(chan []*ReCheck, 1000) // Check 10 directory blocks at a time.
	es.finishedEntries = make(chan int)
	es.finishedDBlocks = make(chan int)
	es.SyncingBlocks = make(map[int][]*ReCheck)
} // we have to reprocess

func has(s *State, entry interfaces.IHash) bool {
	exists, err := s.DB.DoesKeyExist(databaseOverlay.ENTRY, entry.Bytes())
	if exists {
		if err != nil {
			return false
		}
	}
	return exists
}

var _ = fmt.Print

// WriteEntriesToTheDB()
// As Entries come in and are validated, then write them to the database
func (s *State) WriteEntries() {

	for {
		entry := <-s.WriteEntry
		if !has(s, entry.GetHash()) {
			s.DB.StartMultiBatch()
			err := s.DB.InsertEntryMultiBatch(entry)
			if err != nil {
				panic(err)
			}
			err = s.DB.ExecuteMultiBatch()
			if err != nil {
				panic(err)
			}
		}
	}
}

// RequestAndCollectMissingEntries()
// Manage go routines that are requesting and checking for missing entries
func (s *State) RequestAndCollectMissingEntries() {
	es := s.EntrySyncState

	biggest := func() (r int) {
		for k, _ := range es.SyncingBlocks {
			if k > r {
				r = k
			}
		}
		return
	}

	for {
		select {
		case dblock := <-es.finishedDBlocks:
			delete(es.SyncingBlocks, dblock)
			for es.SyncingBlocks[es.Processing] == nil && len(es.SyncingBlocks) > 0 {
				s.EntryBlockDBHeightComplete = uint32(es.Processing)
				s.EntryDBHeightComplete = uint32(es.Processing)
				es.Processing++
				s.EntryBlockDBHeightProcessing = uint32(es.Processing)
				s.EntryDBHeightProcessing = uint32(es.Processing)
			}
			s.LogPrintf("entrysyncing", "length %6d biggest %d", len(es.SyncingBlocks), biggest())
			continue
		case <-es.finishedEntries:
			es.EntriesProcessing--
			continue
		default:
		}

		s.LogPrintf("entrysyncing", "Processing dbht %6d %6d Entries processinng %6d Requests %6d Found %6d queue %6d",
			s.EntryDBHeightComplete,
			es.Processing,
			es.EntriesProcessing,
			es.EntryRequests,
			es.EntriesFound,
			len(es.MissingDBlockEntries))

		for es.EntriesProcessing < 5000 && len(es.MissingDBlockEntries) > 0 {
			dbrcs := <-es.MissingDBlockEntries
			es.SyncingBlocks[dbrcs[0].DBHeight] = dbrcs
			es.EntriesProcessing += len(dbrcs)
			go s.ProcessDBlock(es.finishedDBlocks, es.finishedEntries, dbrcs)
		}

		time.Sleep(1 * time.Second)
	}
}

func (s *State) ProcessDBlock(finishedDBlocks chan int, finishedEntries chan int, dbrcs []*ReCheck) {
	dbht := 0
	missing := true
	left := len(dbrcs)
	for i := 0; missing && len(dbrcs) > 0; i++ {
		missing = false

		for j, rc := range dbrcs {
			if rc == nil {
				continue
			}
			if dbht < rc.DBHeight {
				dbht = rc.DBHeight
			}
			// Handle the optimistic case first.
			if rc.EntryHash == nil || has(s, rc.EntryHash) {
				s.EntrySyncState.EntriesFound++
				if rc.DBHeight > dbht {
					dbht = rc.DBHeight
				}
				dbrcs[j] = nil
				if rc.EntryHash != nil {
					s.LogPrintf("entrysyncing", "Found %x dbht %6d tries %6d", rc.EntryHash.Bytes()[:6], dbht, rc.Tries)
				}
				left--
				if left == 0 {
					finishedDBlocks <- dbht
					s.LogPrintf("entrysyncing", "Directory Block Complete %6d", dbht)
				}
				finishedEntries <- 1
			} else {
				rc.Tries++
				s.EntrySyncState.EntryRequests++
				entryRequest := messages.NewMissingData(s, rc.EntryHash)
				entryRequest.SendOut(s, entryRequest)
				missing = true
			}
		}
		time.Sleep(5 * time.Second)
	}
	time.Sleep(time.Second)
}

// GoSyncEntries()
// Start up all of our supporting go routines, and run through the directory blocks and make sure we have
// all the entries they reference.
func (s *State) GoSyncEntries() {
	time.Sleep(5 * time.Second)
	s.EntrySyncState = new(EntrySync)
	s.EntrySyncState.Init() // Initialize our processes

	go s.WriteEntries()
	go s.RequestAndCollectMissingEntries()

	highestChecked := s.EntryDBHeightComplete

	lookingfor := 0
	for {

		highestSaved := s.GetHighestSavedBlk()

		for scan := highestChecked + 1; scan <= highestSaved; scan++ {

			db := s.GetDirectoryBlockByHeight(scan)

			// Wait for the database if we have to
			for db == nil {
				time.Sleep(1 * time.Second)
				db = s.GetDirectoryBlockByHeight(scan)
			}

			// If loading from the database, then give it a bit of preference by sleeping a bit
			if !s.DBFinished {
				time.Sleep(10 * time.Millisecond)
			}

			// Run through all the entry blocks and entries in each directory block.
			// If any entries are missing, collect them.  Then stuff them into the MissingDBlockEntries channel to
			// collect from the network.
			var entries []interfaces.IHash
			for _, ebKeyMR := range db.GetEntryHashes()[3:] {
				eBlock, err := s.DB.FetchEBlock(ebKeyMR)
				if err != nil {
					panic(err)
				}
				if err != nil {
					panic(err)
				}
				// Don't have an eBlock?  Huh. We can go on, but we can't advance.  We just wait until it
				// does show up.
				for eBlock == nil {
					time.Sleep(1 * time.Second)
					eBlock, _ = s.DB.FetchEBlock(ebKeyMR)
				}

				hashes := eBlock.GetEntryHashes()
				s.EntrySyncState.TotalEntries += len(hashes)
				for _, entryHash := range hashes {
					if entryHash.IsMinuteMarker() {
						continue
					}

					// Make sure we remove any pending commits
					ueh := new(EntryUpdate)
					ueh.Hash = entryHash
					ueh.Timestamp = db.GetTimestamp()
					s.UpdateEntryHash <- ueh

					// MakeMissingEntryRequests()
					// This go routine checks every so often to see if we have any missing entries or entry blocks.  It then requests
					// them if it finds entries in the missing lists.
					if !has(s, entryHash) {
						entries = append(entries, entryHash)
					}
				}
			}

			lookingfor += len(entries)

			//	s.LogPrintf("entrysyncing", "Missing entries total %10d at height %10d directory entries: %10d QueueLen %10d",
			//		lookingfor, scan, len(entries), len(s.EntrySyncState.MissingDBlockEntries))
			var rcs []*ReCheck
			for _, entryHash := range entries {
				rc := new(ReCheck)
				rc.EntryHash = entryHash
				rc.DBHeight = int(scan)
				rc.NumEntries = len(entries)
				rcs = append(rcs, rc)
			}
			// Make sure we have at least one entry to ensure we set the status right.
			// On mainnet we almost always have an entry, so to test to ensure this works, we always add it.
			rc := new(ReCheck)
			rc.DBHeight = int(scan)
			rc.NumEntries = len(entries)
			rcs = append(rcs, rc)

			s.EntrySyncState.MissingDBlockEntries <- rcs

			s.EntryBlockDBHeightProcessing = scan + 1
			s.EntryDBHeightProcessing = scan + 1
		}
		highestChecked = highestSaved
	}
}
