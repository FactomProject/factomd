// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/messages"
	"time"
)

var _ = fmt.Print

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
		s.MissingEntryMutex.Lock()

		// Remove all Entries that we have already found and recorded.
		var keep []MissingEntry
		for _, v := range s.MissingEntries {
			e, _ := s.DB.FetchEntry(v.entryhash)
			if e == nil {
				keep = append(keep, v)
			} else {
				found++
				delete(InPlay, v.entryhash.Fixed())
			}
		}
		s.MissingEntries = keep

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
			eBlockRequest.SendOut(s, eBlockRequest)
		}

		i := 0
		// Ask for missing entries.

		for _, v := range s.MissingEntries {

			if i > 2000 {
				break
			}

			entryTrack := InPlay[v.entryhash.Fixed()]

			if entryTrack == nil || now.Unix()-entryTrack.lastRequest.Unix() > 15 {
				i++
				entryRequest := messages.NewMissingData(s, v.entryhash)
				entryRequest.SendOut(s, entryRequest)
				s.MissingEntryMutex.Unlock()
				time.Sleep(40 * time.Millisecond)
				s.MissingEntryMutex.Lock()
				if entryTrack == nil {
					entryTrack = new(EntryTrack)
					InPlay[v.entryhash.Fixed()] = entryTrack
				}
				entryTrack.lastRequest = now
				entryTrack.cnt++
				if entryTrack.cnt > 20 {
					fmt.Printf("***es Can't get Entry Block %x Entry %x \n", v.ebhash.Bytes(), v.entryhash.Bytes())
				}
			}
		}

		min := 10000000
		max := 0
		maxcnt := 0
		for _, v := range s.MissingEntries {
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

		fmt.Printf("***es Looking for %6d Found %6d In Play %6d Min Height %8d Max Height %8d Max Send: %3d \n",
			len(s.MissingEntries),
			found,
			len(InPlay),
			min,
			max,
			maxcnt)

		s.MissingEntryMutex.Unlock()
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

func (s *State) syncEntries() {

	scan := s.EntryDBHeightProcessing
	alldone := true

	for scan <= s.GetHighestSavedBlk() && len(s.MissingEntries) < 10000 {

		db := s.GetDirectoryBlockByHeight(scan)

		if db == nil {
			return
		}

		for _, ebKeyMR := range db.GetEntryHashes()[3:] {
			// The first three entries (0,1,2) in every directory block are blocks we already have by
			// definition.  If we decide to not have Factoid blocks or Entry Credit blocks in some cases,
			// then this assumption might not hold.  But it does for now.

			eBlock, _ := s.DB.FetchEBlock(ebKeyMR)

			// Dont have an eBlock?  Huh. We can go on, but we can't advance
			if eBlock == nil {
				alldone = false
				continue
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
					// Something missing. stop moving the bookmark.
					alldone = false
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
			s.EntryDBHeightProcessing = scan + 1
		}
		scan++
	}
}

// This routine walks through the directory blocks, looking for missing entry blocks and entries.
// Once it finds something missing, it keeps that as a mark, and starts there the next time it is
// called.

func (s *State) CatchupEBlocks() {

	for {
		s.MissingEntryMutex.Lock()
		s.syncEntryBlocks()
		s.syncEntries()
		s.MissingEntryMutex.Unlock()
		time.Sleep(5 * time.Second)
	}
}
