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
}

type EntrySync struct {
	CheckThese   chan interfaces.IHash // hashes of entries to be checked
	EntryReCheck chan *ReCheck         // Still don't have these guys.  Recheck
}

// Maintain queues of what we want to test, and what we are currently testing.
func (es *EntrySync) Init() {
	es.CheckThese = make(chan interfaces.IHash, 5000)
	es.EntryReCheck = make(chan *ReCheck, 1000) // To avoid deadlocks, we queue requests here,
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

// MakeMissingEntryRequests()
// This go routine checks every so often to see if we have any missing entries or entry blocks.  It then requests
// them if it finds entries in the missing lists.
func (s *State) MakeMissingEntryRequests() {
	// Check our hard drive at full blast
	for {
		entryHash := <-s.EntrySyncState.CheckThese
		if !has(s, entryHash) {
			rc := new(ReCheck)
			rc.EntryHash = entryHash
			rc.TimeToCheck = time.Now().Unix() + int64(s.DirectoryBlockInSeconds/100) // Don't check again for seconds
			s.EntrySyncState.EntryReCheck <- rc
		}
	}
}

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

	// Check if they have shown up
	for {
		rc := <-s.EntrySyncState.EntryReCheck
		now := time.Now().Unix()
		if now < rc.TimeToCheck { // If we are not there yet, sleep
			time.Sleep(time.Duration(rc.TimeToCheck-now) * time.Second) // until it is time to check this guy.
		}
		if !has(s, rc.EntryHash) {
			entryRequest := messages.NewMissingData(s, rc.EntryHash)
			entryRequest.SendOut(s, entryRequest)
			rc.TimeToCheck = time.Now().Unix() + int64(s.DirectoryBlockInSeconds/100) // Don't check again for seconds
			go func() { s.EntrySyncState.EntryReCheck <- rc }()
		}
	}
}

// GoSyncEntries()
// Start up all of our supporting go routines, and run through the directory blocks and make sure we have
// all the entries they reference.
func (s *State) GoSyncEntries() {
	time.Sleep(5 * time.Second)
	s.EntrySyncState.Init()         // Initialize our processes
	go s.MakeMissingEntryRequests() // Start our go routines
	go s.WriteEntries()
	go s.RecheckMissingEntryRequests()

	highestChecked := s.EntryDBHeightComplete
	for {

		if !s.DBFinished {
			time.Sleep(time.Second / 30)
		}

		highestSaved := s.GetHighestSavedBlk()
		if highestSaved <= highestChecked {
			if len(s.EntrySyncState.CheckThese) == 0 &&
				len(s.EntrySyncState.EntryReCheck) == 0 {
				s.EntryDBHeightComplete = highestSaved
				s.EntryBlockDBHeightComplete = highestSaved
			}
			time.Sleep(time.Duration(s.DirectoryBlockInSeconds/20) * time.Second)
			continue
		}

		for scan := highestChecked + 1; scan <= highestSaved; scan++ {
			s.EntryBlockDBHeightProcessing = scan
			s.EntryDBHeightProcessing = scan

			db := s.GetDirectoryBlockByHeight(scan)

			// Wait for the database if we have to
			for db == nil {
				time.Sleep(1 * time.Second)
				db = s.GetDirectoryBlockByHeight(scan)
			}

			// Run through all the eblocks, and make sure we have updated the Entry Hash for every Entry
			// Hash in the EBlocks.  This only has to be done one for all the EBlocks of a directory Block,
			// and we have the entry hashes even if we don't yet have the entries, so this is really simple.
			for _, ebKeyMR := range db.GetEntryHashes()[3:] {
				eBlock, _ := s.DB.FetchEBlock(ebKeyMR)

				// Don't have an eBlock?  Huh. We can go on, but we can't advance.  We just wait until it
				// does show up.
				for eBlock == nil {
					time.Sleep(1 * time.Second)
					eBlock, _ = s.DB.FetchEBlock(ebKeyMR)
				}

				for _, entryhash := range eBlock.GetEntryHashes() {
					if entryhash.IsMinuteMarker() {
						continue
					}

					// Make sure we remove any pending commits
					ueh := new(EntryUpdate)
					ueh.Hash = entryhash
					ueh.Timestamp = db.GetTimestamp()
					s.UpdateEntryHash <- ueh

					s.EntrySyncState.CheckThese <- entryhash
				}
			}
		}
		highestChecked = highestSaved
	}
}
