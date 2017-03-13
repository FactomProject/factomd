// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"math"
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

	lastFeedback := 0
	startt := time.Now()

	secs := func() int {
		return int(time.Now().Unix() - startt.Unix())
	}

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
		newrequest := 0
		// Make a copy of Missing Entries while locked down, and do all the Entry Block syncing (because mostly
		// we are not doing any engry block syncing because they are synced with DBStates, but we do so for
		// sanity purposes.

		var missingeb []MissingEntryBlock

		s.MissingEntryMutex.Lock()
		missingeb = append(missingeb, s.MissingEntryBlocks...)
		s.MissingEntryMutex.Unlock()

		{
			// Remove all entry blocks we have already found and recorded
			var keepeb []MissingEntryBlock
			for _, v := range missingeb {
				eb, _ := s.DB.FetchEBlock(v.ebhash)
				if eb == nil {
					keepeb = append(keepeb, v)
				}
			}
			s.MissingEntryMutex.Lock()
			s.MissingEntryBlocks = keepeb
			s.MissingEntryMutex.Unlock()

			// Ask for missing entry blocks
			for i, v := range s.MissingEntryBlocks {

				if i > 50 {
					break
				}

				eBlockRequest := messages.NewMissingData(s, v.ebhash)
				eBlockRequest.SendOut(s, eBlockRequest)
			}
		}

		var keep []MissingEntry

		// Every call to update recomputes the keep array from the s.MissingEntries, and s.MissingEntries
		update := func() {
			var missinge []MissingEntry
			keep = make([]MissingEntry, 0)

			s.MissingEntryMutex.Lock()
			missinge = append(missinge, s.MissingEntries...)
			s.MissingEntryMutex.Unlock()

			// Remove all Entries that we have already found and recorded. "keep" the ones we are still looking for.
			for _, v := range missinge {

				if !has(s, v.entryhash) {
					keep = append(keep, v)
				} else {
					found++
					newfound++
					delete(InPlay, v.entryhash.Fixed())
				}
			}
			delay := 1000 - newfound
			if delay > 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
			// Okay, so at this point we have what entries in the missingEntries we should keep, as of when we
			// entered this routine.  But the scanner might have found some more missing entries.  They are on the
			// end.  So we will get those in newStuff, and append that to the end of what we figured out we should
			// keep.
			//
			// This is assumed the only code that removes missingEntries.  If any other code removes from the
			// MissingEntries slice, it is going to mess up everything.

			s.MissingEntryMutex.Lock()
			// Let the outside world know which entries we are looking for.
			newStuff := []MissingEntry{}
			newStuff = append(newStuff, s.MissingEntries[len(missinge):]...)
			var newme []MissingEntry
			keep = append(keep, newStuff...)
			s.MissingEntries = append(newme, keep...)
			s.MissingEntryMutex.Unlock()
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

			if min != math.MaxInt32 && min > 0 && min > int(s.EntryDBHeightComplete) {
				s.EntryDBHeightComplete = uint32(min - 1)
			}

			if true {
				foundstr := fmt.Sprint(newrequest, "/", newfound, "/", found)
				newfound = 0
				newrequest = 0
				mmin := min
				if mmin == math.MaxInt32 {
					mmin = 0
				}
				zsecs := secs()
				ss := zsecs % 60       // seconds
				m := (zsecs / 60) % 60 // minutes
				h := (zsecs / 60 / 60)
				fmt.Printf("***es %s"+
					" time %2d:%02d:%02d"+
					" #missing: %4d"+
					" Req/New/Found: %15s"+
					" In Play: %4d"+
					" Min Height: %6d "+
					" Max Height: %6d "+
					" Avg Send: %d.%03d"+
					" Sending: %4d"+
					" Max Send: %2d"+
					" Highest Saved %6d "+
					" Entry complete %6d\n",
					s.FactomNodeName,
					h, m, ss,
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
		}

		for i, v := range keep {

			if i > 2000 {
				break
			}

			thesecs := secs()
			if thesecs-lastFeedback > 3 {
				update()
				feedback()
				lastFeedback = thesecs
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
				newrequest++
				if len(InPlay) > 20 {
					time.Sleep(time.Duration(len(InPlay)/20) * time.Millisecond)
				}
				et.lastRequest = now
				et.cnt++
				if et.cnt%25 == 25 {
					fmt.Printf("***es Can't get Entry Block %x Entry %x in %v attempts.\n", v.ebhash.Bytes(), v.entryhash.Bytes(), et.cnt)
				}
			}
		}

		update()
		feedback()

		if len(keep) == 0 {
			time.Sleep(300 * time.Millisecond)
		}
		// slow down as the number of retries per message goes up
		time.Sleep(time.Duration((avg - 1000)) * time.Millisecond)

	}
}

func (s *State) GoSyncEntryBlocks() {

	for {

		// All done is true, and as long as it says true, we walk our bookmark forward.  Once we find something
		// missing, we stop moving the bookmark, and rely on caching to keep us from thrashing the disk as we
		// review the directory block over again the next time.
		alldone := true

		var keep []MissingEntryBlock
		var missingeb []MissingEntryBlock

		s.MissingEntryMutex.Lock()
		missingeb = append(missingeb, s.MissingEntryBlocks...)
		s.MissingEntryMutex.Unlock()

		for s.EntryBlockDBHeightProcessing <= s.GetHighestSavedBlk() && len(missingeb) < 1000 {

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
					for _, eb := range missingeb {
						if eb.ebhash.Fixed() == ebKeyMR.Fixed() {
							addit = false
							break
						}
					}

					if addit {
						missingEntryBlock := MissingEntryBlock{ebhash: ebKeyMR, dbheight: s.EntryBlockDBHeightProcessing}
						keep = append(keep, missingEntryBlock)
					}
					// Something missing, stop moving the bookmark.
					alldone = false
					continue
				}
			}

			s.MissingEntryMutex.Lock()
			s.MissingEntryBlocks = keep
			s.MissingEntryMutex.Unlock()

			if alldone {
				// we had three bookmarks.  Now they are all in lockstep. TODO: get rid of extra bookmarks.
				s.EntryBlockDBHeightComplete++
				s.EntryBlockDBHeightProcessing++
			}
		}
		s.MissingEntryMutex.Lock()
		t := len(s.MissingEntryBlocks)
		s.MissingEntryMutex.Unlock()
		if t <= 3 {
			time.Sleep(3 * time.Second)
		}
		time.Sleep(time.Duration(t/10) * time.Millisecond)

	}
}

func (s *State) GoWriteEntries() {
	for {
		time.Sleep(300 * time.Millisecond)

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

func (s *State) GoSyncEntries() {

	go s.GoWriteEntries() // Start a go routine to write the Entries to the DB
	go s.GoSyncEntryBlocks()

	starting := uint32(0)

	for {

		scan := uint32(starting)
		var missinge []MissingEntry
		var newentries []MissingEntry
		alldone := true
		s.MissingEntryMutex.Lock()
		missinge = append(missinge, s.MissingEntries...)
		s.MissingEntryMutex.Unlock()
	scanentries:
		for scan <= s.GetHighestSavedBlk() {

			db := s.GetDirectoryBlockByHeight(scan)

			if db == nil {
				break
			}

			for _, ebKeyMR := range db.GetEntryHashes()[3:] {
				// The first three entries (0,1,2) in every directory block are blocks we already have by
				// definition.  If we decide to not have Factoid blocks or Entry Credit blocks in some cases,
				// then this assumption might not hold.  But it does for now.

				eBlock, _ := s.DB.FetchEBlock(ebKeyMR)

				// Dont have an eBlock?  Huh. We can go on, but we can't advance
				if eBlock == nil {
					alldone = false
					break scanentries
				}

				if len(newentries)+len(missinge) > 1000 {
					alldone = false
					break scanentries
				}

				for _, entryhash := range eBlock.GetEntryHashes() {

					// slow down if we are behind.
					time.Sleep(time.Duration(len(s.WriteEntry)/10) * time.Millisecond)

					if entryhash.IsMinuteMarker() {
						continue
					}

					// If I have the entry, then remove it from the Missing Entries list.
					if has(s, entryhash) {
						// If I am missing the entry, add it to th eMissing Entries list
					} else {
						//Check lists and not add if already there.
						addit := true
						for _, e := range missinge {
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
							newentries = append(newentries, v)
						}
					}
					ueh := new(EntryUpdate)
					ueh.Hash = entryhash
					ueh.Timestamp = db.GetTimestamp()
					s.UpdateEntryHash <- ueh
				}

			}

			s.MissingEntryMutex.Lock()

			if alldone && len(s.MissingEntries) == 0 && len(newentries) == 0 {
				starting = scan
			}
			scan++
			s.MissingEntryMutex.Unlock()
		}

		zerolen := false
		s.MissingEntryMutex.Lock()
		s.MissingEntries = append(s.MissingEntries, newentries...)
		zerolen = len(s.MissingEntries) == 0

		if zerolen {
			s.EntryDBHeightComplete = s.GetHighestSavedBlk()
			starting = s.GetHighestSavedBlk()
		}
		s.MissingEntryMutex.Unlock()

		if zerolen {
			time.Sleep(300 * time.Millisecond)
		}

		// sleep some time no matter what.
		time.Sleep(200 * time.Millisecond)

	}
}
