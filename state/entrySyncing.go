// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/database/databaseOverlay"

	"sync"
)

func (s *MakeMissingEntryRequestsInfo) has(DB interfaces.DBOverlaySimple, entry interfaces.IHash) bool {
	if s.HighestKnownBlock-s.HighestSavedBlk > 100 {
		if s.useTorrents {
			// Torrents complete second pass
		} else {
			time.Sleep(30 * time.Millisecond)
		}
	}
	exists, err := DB.DoesKeyExist(databaseOverlay.ENTRY, entry.Bytes())
	if exists {
		if err != nil {
			return false
		}
		entry, err2 := DB.FetchEntry(entry)
		if err2 != nil || entry == nil {
			return false
		}
	}
	return exists
}

var _ = fmt.Print

/*
type MakeMissingEntryRequestsInfo struct {
	useTorrents             bool
	HighestSavedBlk uint32
	HighestKnownBlock uint32
	LLeaderHeight   uint32
	EntryDBHeightComplete uint32
}

*/

// This go routine checks every so often to see if we have any missing entries or entry blocks.  It then requests
// them if it finds entries in the missing lists.
func (s *ShareWithEntrySync) MakeMissingEntryRequests(MakeMissingEntryRequestsInfoChannel chan MakeMissingEntryRequestsInfo, ss *MakeMissingEntryRequestsStatic) {

	var info MakeMissingEntryRequestsInfo

	missing := 0
	found := 0

	MissingEntryMap := make(map[[32]byte]*MissingEntry)

	for {
		now := time.Now()
		newrequest := 0
		cnt := 0
		sum := 0
		avg := 0
		highest := 0

		// Look through our map, and remove any entries we now have in our database.
		for k := range MissingEntryMap {
			if s.has(ss.DB, MissingEntryMap[k].EntryHash) {
				found++
				delete(MissingEntryMap, k)
			} else {
				cnt++
				sum += MissingEntryMap[k].Cnt
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
		for len(MissingEntryMap) < 3000 {
			select {
			case et := <-ss.MissingEntries:
				missing++
				MissingEntryMap[et.EntryHash.Fixed()] = et
			default:
				break fillMap
			}
		}

		sent := 0
		if ss.inMsgQueue.Length() < constants.INMSGQUEUE_MED {
			// Make requests for entries we don't have.
			for k := range MissingEntryMap {

				et := MissingEntryMap[k]

				if et.Cnt == 0 {
					et.Cnt = 1
					et.LastTime = now.Add(time.Duration((rand.Int() % 5000)) * time.Millisecond)
					continue
				}

				max := 100
				// If using torrent and the saved height is more than 750 behind, let torrent do it's work, and don't send out
				// missing message requests
				if info.useTorrents && info.LLeaderHeight > 1000 && info.HighestSavedBlk < info.LLeaderHeight-750 {
					max = 1
				}

				if now.Unix()-et.LastTime.Unix() > 5 && sent < max {
					sent++
					entryRequest := messages.NewMissingData(ss.GetTimestamp(), et.EntryHash)
					entryRequest.SendOut(ss.state, entryRequest)
					newrequest++
					et.LastTime = now.Add(time.Duration((rand.Int() % 5000)) * time.Millisecond)
					et.Cnt++
				}

			}
		} else {
			time.Sleep(20 * time.Second)
		}

		// Insert the entries we have found into the database.
	InsertLoop:
		for {

			select {

			case entry := <-ss.WriteEntry:

				asked := MissingEntryMap[entry.GetHash().Fixed()] != nil

				if asked {
					ss.DB.StartMultiBatch()
					err := ss.DB.InsertEntryMultiBatch(entry)
					if err != nil {
						panic(err)
					}
					err = ss.DB.ExecuteMultiBatch()
					if err != nil {
						panic(err)
					}
				}

			default:
				break InsertLoop
			}
		}

		// get info check if we need to make missing entries
		info := <-MakeMissingEntryRequestsInfoChannel // block if no update available

		if sent == 0 {
			if info.HighestKnownBlock-info.HighestSavedBlk > 100 {
				time.Sleep(10 * time.Second)
			} else {
				time.Sleep(100 * time.Millisecond)
			}
			if info.EntryDBHeightComplete == info.HighestSavedBlk {
				time.Sleep(20 * time.Second)
			}
		}
	}
}



func (s *ShareWithEntrySync) GoSyncEntries(wg *sync.WaitGroup, ss *ShareWithEntrySyncStatic) {

	// Feeds for worker threads
	var MakeMissingEntryRequestsInfoChannel chan MakeMissingEntryRequestsInfo = make(chan MakeMissingEntryRequestsInfo) // Info needed by MakeMissingEntries()

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

	wg.Done()

	// start a thread to make requests for missing entries (rely on GoSync being started late enough for the necessary init to be done.
	go s.MakeMissingEntryRequests(MakeMissingEntryRequestsInfoChannel, &ss.MakeMissingEntryRequestsStatic) // Start the MakeMissingEntryRequests() thread ..



	for {
		// Update Prometheus Stats
		ESMissing.Set(float64(len(missingMap)))
		ESMissingQueue.Set(float64(len(ss.MissingEntries)))
		ESDBHTComplete.Set(float64(s.EntryDBHeightComplete))
		ESFirstMissing.Set(float64(lastfirstmissing))
		ESHighestMissing.Set(float64(s.HighestSavedBlk))

		// feed the MakeMissingEntryRequests() thread
		// Send all the fields MakeMissingEntryRequests cares about
		MakeMissingEntryRequestsInfoChannel <- s.MakeMissingEntryRequestsInfo

		entryMissing = 0
		for k := range missingMap {
			if s.has(ss.DB, missingMap[k]) {
				found++
				delete(missingMap, k)
			}
		}

		// Scan all the directory blocks, from start to the highest saved.  Once we catch up,
		// start will be the last block saved.

		// First reset first Missing back to -1 every time.
		firstMissing = -1

	dirblkSearch:
		for scan := start; scan <= s.HighestSavedBlk; scan++ {

			if firstMissing < 0 {
				if scan > 1 {
					s.EntryDBHeightComplete = scan - 1
					start = scan
				}
			}

			db := ss.GetDirectoryBlockByHeight(scan)

			// Wait for the database if we have to
			for db == nil {
				time.Sleep(1 * time.Second)
				db = ss.GetDirectoryBlockByHeight(scan)
			}

			for _, ebKeyMR := range db.GetEntryHashes()[3:] {
				// The first three entries (0,1,2) in every directory block are blocks we already have by
				// definition.  If we decide to not have Factoid blocks or Entry Credit blocks in some cases,
				// then this assumption might not hold.  But it does for now.

				eBlock, _ := ss.DB.FetchEBlock(ebKeyMR)

				// Don't have an eBlock?  Huh. We can go on, but we can't advance.  We just wait until it
				// does show up.
				for eBlock == nil {
					time.Sleep(1 * time.Second)
					eBlock, _ = ss.DB.FetchEBlock(ebKeyMR)
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
						ss.UpdateEntryHash <- ueh
					}

					// If I have the entry, then remove it from the Missing Entries list.
					if s.has(ss.DB, entryhash) {
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
						if cap(ss.MissingEntries)-len(ss.MissingEntries) < 2 {
							break dirblkSearch
						}

						var v MissingEntry

						v.DBHeight = eBlock.GetHeader().GetDBHeight()
						v.EntryHash = entryhash
						v.EBHash = ebKeyMR
						entryMissing++
						missingMap[entryhash.Fixed()] = entryhash
						ss.MissingEntries <- &v
					}
				}
			}

			if s.EntryDBHeightComplete%1000 == 0 {
				if firstMissing < 0 {
					//Only save EntryDBHeightComplete IF it's a multiple of 1000 AND there are no missing entries
					err := ss.DB.SaveDatabaseEntryHeight(s.EntryDBHeightComplete)
					if err != nil {
						fmt.Printf("ERROR: %v\n", err)
					}
				}
			}
		}
		lastfirstmissing = firstMissing
		if firstMissing < 0 {
			s.EntryDBHeightComplete = s.HighestSavedBlk
			time.Sleep(5 * time.Second)
		}

		time.Sleep(100 * time.Millisecond)

	}
}
