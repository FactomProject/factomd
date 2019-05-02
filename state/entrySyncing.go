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

const (
	pendingRequests     = 10000 // Lower bound on pending requests while syncing entries
	timeBetweenRequests = 200   // Seconds
	purgeEveryXEntries  = 1000  // Every 1000 entries or so, go through the written map and purge old entries
)

type ReCheck struct {
	TimeToCheck int64            //Time in seconds to recheck
	EntryHash   interfaces.IHash //Entry Hash to check
	DBHeight    int
	NumEntries  int
	Tries       int
}
type EntrySync struct {
	MissingDBlockEntries     chan []*ReCheck            // We don't have these entries.  Each list is from a directory block.
	DBHeightBase             int                        // This is the highest block with entries not yet checked or are missing
	TotalEntries             int                        // Total Entries in the database
	SyncingBlocks            map[int][]*ReCheck         // Map of Directory blocks by height
	finishedDBlocks          chan int                   // Channel of finished Directory blocks
	finishedEntries          chan int                   // We get a ping every time an entry is done
	SendRequest              chan *messages.MissingData // Channel for messages being requested
	Processing               int                        // Directory block we are processing
	EntriesProcessing        int                        // Total of Entries being processed
	EntryRequests            int                        // Requests made
	EntriesFound             int                        // Entries found
	DirectoryBlocksInProcess int                        // Number of Directory blocks we are processing
}

// Maintain queues of what we want to test, and what we are currently testing.
func (es *EntrySync) Init() {
	es.MissingDBlockEntries = make(chan []*ReCheck, 1000) // Check 10 directory blocks at a time.
	es.finishedEntries = make(chan int, 10000)
	es.finishedDBlocks = make(chan int, 10000)
	es.SendRequest = make(chan *messages.MissingData, 1000)

	es.SyncingBlocks = make(map[int][]*ReCheck)
} // we have to reprocess

// todo: likely benefit if we cache the hash's last 1K written
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
		if entry != nil && !has(s, entry.GetHash()) {
			s.DB.StartMultiBatch()
			err := s.DB.InsertEntryMultiBatch(entry)
			if err != nil {
				panic(err)
			}
			s.LogPrintf("ehashes", "Add %x", entry.GetHash().Bytes()[:4])
			err = s.DB.ExecuteMultiBatch()
			if err != nil {
				panic(err)
			}
		} else if entry != nil {
			s.LogPrintf("ehashes", "Has %x", entry.GetHash().Bytes()[:4])
		} else {
			s.LogPrintf("ehashes", "Has a nil entry")
		}
	}
}

// SendManager keeps us from double sending entries on repeats.
func (s *State) SendManager() {
	es := s.EntrySyncState
	var EntriesRequested map[[32]byte]int64     // Time we last sent a request for this entry
	EntriesRequested = make(map[[32]byte]int64) // Make our map

	purge := purgeEveryXEntries

	for {
		missingData := <-es.SendRequest
		now := time.Now().Unix()

		// Every 1000 messages or so, purge our hash map.
		if purge <= 0 {
			for k, v := range EntriesRequested {
				if (now - v) >= timeBetweenRequests {
					delete(EntriesRequested, k)
				}
			}
			purge = purgeEveryXEntries
		}
		purge--

		lastCall, ok := EntriesRequested[missingData.RequestHash.Fixed()]
		if !ok || (now-lastCall) > timeBetweenRequests {
			if !has(s, missingData.RequestHash) {
				EntriesRequested[missingData.RequestHash.Fixed()] = now
				missingData.SendOut(s, missingData)
				s.EntrySyncState.EntryRequests++
				continue
			}
		}
	} // forever ...
}

// RequestAndCollectMissingEntries()
// Manage go routines that are requesting and checking for missing entries
func (s *State) RequestAndCollectMissingEntries() {
	es := s.EntrySyncState

	var highestDblock int

	for {
		select {
		case dblock := <-es.finishedDBlocks:
			es.DirectoryBlocksInProcess--
			if dblock > highestDblock {
				highestDblock = dblock
			}
			delete(es.SyncingBlocks, dblock)
		case <-es.finishedEntries:
			es.EntriesProcessing--
		default:
			time.Sleep(1 * time.Second)
		}

		// Update es.Processing (which tracks what directory block we are working on) and the state variables
		// others look at.
		for es.SyncingBlocks[es.Processing] == nil {
			if es.Processing <= highestDblock && highestDblock > 0 {
				if uint32(es.Processing) > s.EntryDBHeightComplete {
					s.EntryBlockDBHeightComplete = uint32(es.Processing)
					s.EntryDBHeightComplete = uint32(es.Processing)
					s.DB.SaveDatabaseEntryHeight(uint32(es.Processing))
				}

				es.Processing++
			} else {
				break
			}
		}

		s.LogPrintf("entrysyncing", "Processing dbht %6d %6d Entries processing %6d Requests %6d Found %6d queue %6d DBlocks %6d",
			s.EntryDBHeightComplete,
			es.Processing,
			es.EntriesProcessing,
			es.EntryRequests,
			es.EntriesFound,
			len(es.MissingDBlockEntries),
			es.DirectoryBlocksInProcess)

		for es.EntriesProcessing < pendingRequests && len(es.MissingDBlockEntries) > 0 {
			dbrcs := <-es.MissingDBlockEntries
			es.DirectoryBlocksInProcess++
			es.SyncingBlocks[dbrcs[0].DBHeight] = dbrcs
			es.EntriesProcessing += len(dbrcs)
			go s.ProcessDBlock(es.finishedDBlocks, es.finishedEntries, dbrcs)
		}

	}
}

func (s *State) ProcessDBlock(finishedDBlocks chan int, finishedEntries chan int, dbrcs []*ReCheck) {
	dbht := dbrcs[0].DBHeight
mainloop:
	for {

		// The empty directory block case.
		if len(dbrcs) == 1 && dbrcs[0].EntryHash == nil {
			s.EntrySyncState.finishedDBlocks <- dbrcs[0].DBHeight
			s.EntrySyncState.finishedEntries <- 0
			return
		}

		// This function does one pass over our directory block's entries
		LookForEntries := func() (progress bool) {
			allfound := false
			for i := 0; i < 101 && !allfound; i++ {
				allfound = true
				for ipass, rc := range dbrcs {
					switch {
					case rc == nil:
					case rc.EntryHash == nil:
						dbrcs[ipass] = nil
						finishedEntries <- 0 // It isn't a real entry, but we have to account for it.
					case has(s, rc.EntryHash):
						dbrcs[ipass] = nil
						s.EntrySyncState.EntriesFound++
						finishedEntries <- 0
					case i == 0: // For only the first pass do we ask for missing entries
						allfound = false // Only get here if the entryhash isn't found
						//	s.LogPrintf("entrysyncing", "looking for %x [%6d] dbht %6d tries %6d",
						//		rc.EntryHash.Bytes(), ipass, dbht, rc.Tries)
						entryRequest := messages.NewMissingData(s, rc.EntryHash).(*messages.MissingData)
						s.EntrySyncState.SendRequest <- entryRequest
						progress = true
						rc.Tries++
					default: //Don't get here unless the entry isn't found, so say that we haven't found everything
						allfound = false
					}
				}
				// Sleep for 1/100 of our send frequency.  Convert our frequency to microseconds and divide
				// by 1000 to get the 1/100 of our seconds in microseconds.  Then sleep that many microseconds.
				time.Sleep(time.Duration(timeBetweenRequests*1000000/100) * time.Microsecond)
			}
			return
		}

		// See if we have more to do.
		for _, rc := range dbrcs {
			// If I have a rc still, then I have more to do.
			if rc != nil {
				if LookForEntries() {
					continue mainloop
				}
			}
		}
		// We get here if there is nothing left to do.  Tell our parent process what directory block we finished
		finishedDBlocks <- dbht
		s.LogPrintf("entrysyncing", "Directory Block Complete %6d all Entries found %6d", dbht, s.EntrySyncState.EntriesFound)
		return
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
	go s.SendManager()
	go s.RequestAndCollectMissingEntries()

	highestChecked := s.EntryDBHeightComplete

	lookingfor := 0

	for {

		highestSaved := s.GetHighestSavedBlk()

		// Sleep often if we are caught up (to the best of our knowledge)
		if highestSaved == highestChecked {
			time.Sleep(time.Second)
		}

		for scan := highestChecked + 1; scan <= highestSaved; scan++ {

			db := s.GetDirectoryBlockByHeight(scan)

			// Wait for the database if we have to
			for db == nil {
				time.Sleep(1 * time.Second)
				db = s.GetDirectoryBlockByHeight(scan)
			}

			// If loading from the database, then give it a bit of preference by sleeping a bit
			if !s.DBFinished {
				time.Sleep(1 * time.Millisecond)
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
