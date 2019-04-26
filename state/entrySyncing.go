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
}

type EntrySync struct {
	EntryReCheck chan *ReCheck // Still don't have these guys.  Recheck
	DBHeightBase int           // This is the highest block with entries not yet checked or are missing
	EntryCounts  []int         // [0] -- highest block with entries to be checked, [1] is the next, etc.
}

// Maintain queues of what we want to test, and what we are currently testing.
func (es *EntrySync) Init() {
	es.EntryReCheck = make(chan *ReCheck, 10000) // To avoid deadlocks, we queue requests here,
} // we have to reprocess

func has(s *State, entry interfaces.IHash) bool {
	if s.GetHighestKnownBlock()-s.GetHighestSavedBlk() > 100 {
		if s.UsingTorrent() {
			// Torrents complete second pass
		} else {
			time.Sleep(3 * time.Millisecond)
		}
	}
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

// RecheckMissingEntryRequests()
// We were missing these entries.  Check to see if we have them yet.  If we don't then schedule to recheck.
func (s *State) RecheckMissingEntryRequests() {
	es := s.EntrySyncState
	// Check if they have shown up
	for {
		rc := <-es.EntryReCheck
		idx := rc.DBHeight - es.DBHeightBase

		// Make sure the request is reasonable, and die if it isn't
		if idx < 0 {
			panic("A Hash that we are checking should not be from a dbheight lower than our base height")
		}

		// Make sure our tracking covers the height.  New entries are negative.
		for len(es.EntryCounts) <= idx {
			es.EntryCounts = append(es.EntryCounts, -1)
		}

		// Update the entry if it is new with the number of entries for the directory block.
		if es.EntryCounts[idx] == -1 {
			es.EntryCounts[idx] = rc.NumEntries
		}

		now := time.Now().Unix()
		if now < rc.TimeToCheck { // If we are not there yet, sleep
			time.Sleep(time.Duration(rc.TimeToCheck-now) * time.Second) // until it is time to check this guy.
		}

		if !has(s, rc.EntryHash) {
			entryRequest := messages.NewMissingData(s, rc.EntryHash)
			entryRequest.SendOut(s, entryRequest)
			rc.TimeToCheck = time.Now().Unix() + int64(s.DirectoryBlockInSeconds/100) // Don't check again for seconds
			go func() { es.EntryReCheck <- rc }()
		} else {
			// If we have this entry, then great!  Knock it off the count for a directory block.
			// When that goes to zero, we update the world the directory block if finished, if we can.
			es.EntryCounts[idx]--

			// Note that where idx == 0, further passes idx no longer indexes our rc's dbheight.  That's okay,
			// we are just checking because our entrycount at idx went to zero.  We could add an if around
			// our for loop, or we could just check it each pass (really cheap).  I'm doing the latter.
			for idx == 0 && len(es.EntryCounts) > 0 && es.EntryCounts[0] == 0 {
				// Copy the counts forward one slot, and reduce the size of the array
				copy(es.EntryCounts[0:], es.EntryCounts[1:])
				es.EntryCounts = es.EntryCounts[:len(es.EntryCounts)-1]
				// Move the base Directory Block forward one.
				es.DBHeightBase++
				// Update the world that a Directory Block has been verified
				s.EntryBlockDBHeightComplete = uint32(es.DBHeightBase - 1)
				s.EntryDBHeightComplete = uint32(es.DBHeightBase - 1)
			}
		}
	}
}

// GoSyncEntries()
// Start up all of our supporting go routines, and run through the directory blocks and make sure we have
// all the entries they reference.
func (s *State) GoSyncEntries() {
	time.Sleep(5 * time.Second)
	s.EntrySyncState = new(EntrySync)
	s.EntrySyncState.Init() // Initialize our processes
	go s.WriteEntries()
	go s.RecheckMissingEntryRequests()

	highestChecked := s.EntryDBHeightComplete

	for {

		if !s.DBFinished {
			time.Sleep(time.Second / 30)
		}

		highestSaved := s.GetHighestSavedBlk()
		if highestSaved <= highestChecked {
			time.Sleep(time.Duration(s.DirectoryBlockInSeconds/20) * time.Second)
			continue
		}
		somethingMissing := false
		for scan := highestChecked + 1; scan <= highestSaved; scan++ {
			// Okay, stuff we pull from wherever but there is nothing missing, then update our variables.
			if !somethingMissing && scan > 0 && s.EntryDBHeightComplete < scan-1 {
				s.EntryBlockDBHeightComplete = scan - 1
				s.EntryDBHeightComplete = scan - 1
				s.EntrySyncState.DBHeightBase = int(scan) // The base is the height of the block that might have something missing.
			}

			s.EntryBlockDBHeightProcessing = scan
			s.EntryDBHeightProcessing = scan

			db := s.GetDirectoryBlockByHeight(scan)

			// Wait for the database if we have to
			for db == nil {
				time.Sleep(1 * time.Second)
				db = s.GetDirectoryBlockByHeight(scan)
			}

			// Run through all the entry blocks and entries in each directory block.
			// If any entries are missing, collect them.  Then stuff them into the EntryReCheck channel to
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
						somethingMissing = true
					}
				}
			}
			for cap(s.EntrySyncState.EntryReCheck) < len(s.EntrySyncState.EntryReCheck)+cap(s.EntrySyncState.EntryReCheck)/1000 {
				time.Sleep(time.Second)
			}
			for _, entryHash := range entries {
				rc := new(ReCheck)
				rc.EntryHash = entryHash
				rc.TimeToCheck = time.Now().Unix() + int64(s.DirectoryBlockInSeconds/100) // Don't check again for seconds
				rc.DBHeight = int(scan)
				rc.NumEntries = len(entries)
				s.EntrySyncState.EntryReCheck <- rc
			}
		}
		highestChecked = highestSaved
	}
}
