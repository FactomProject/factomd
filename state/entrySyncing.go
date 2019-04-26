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
	MissingDBlockEntries chan []*ReCheck // We don't have these entries.  Each list is from a directory block.
	DBHeightBase         int             // This is the highest block with entries not yet checked or are missing
	TotalEntries         int             // Total Entries in the database

}

// Maintain queues of what we want to test, and what we are currently testing.
func (es *EntrySync) Init() {
	es.MissingDBlockEntries = make(chan []*ReCheck, 10) // Check 10 directory blocks at a time.
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

// RequestAndCollectMissingEntries()
// We were missing these entries.  Check to see if we have them yet.  If we don't then schedule to recheck.
func (s *State) RequestAndCollectMissingEntries() {
	es := s.EntrySyncState
	for {
		dbht := 0
		missing := true
		for missing {
			dbrcs := <-es.MissingDBlockEntries
			missing = false
			found := 0
			for i, rc := range dbrcs {
				dbht = rc.DBHeight
				if rc != nil && !has(s, rc.EntryHash) {
					entryRequest := messages.NewMissingData(s, rc.EntryHash)
					entryRequest.SendOut(s, entryRequest)
					rc.Tries++
					missing = true
				} else {
					found++
					s.LogPrintf("entrysyncing", "%20s Entry:%x dbht:%6d tries:%d found:%6d/%6d",
						"Found Entry",
						rc.EntryHash.Bytes()[:6], rc.DBHeight, rc.Tries, found, len(dbrcs))
					dbrcs[i] = nil
				}
				rc.Tries++
			}
			time.Sleep(10 * time.Millisecond)
		}
		s.LogPrintf("entrysyncing", "%20s dbht %d",
			"Found Entry", dbht)
		s.EntryDBHeightComplete = uint32(dbht)
		s.EntryBlockDBHeightComplete = uint32(dbht)
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
	go s.RequestAndCollectMissingEntries()

	highestChecked := s.EntryDBHeightComplete
	lookingfor := 0
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
				if scan%100 == 0 {
					s.LogPrintf("entrysyncing", "DBHeight Complete %d", scan-1)
				}
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
						somethingMissing = true
					}
				}
			}
			for cap(s.EntrySyncState.MissingDBlockEntries) < len(s.EntrySyncState.MissingDBlockEntries)+cap(s.EntrySyncState.MissingDBlockEntries)/1000 {
				time.Sleep(time.Second)
			}

			lookingfor += len(entries)
			if scan%100 == 0 && lookingfor > 0 {
				s.LogPrintf("entrysyncing", "Missing entries total %10d at height %10d: %10d Queue %10d",
					lookingfor, scan, len(entries), len(s.EntrySyncState.MissingDBlockEntries))
			}

			if len(entries) > 0 {
				var rcs []*ReCheck
				for _, entryHash := range entries {
					rc := new(ReCheck)
					rc.EntryHash = entryHash
					rc.TimeToCheck = time.Now().Unix() + int64(s.DirectoryBlockInSeconds/100) // Don't check again for seconds
					rc.DBHeight = int(scan)
					rc.NumEntries = len(entries)
					rcs = append(rcs, rc)
				}
				s.EntrySyncState.MissingDBlockEntries <- rcs
			}
		}
		highestChecked = highestSaved
	}
}
