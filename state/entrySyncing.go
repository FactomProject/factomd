// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/messages"
	"math"
	"time"
)

var _ = fmt.Print

func fetchByTorrent(s *State, height uint32) {
	err := s.GetMissingDBState(height)
	if err != nil {
		fmt.Println("DEBUG: Error in torrent retrieve: " + err.Error())
	}
}

// This go routine checks every so often to see if we have any missing entries or entry blocks.  It then requests
// them if it finds entries in the missing lists.
func (s *State) MakeMissingEntryRequests() {

	type EntryTrack struct {
		lastRequest time.Time
		cnt         int
	}

	InPlay := make(map[[32]byte]*EntryTrack)

	found := 0

	for {
		now := time.Now()
		newfound := 0
		// Make a copy of Missing Entries while locked down, and do all the Entry Block syncing (because mostly
		// we are not doing any engry block syncing because they are synced with DBStates, but we do so for
		// sanity purposes.

		s.MissingEntryMutex.Lock()
		{
			// Remove all entry blocks we have already found and recorded
			var keepeb []MissingEntryBlock
			for _, v := range s.MissingEntryBlocks {
				eb, _ := s.DB.FetchEBlock(v.ebhash)
				if eb == nil {
					keepeb = append(keepeb, v)
				}
			}
			s.MissingEntryBlocks = keepeb

			// Ask for missing entry blocks
			for i, v := range s.MissingEntryBlocks {

				if i > 50 {
					break
				}

				eBlockRequest := messages.NewMissingData(s, v.ebhash)
				if s.UsingTorrent() {
					fetchByTorrent(s, v.dbheight)
				} else {
					eBlockRequest.SendOut(s, eBlockRequest)
				}
			}
		}

		s.MissingEntryMutex.Unlock()
		var keep []MissingEntry

		// Every call to update recomputes the keep array from the s.MissingEntries, and s.MissingEntries
		update := func() {

			s.MissingEntryMutex.Lock()
			defer s.MissingEntryMutex.Unlock()

			keep = make([]MissingEntry, 0)
			// Remove all Entries that we have already found and recorded. "keep" the ones we are still looking for.
			for _, v := range s.MissingEntries {
				e, _ := s.DB.FetchEntry(v.entryhash)
				if e == nil {
					keep = append(keep, v)
				} else {
					found++
					newfound++
					delete(InPlay, v.entryhash.Fixed())
				}
			}
			// Let the outside world know which entries we are looking for.
			s.MissingEntries = make([]MissingEntry, len(keep))
			copy(s.MissingEntries, keep)
		}

		update()

		// Ask for missing entries.

		feedback := func() {
			min := int(math.MaxInt32)
			max := 0
			maxcnt := 0
			for _, v := range keep {
				if min > int(v.dbheight) {
					min = int(v.dbheight)
				}
				if max < int(v.dbheight) {
					max = int(v.dbheight)
				}
				et := InPlay[v.entryhash.Fixed()]
				if et != nil && maxcnt < et.cnt {
					maxcnt = et.cnt
				}
			}

			if min >= int(math.MaxInt32) {
				min = 0
			}

			s.EntryDBHeightComplete = uint32(min)

			foundstr := fmt.Sprint(newfound, "/", found)
			newfound = 0
			fmt.Printf("***es Looking for: %8d Found: %13s In Play: %6d Min Height: %8d Max Height: %8d Max Send: %3d \n",
				len(keep),
				foundstr,
				len(InPlay),
				min,
				max,
				maxcnt)
		}

		var loopList []MissingEntry
		loopList = append(loopList, keep...)

		for i, v := range loopList {

			if i > 2000 {
				break
			}

			if (i+1)%100 == 0 {
				update()
				feedback()
			}

			entryTrack := InPlay[v.entryhash.Fixed()]

			if entryTrack == nil || now.Unix()-entryTrack.lastRequest.Unix() > 40 {
				entryRequest := messages.NewMissingData(s, v.entryhash)
				if s.UsingTorrent() {
					fetchByTorrent(s, v.dbheight)
				} else {
					entryRequest.SendOut(s, entryRequest)
				}

				time.Sleep(5 * time.Millisecond)

				if entryTrack == nil {
					entryTrack = new(EntryTrack)
					InPlay[v.entryhash.Fixed()] = entryTrack
				}
				entryTrack.lastRequest = now
				entryTrack.cnt++
				if entryTrack.cnt > 25 {
					fmt.Printf("***es Can't get Entry Block %x Entry %x \n", v.ebhash.Bytes(), v.entryhash.Bytes())
				}
			}
		}

		if len(InPlay) == 0 {
			time.Sleep(60 * time.Second)
		}
		feedback()
		time.Sleep(1 * time.Second)
	}
}

func (s *State) syncEntryBlocks() {
	// All done is true, and as long as it says true, we walk our bookmark forward.  Once we find something
	// missing, we stop moving the bookmark, and rely on caching to keep us from thrashing the disk as we
	// review the directory block over again the next time.
	alldone := true
	for s.EntryBlockDBHeightProcessing <= s.GetHighestSavedBlk() && len(s.MissingEntryBlocks) < 1000 {

		db := s.GetDirectoryBlockByHeight(s.EntryBlockDBHeightProcessing)

		if db == nil {
			return
		}

		for _, ebKeyMR := range db.GetEntryHashes()[3:] {
			// The first three entries (0,1,2) in every directory block are blocks we already have by
			// definition.  If we decide to not have Factoid blocks or Entry Credit blocks in some cases,
			// then this assumption might not hold.  But it does for now.

			eBlock, _ := s.DB.FetchEBlock(ebKeyMR)

			// Ask for blocks we don't have.
			if eBlock == nil {

				// Check lists and not add if already there.
				addit := true
				for _, eb := range s.MissingEntryBlocks {
					if eb.ebhash.Fixed() == ebKeyMR.Fixed() {
						addit = false
						break
					}
				}

				if addit {
					s.MissingEntryBlocks = append(s.MissingEntryBlocks,
						MissingEntryBlock{ebhash: ebKeyMR, dbheight: s.EntryBlockDBHeightProcessing})
				}
				// Something missing, stop moving the bookmark.
				alldone = false
				continue
			}
		}
		if alldone {
			// we had three bookmarks.  Now they are all in lockstep. TODO: get rid of extra bookmarks.
			s.EntryBlockDBHeightComplete++
			s.EntryBlockDBHeightProcessing++
		}
	}
}

func (s *State) SyncEntries() {
	scan := uint32(0)
	alldone := true
	for {

		s.MissingEntryMutex.Lock()

	scanEntries:
		for scan <= s.GetHighestSavedBlk() && len(s.MissingEntries) < 10000 {

			db := s.GetDirectoryBlockByHeight(scan)

			if db == nil {
				break scanEntries
			}

			for _, ebKeyMR := range db.GetEntryHashes()[3:] {
				// The first three entries (0,1,2) in every directory block are blocks we already have by
				// definition.  If we decide to not have Factoid blocks or Entry Credit blocks in some cases,
				// then this assumption might not hold.  But it does for now.

				eBlock, _ := s.DB.FetchEBlock(ebKeyMR)

				// Dont have an eBlock?  Huh. We can go on, but we can't advance
				if eBlock == nil {
					alldone = false
					break scanEntries
				}

				for _, entryhash := range eBlock.GetEntryHashes() {
					if entryhash.IsMinuteMarker() {
						continue
					}
					e, _ := s.DB.FetchEntry(entryhash)
					// If I have the entry, then remove it from the Missing Entries list.
					if e != nil {
						// If I am missing the entry, add it to th eMissing Entries list
					} else {
						alldone = false
						//Check lists and not add if already there.
						addit := true
						for _, e := range s.MissingEntries {
							if e.entryhash.Fixed() == entryhash.Fixed() {
								addit = false
								break
							}
						}

						if addit {
							var v MissingEntry

							v.dbheight = eBlock.GetHeader().GetDBHeight()
							v.entryhash = entryhash
							v.ebhash = ebKeyMR

							s.MissingEntries = append(s.MissingEntries, v)
						}
					}
					// Save the entry hash, and remove from commits IF this hash is valid in this current timeframe.
					s.Replay.SetHashNow(constants.REVEAL_REPLAY, entryhash.Fixed(), db.GetTimestamp())
					// If the save worked, then remove any commit that might be around.
					if !s.Replay.IsHashUnique(constants.REVEAL_REPLAY, entryhash.Fixed()) {
						delete(s.Commits, entryhash.Fixed())
					}
				}
			}
			if alldone {
				s.EntryDBHeightComplete = scan
			}
			scan++
		}
		s.MissingEntryMutex.Unlock()
		if scan == s.GetHighestSavedBlk() {
			time.Sleep(60 * time.Second)
		}
		time.Sleep(5 * time.Second)
	}
}

// This routine walks through the directory blocks, looking for missing entry blocks and entries.
// Once it finds something missing, it keeps that as a mark, and starts there the next time it is
// called.

func (s *State) CatchupEBlocks() {

	for {
		s.MissingEntryMutex.Lock()
		s.syncEntryBlocks()
		s.MissingEntryMutex.Unlock()
		time.Sleep(5 * time.Second)
	}
}
