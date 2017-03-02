// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"math"
	"time"
)

var _ = fmt.Print

// This go routine checks every so often to see if we have any missing entries or entry blocks.  It then requests
// them if it finds entries in the missing lists.
func (s *State) MakeMissingEntryRequests() {

	type EntryTrack struct {
		lastRequest time.Time
		dbheight    uint32
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
				eBlockRequest.SendOut(s, eBlockRequest)
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

		for k := range InPlay {
			h := primitives.NewHash(k[:])
			e, _ := s.DB.FetchEntry(h)
			if e != nil {
				delete(InPlay, k)
			}
		}

		// Ask for missing entries.

		min := int(math.MaxInt32)
		max := 0
		maxcnt := 0
		avg := 0
		cnt := 0
		feedback := func() {
			min = int(math.MaxInt32)
			max = 0
			maxcnt = 0
			sum := 0
			cnt = 0
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

				if et != nil && et.cnt > 0 {
					cnt++
					sum += et.cnt
				}
			}

			avg = 0
			if cnt > 0 {
				avg = (1000 * sum) / cnt
			}

			if min != math.MaxInt32 && min > 0 {
				s.EntryDBHeightComplete = uint32(min - 1)
			}

			foundstr := fmt.Sprint(newfound, "/", found)
			newfound = 0
			mmin := min
			if mmin == math.MaxInt32 {
				mmin = 0
			}
			fmt.Printf("***es %s #missing: %d"+
				" NewFound/Found: %13s"+
				" In Play: %6d"+
				" Min Height: %d "+
				" Max Height: %d "+
				" Avg Send: %d.%03d"+
				" Sending: %d"+
				" Max Send: %2d"+
				" Highest Saved %d "+
				" Entry complete %d\n",
				s.FactomNodeName,
				len(keep),
				foundstr,
				len(InPlay),
				mmin,
				max,
				avg/1000, avg%1000,
				cnt,
				maxcnt,
				s.GetHighestSavedBlk(),
				s.EntryDBHeightComplete)
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

			et := InPlay[v.entryhash.Fixed()]

			if et == nil {
				et = new(EntryTrack)
				et.dbheight = v.dbheight
				InPlay[v.entryhash.Fixed()] = et
			}

			if et.cnt == 0 || now.Unix()-et.lastRequest.Unix() > 40 {
				entryRequest := messages.NewMissingData(s, v.entryhash)
				entryRequest.SendOut(s, entryRequest)
				if len(s.WriteEntry) > 2000 {
					time.Sleep(time.Duration(len(s.WriteEntry)/10) * time.Millisecond)
				}
				et.lastRequest = now
				et.cnt++
				if et.cnt%25 == 25 {
					fmt.Printf("***es Can't get Entry Block %x Entry %x in %v attempts.\n", v.ebhash.Bytes(), v.entryhash.Bytes(), et.cnt)
				}
			}
		}

		// slow down as the number of retries per message goes up
		time.Sleep(time.Duration((avg - 1000)) * time.Millisecond)

		update()
		feedback()
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

func (s *State) GoWriteEntries() {
	for {

		time.Sleep(3 * time.Second)

	entryWrite:
		for {

			select {

			case entry := <-s.WriteEntry:

				var missinge []MissingEntry

				s.MissingEntryMutex.Lock()
				missinge = append(missinge, s.MissingEntries...)
				s.MissingEntryMutex.Unlock()

				for _, missing := range missinge {
					e := missing.entryhash

					if e.Fixed() == entry.GetHash().Fixed() {
						s.DB.InsertEntry(entry)
						break
					}
				}

			default:
				break entryWrite
			}
		}
	}
}

func (s *State) SyncEntries() {

	go s.GoWriteEntries()

	for {

		scan := uint32(0)
		alldone := true

		s.MissingEntryMutex.Lock()
		lenEntries := len(s.MissingEntries)
		s.MissingEntryMutex.Unlock()
	scanEntries:
		for scan <= s.GetHighestSavedBlk() && lenEntries < 10000 {

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
						s.MissingEntryMutex.Lock()
						for _, e := range s.MissingEntries {
							if e.entryhash.Fixed() == entryhash.Fixed() {
								addit = false
								break
							}
						}
						s.MissingEntryMutex.Unlock()

						if addit {
							var v MissingEntry

							v.dbheight = eBlock.GetHeader().GetDBHeight()
							v.entryhash = entryhash
							v.ebhash = ebKeyMR
							s.MissingEntryMutex.Lock()
							s.MissingEntries = append(s.MissingEntries, v)
							s.MissingEntryMutex.Unlock()
						}
					}
					ueh := new(EntryUpdate)
					ueh.Hash = entryhash
					ueh.Timestamp = db.GetTimestamp()
					s.UpdateEntryHash <- ueh
				}
			}
			s.MissingEntryMutex.Lock()
			if alldone && len(s.MissingEntries) == 0 {
				s.EntryDBHeightComplete = scan
			}
			scan++
			s.MissingEntryMutex.Unlock()
		}

		s.MissingEntryMutex.Lock()
		if len(s.MissingEntries) == 0 {
			s.EntryDBHeightComplete = s.GetHighestSavedBlk()
		}
		s.MissingEntryMutex.Unlock()

		time.Sleep(time.Duration(len(s.WriteEntry)*20) * time.Millisecond)

		if scan >= s.GetHighestSavedBlk() {
			time.Sleep(5 * time.Second)
		}

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
		time.Sleep(time.Duration(len(s.WriteEntry)/100) * time.Second)
	}
}
