// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/database/databaseOverlay"
)

func has(s *State, entry interfaces.IHash) bool {
	if s.GetHighestKnownBlock()-s.GetHighestSavedBlk() > 100 {
		if s.UsingTorrent() {
			// Torrents complete second pass
		} else {
			time.Sleep(30 * time.Millisecond)
		}
	}
	exists, err := s.DB.DoesKeyExist(databaseOverlay.ENTRY, entry.Bytes())
	if exists {
		if err != nil {
			return false
		}

		entry, err2 := s.DB.FetchEntry(entry)
		if err2 != nil || entry == nil {
			panic("Should not happen;  key exists but not entry")
			return false
		}
	}
	return exists
}

var _ = fmt.Print

// This go routine checks every so often to see if we have any missing entries or entry blocks.  It then requests
// them if it finds entries in the missing lists.
func (s *State) MakeMissingEntryRequests() {

	missing := 0
	found := 0

	MissingEntryMap := make(map[[32]byte]*MissingEntry)

	cnt := 0
	sum := 0
	avg := 0
	for {
		now := time.Now()

		newrequest := 0

		highest := 0

		// Look through our map, and remove any entries we now have in our database.
		for k := range MissingEntryMap {
			if has(s, MissingEntryMap[k].EntryHash) {
				found++
				ESMissingRequestLoopBackTime.Observe(float64(time.Since(MissingEntryMap[k].LastTime).Nanoseconds()))
				cnt++
				sum += MissingEntryMap[k].Cnt
				delete(MissingEntryMap, k)
			} else {
				if MissingEntryMap[k].DBHeight > uint32(highest) {
					highest = int(MissingEntryMap[k].DBHeight)
				}
			}
		}
		if cnt > 0 {
			avg = (1000 * sum) / cnt
		}

		ESAsking.Set(float64(len(MissingEntryMap)))
		ESAsking.Set(float64(cnt))
		ESFound.Set(float64(found))
		ESAvgRequests.Set(float64(avg) / 1000)
		ESHighestAsking.Set(float64(highest))

		// Keep our map of entries that we are asking for filled up.
	fillMap:
		for len(MissingEntryMap) < 1500 {
			select {
			case et := <-s.MissingEntries:
				missing++
				MissingEntryMap[et.EntryHash.Fixed()] = et
			default:
				break fillMap
			}
		}

		sent := 0
		if s.inMsgQueue.Length() < constants.INMSGQUEUE_MED {
			// Make requests for entries we don't have.
			for k := range MissingEntryMap {

				et := MissingEntryMap[k]

				max := 1500
				// If using torrent and the saved height is more than 750 behind, let torrent do it's work, and don't send out
				// missing message requests
				if s.UsingTorrent() && s.GetLeaderHeight() > 1000 && s.GetHighestSavedBlk() < s.GetLeaderHeight()-750 {
					max = 1
				}

				if now.Unix()-et.LastTime.Unix() > 5 && sent < max {
					sent++
					entryRequest := messages.NewMissingData(s, et.EntryHash)
					entryRequest.SendOut(s, entryRequest)
					ESMissingRequestCount.Inc()
					newrequest++
					et.LastTime = now.Add(3000 * time.Millisecond)
					et.Cnt++

					// Do (up to) 1 write per request to prevent bursts of writes/requests
					writeEntryToDisk(MissingEntryMap, s)
				}

				if sent >= max {
					break
				}
			}
		} else {
			time.Sleep(20 * time.Second)
		}

		// Insert the entries we have found into the database.
		for writeEntryToDisk(MissingEntryMap, s) {
		}

		if sent == 0 {
			if s.GetHighestKnownBlock()-s.GetHighestSavedBlk() > 100 {
				time.Sleep(10 * time.Second)
			} else {
				time.Sleep(100 * time.Millisecond)
			}
			if s.EntryDBHeightComplete == s.GetHighestSavedBlk() {
				time.Sleep(20 * time.Second)
			}
		}
	}
}
func writeEntryToDisk(MissingEntryMap map[[32]byte]*MissingEntry, s *State) bool {
	select {
	case entry := <-s.WriteEntry:
		asked := MissingEntryMap[entry.GetHash().Fixed()] != nil
		if asked {
			s.DB.StartMultiBatch()
			err := s.DB.InsertEntryMultiBatch(entry)
			if err != nil {
				panic(err)
			}
			err = s.DB.ExecuteMultiBatch()
			if err != nil {
				panic(err)
			}
			ESWritingToDiskCount.Inc()
		}
	default:
		return false
	}
	return true
}

func (s *State) GoSyncEntries() {
	go s.MakeMissingEntryRequests()

	// Map to track what I know is missing
	missingMap := make(map[[32]byte]interfaces.IHash)

	// Once I have found all the entries, we quit searching so much for missing entries.
	start := uint32(1)

	if s.EntryDBHeightComplete > 0 {
		start = s.EntryDBHeightComplete
	}

	entryMissing := 0

	// If I find no missing entries, then the firstMissing will be -1
	firstMissing := -1

	lastfirstmissing := 0

	found := 0

	for {

		ESMissing.Set(float64(len(missingMap)))
		ESMissingQueue.Set(float64(len(s.MissingEntries)))
		ESDBHTComplete.Set(float64(s.EntryDBHeightComplete))
		ESFirstMissing.Set(float64(lastfirstmissing))
		ESHighestMissing.Set(float64(s.GetHighestSavedBlk()))

		entryMissing = 0

		for k := range missingMap {
			if has(s, missingMap[k]) {
				found++
				delete(missingMap, k)
			}
		}

		// Scan all the directory blocks, from start to the highest saved.  Once we catch up,
		// start will be the last block saved.

		// First reset first Missing back to -1 every time.
		firstMissing = -1

	dirblkSearch:
		for scan := start; scan <= s.GetHighestSavedBlk(); scan++ {

			if firstMissing < 0 {
				if scan > 1 {
					s.EntryDBHeightComplete = scan - 1
					start = scan
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
					if entryhash.IsMinuteMarker() {
						continue
					}

					// Only update the replay hashes in the last 24 hours.
					if time.Now().Unix()-db.GetTimestamp().GetTimeSeconds() < 24*60*60 {
						ueh := new(EntryUpdate)
						ueh.Hash = entryhash
						ueh.Timestamp = db.GetTimestamp()
						s.UpdateEntryHash <- ueh
					}

					// If I have the entry, then remove it from the Missing Entries list.
					if has(s, entryhash) {
						found++
						delete(missingMap, entryhash.Fixed())
						continue
					}

					if firstMissing < 0 {
						firstMissing = int(scan)
					}

					eh := missingMap[entryhash.Fixed()]
					if eh == nil {

						// If we have a full queue, break so we don't stall.
						// If we stall, we don't properly update the s.EntryDBHeightComplete state, and then we
						// don't reasonably report the height of Entry Blocks scanned...
						if cap(s.MissingEntries)-len(s.MissingEntries) < 2 {
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
			}

			if s.EntryDBHeightComplete%1000 == 0 {
				if firstMissing < 0 {
					//Only save EntryDBHeightComplete IF it's a multiple of 1000 AND there are no missing entries
					err := s.DB.SaveDatabaseEntryHeight(s.EntryDBHeightComplete)
					if err != nil {
						fmt.Printf("ERROR: %v\n", err)
					}
				}
			}
		}
		lastfirstmissing = firstMissing
		if firstMissing < 0 {
			s.EntryDBHeightComplete = s.GetHighestSavedBlk()
			time.Sleep(5 * time.Second)
		}

		time.Sleep(100 * time.Millisecond)

	}
}
