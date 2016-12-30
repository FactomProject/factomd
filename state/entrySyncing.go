// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

// Set timers and send requests
func (s *State) setEntryTimers() {

	now := s.GetTimestamp()

	// If we have no Entry Blocks in our queue, reset our timer.
	if len(s.MissingEntryBlocks) == 0 || s.MissingEntryBlockRepeat == nil {
		s.MissingEntryBlockRepeat = now
		return
	}

	// If our delay has been reached, then ask for some missing Entry blocks
	// This is a replay, because sometimes requests are ignored or lost.
	if now.GetTimeSeconds()-s.MissingEntryBlockRepeat.GetTimeSeconds() > 5 {
		s.MissingEntryBlockRepeat = now

		for _, eb := range s.MissingEntryBlocks {
			eBlockRequest := messages.NewMissingData(s, eb.ebhash)
			eBlockRequest.SendOut(s, eBlockRequest)
		}
	}

	if len(s.MissingEntries) == 0 {
		s.MissingEntryRepeat = nil
		s.EntryDBHeightComplete = s.EntryBlockDBHeightComplete
		s.EntryHeightComplete = s.EntryHeightComplete
	} else {
		if s.MissingEntryRepeat == nil {
			s.MissingEntryRepeat = now
		}

		// If our delay has been reached, then ask for some missing Entry blocks
		// This is a replay, because sometimes requests are ignored or lost.
		if now.GetTimeSeconds()-s.MissingEntryRepeat.GetTimeSeconds() > 5 {
			s.MissingEntryRepeat = now

			for i, eb := range s.MissingEntries {
				if i > 200 {
					// Only send out 200 requests at a time.
					break
				}
				entryRequest := messages.NewMissingData(s, eb.entryhash)
				entryRequest.SendOut(s, entryRequest)
			}
		}
	}
}

func (s *State) askForEntries(db interfaces.IDirectoryBlock, ebKeyMR interfaces.IHash, eBlock interfaces.IEntryBlock) (alldone bool) {
	for _, entryhash := range eBlock.GetEntryHashes() {
		if entryhash.IsMinuteMarker() {
			continue
		}
		e, _ := s.DB.FetchEntry(entryhash)
		if e == nil {
			//Check lists and not add if already there.
			addit := false
			for _, e := range s.MissingEntries {
				if e.ebhash.Fixed() == entryhash.Fixed() {
					addit = false
				}
				return
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
	return
}

func (s *State) checkEntryBlock(db interfaces.IDirectoryBlock, ebKeyMR interfaces.IHash) (alldone bool) {
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
		return
	}

	return alldone && s.askForEntries(db, ebKeyMR, eBlock)

}

// This routine walks through the directory blocks, looking for missing entry blocks and entries.
// Once it finds something missing, it keeps that as a mark, and starts there the next time it is
// called.

func (s *State) catchupEBlocks() {

	s.setEntryTimers()

	// If we still have 5 entry blocks that we are asking for, then let's not add to the list.
	if len(s.MissingEntryBlocks) > 5 {
		return
	}

	// All done is true, and as long as it says true, we walk our bookmark forward.  Once we find something
	// missing, we stop moving the bookmark, and rely on caching to keep us from thrashing the disk as we
	// review the directory block over again the next time.
	alldone := true
	for s.EntryBlockDBHeightProcessing < s.GetHighestCompletedBlk() && len(s.MissingEntryBlocks) < 10 {
		dbstate := s.DBStates.Get(int(s.EntryBlockDBHeightProcessing))

		if dbstate == nil {
			break
		}
		db := s.GetDirectoryBlockByHeight(s.EntryBlockDBHeightProcessing)

		for i, ebKeyMR := range db.GetEntryHashes() {
			// The first three entries (0,1,2) in every directory block are blocks we already have by
			// definition.  If we decide to not have Factoid blocks or Entry Credit blocks in some cases,
			// then this assumption might not hold.  But it does for now.
			if i <= 2 {
				continue
			}

			s.checkEntryBlock(db, ebKeyMR)

		}

		if alldone {
			// we had three bookmarks.  Now they are all in lockstep. TODO: get rid of extra bookmarks.
			s.EntryBlockDBHeightComplete++
			s.EntryDBHeightComplete++
			s.EntryHeightComplete++
			s.EntryBlockDBHeightProcessing++
		}
	}

}
