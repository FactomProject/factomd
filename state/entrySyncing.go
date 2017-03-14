// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
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

		s.MissingEntryMutex.Lock()
		for k := range s.MissingEntryMap {
			if has(s, s.MissingEntryMap[k].EntryHash) {
				delete(s.MissingEntryMap, k)
			}
		}
		s.MissingEntryMutex.Unlock()

		// Ask for missing entries.

		min := int(math.MaxInt32)
		max := 0
		maxcnt := 0
		avg := 0
		cnt := 0

		feedback := func() {
			s.MissingEntryMutex.Lock()
			defer s.MissingEntryMutex.Unlock()
			min = int(math.MaxInt32)
			max = 0
			maxcnt = 0
			sum := 0
			cnt = 0
			for k := range s.MissingEntryMap {
				v := s.MissingEntryMap[k]
				if min > int(v.DBHeight) {
					min = int(v.DBHeight)
				}
				if max < int(v.DBHeight) {
					max = int(v.DBHeight)
				}

				if maxcnt < v.Cnt {
					maxcnt = v.Cnt
				}

				if v.Cnt > 0 {
					cnt++
					sum += v.Cnt
				}
			}

			avg = 0
			if cnt > 0 {
				avg = (1000 * sum) / cnt
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
					len(s.MissingEntryMap),
					foundstr,
					len(s.MissingEntryMap),
					mmin,
					max,
					avg/1000, avg%1000,
					cnt,
					maxcnt,
					s.GetHighestSavedBlk(),
					s.EntryDBHeightComplete)
			}
		}

		s.MissingEntryMutex.Lock()
		for len(s.MissingEntryMap) < 200 {
			select {
			case et := <-s.MissingEntries:
				s.MissingEntryMap[et.EntryHash.Fixed()] = &et
			default:
				break
			}
		}
		s.MissingEntryMutex.Unlock()

		Resend := func() {
			s.MissingEntryMutex.Lock()
			for k := range s.MissingEntryMap {
				et := s.MissingEntryMap[k]
				s.MissingEntryMutex.Unlock()

				thesecs := secs()
				if thesecs-lastFeedback > 3 {
					feedback()
					lastFeedback = thesecs
				}

				if et.Cnt == 0 || now.Unix()-et.LastTime.Unix() > 40 {
					entryRequest := messages.NewMissingData(s, et.EntryHash)
					entryRequest.SendOut(s, entryRequest)
					newrequest++
					time.Sleep(10 * time.Millisecond)
					et.LastTime = now
					et.Cnt++
					if et.Cnt%25 == 25 {
						fmt.Printf("***es Can't get Entry Block %x Entry %x in %v attempts.\n", et.EBHash.Bytes(), et.EntryHash.Bytes(), et.Cnt)
					}
				}
				s.MissingEntryMutex.Lock()
			}
			s.MissingEntryMutex.Unlock()
		}

		Resend()
		feedback()
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

				s.DB.InsertEntry(entry)

				s.MissingEntryMutex.Lock()
				asked := s.MissingEntryMap[entry.GetHash().Fixed()] != nil
				s.MissingEntryMutex.Unlock()

				if asked {
					s.DB.InsertEntry(entry)
					s.MissingEntryMutex.Lock()
					delete(s.MissingEntryMap, entry.GetHash().Fixed())
					s.MissingEntryMutex.Unlock()
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

	start := uint32(0)
	for {

		for scan := start; scan < s.GetHighestSavedBlk(); scan++ {

			db := s.GetDirectoryBlockByHeight(scan)

			for db == nil {
				time.Sleep(1 * time.Second)
				db = s.GetDirectoryBlockByHeight(scan)
			}

			for _, ebKeyMR := range db.GetEntryHashes()[3:] {
				// The first three entries (0,1,2) in every directory block are blocks we already have by
				// definition.  If we decide to not have Factoid blocks or Entry Credit blocks in some cases,
				// then this assumption might not hold.  But it does for now.

				eBlock, _ := s.DB.FetchEBlock(ebKeyMR)

				// Dont have an eBlock?  Huh. We can go on, but we can't advance
				for eBlock == nil {
					eBlock, _ = s.DB.FetchEBlock(ebKeyMR)
					time.Sleep(1 * time.Second)
				}

				for _, entryhash := range eBlock.GetEntryHashes() {

					if entryhash.IsMinuteMarker() {
						continue
					}

					// If I have the entry, then remove it from the Missing Entries list.
					if !has(s, entryhash) {

						var v MissingEntry

						v.DBHeight = eBlock.GetHeader().GetDBHeight()
						v.EntryHash = entryhash
						v.EBHash = ebKeyMR
						s.MissingEntries <- v
					}
					ueh := new(EntryUpdate)
					ueh.Hash = entryhash
					ueh.Timestamp = db.GetTimestamp()
					s.UpdateEntryHash <- ueh
				}

			}
			start = scan
		}
		// sleep some time no matter what.
		time.Sleep(200 * time.Millisecond)
	}
}
