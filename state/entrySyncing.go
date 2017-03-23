// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"time"
)

func has(s *State, entry interfaces.IHash) bool {
	exists, _ := s.DB.DoesKeyExist(databaseOverlay.ENTRY, entry.Bytes())
	return exists
}

var _ = fmt.Print

// This go routine checks every so often to see if we have any missing entries or entry blocks.  It then requests
// them if it finds entries in the missing lists.
func (s *State) MakeMissingEntryRequests() {

	missing := 0
	found := 0

	MissingEntryMap := make(map[[32]byte]*MissingEntry)

	for {

		now := time.Now()

		newrequest := 0

		cnt := 0
		sum := 0
		avg := 0

		// Look through our map, and remove any entries we now have in our database.
		for k := range MissingEntryMap {
			if has(s, MissingEntryMap[k].EntryHash) {
				found++
				delete(MissingEntryMap, k)
			} else {
				cnt++
				sum += MissingEntryMap[k].Cnt
			}
		}
		if cnt > 0 {
			avg = (1000 * sum) / cnt
		}

		fmt.Printf("***es %-10s "+
			"EComplete: %6d "+
			"Len(MissingEntyrMap): %6d "+
			"Avg: %6d.%03d "+
			"Missing: %6d  "+
			"Found: %6d "+
			"Queue: %d\n",
			s.FactomNodeName,
			s.EntryDBHeightComplete,
			len(MissingEntryMap),
			avg/1000, avg%1000,
			missing,
			found,
			len(s.MissingEntries))

		// Keep our map of entries that we are asking for filled up.
	fillMap:
		for len(MissingEntryMap) < 3000 {
			select {
			case et := <-s.MissingEntries:
				if !has(s, et.EntryHash) {
					missing++
					MissingEntryMap[et.EntryHash.Fixed()] = et
				}
			default:
				break fillMap
			}
		}

		if len(s.inMsgQueue) < 500 {
			// Make requests for entries we don't have.
			for k := range MissingEntryMap {
				et := MissingEntryMap[k]

				if et.Cnt == 0 || now.Unix()-et.LastTime.Unix() > 60 {
					entryRequest := messages.NewMissingData(s, et.EntryHash)
					entryRequest.SendOut(s, entryRequest)
					newrequest++
					et.LastTime = now
					et.Cnt++
					if et.Cnt%25 == 25 {
						fmt.Printf("***es Can't get Entry Block %x Entry %x in %v attempts.\n", et.EBHash.Bytes(), et.EntryHash.Bytes(), et.Cnt)
					}
				}
			}
		} else {
			time.Sleep(20 * time.Second)
		}

		// Insert the entries we have found into the database.
	InsertLoop:
		for {

			select {

			case entry := <-s.WriteEntry:

				asked := MissingEntryMap[entry.GetHash().Fixed()] != nil

				if asked {
					s.DB.InsertEntry(entry)
				}

			default:
				break InsertLoop
			}
		}
		if s.GetHighestKnownBlock()-s.GetHighestSavedBlk() > 100 {
			time.Sleep(30 * time.Second)
		} else {
			time.Sleep(5 * time.Second)
		}
		if s.EntryDBHeightComplete == s.GetHighestSavedBlk() {
			time.Sleep(20 * time.Second)
		}
	}
}

func (s *State) GoSyncEntries() {
	go s.MakeMissingEntryRequests()

	now := time.Now().Unix()
	// Map to track what I know is missing
	missingMap := make(map[[32]byte]interfaces.IHash)

	// Once I have found all the entries, we quit searching so much for missing entries.
	start := uint32(1)
	entryMissing := 0

	// If I find no missing entries, then the firstMissing will be -1
	firstMissing := -1

	lastfirstmissing := 0

	num := 0
	if nil != s.NetworkControler {
		num = s.NetworkControler.NumConnections
	}

	for {
		fmt.Printf("***es %10s"+
			" connections %d"+
			" t %6d"+
			" EntryDBHeightComplete %d"+
			" start %6d"+
			" end %7d"+
			" Missing: %6d"+
			" MissingMap %6d"+
			" FirstMissing %6d\n",
			s.FactomNodeName,
			num,
			time.Now().Unix()-now,
			s.EntryDBHeightComplete,
			start,
			s.GetHighestSavedBlk(),
			entryMissing,
			len(missingMap),
			lastfirstmissing)
		entryMissing = 0

		for k := range missingMap {
			if has(s, missingMap[k]) {
				delete(missingMap, k)
			}
		}

		// Scan all the directory blocks, from start to the highest saved.  Once we catch up,
		// start will be the last block saved.
	dirblkSearch:
		for scan := start; scan <= s.GetHighestSavedBlk(); scan++ {

			if firstMissing < 0 {
				if scan > 1 {
					s.EntryDBHeightComplete = scan - 1
				}
			}

			db := s.GetDirectoryBlockByHeight(scan)

			// Wait for the database if we have to
			for db == nil {
				time.Sleep(1 * time.Second)
				db = s.GetDirectoryBlockByHeight(scan)
			}

			for _, ebKeyMR := range db.GetEntryHashes()[3:] {
				// The first three entries (0,1,2) in every directory block are blocks we already have by
				// definition.  If we decide to not have Factoid blocks or Entry Credit blocks in some cases,
				// then this assumption might not hold.  But it does for now.

				eBlock, _ := s.DB.FetchEBlock(ebKeyMR)

				// Dont have an eBlock?  Huh. We can go on, but we can't advance.  We just wait until it
				// does show up.
				for eBlock == nil {
					time.Sleep(1 * time.Second)
					eBlock, _ = s.DB.FetchEBlock(ebKeyMR)
				}

				// Go through all the entry hashes.
				for _, entryhash := range eBlock.GetEntryHashes() {
					if !entryhash.IsMinuteMarker() {

						// If I have the entry, then remove it from the Missing Entries list.
						if has(s, entryhash) {
							delete(missingMap, entryhash.Fixed())
						} else {

							if firstMissing < 0 {
								firstMissing = int(scan)
							}

							eh := missingMap[entryhash.Fixed()]
							if eh == nil {

								// If we have a full queue, break so we don't stall.
								if len(s.MissingEntries) > 9000 {
									break dirblkSearch
								}

								var v MissingEntry

								v.DBHeight = eBlock.GetHeader().GetDBHeight()
								v.EntryHash = entryhash
								v.EBHash = ebKeyMR
								entryMissing++
								missingMap[entryhash.Fixed()] = entryhash
								s.MissingEntries <- &v
							}
						}
						ueh := new(EntryUpdate)
						ueh.Hash = entryhash
						ueh.Timestamp = db.GetTimestamp()
						s.UpdateEntryHash <- ueh
					}
				}
			}
			start = scan
		}
		lastfirstmissing = firstMissing
		if firstMissing < 0 {
			s.EntryDBHeightComplete = s.GetHighestSavedBlk()
			time.Sleep(60 * time.Second)
		}

		start = s.EntryDBHeightComplete

		// reset first Missing back to -1 every time.
		firstMissing = -1

		if s.GetHighestKnownBlock()-s.GetHighestSavedBlk() > 100 {
			time.Sleep(20 * time.Second)
		} else {
			time.Sleep(1 * time.Second)
		}

	}
}
