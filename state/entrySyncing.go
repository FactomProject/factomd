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

type ReCheck struct {
	TimeToCheck int64            //Time in seconds to recheck
	EntryHash   interfaces.IHash //Entry Hash to check
	DBHeight    int
	NumEntries  int
	Tries       int
}

type EntrySync struct {
	MissingDBlockEntries chan []*ReCheck // We don't have these entries.  Each list is from a directory block.
	DBHeightBase         int             // This is the highest block with entries not yet checked or are missing
	TotalEntries         int             // Total Entries in the database

}

// Maintain queues of what we want to test, and what we are currently testing.
func (es *EntrySync) Init() {
	es.MissingDBlockEntries = make(chan []*ReCheck, 10000) // Check 10 directory blocks at a time.
} // we have to reprocess

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
		if !has(s, entry.GetHash()) {
			s.DB.StartMultiBatch()
			err := s.DB.InsertEntryMultiBatch(entry)
			if err != nil {
				panic(err)
			}
			err = s.DB.ExecuteMultiBatch()
			if err != nil {
				panic(err)
			}
		}
	}
}

// RequestAndCollectMissingEntries()
// We were missing these entries.  Check to see if we have them yet.  If we don't then schedule to recheck.
func (s *State) RequestAndCollectMissingEntries() {
	es := s.EntrySyncState

	// We control delay to adjust to network and deployment conditions to keep the request rate reasonable.
	var delay = int64(400000000) // Start with 400 milliseconds

	var rErrorSum float64
	var LastError float64
	var avgTries float64 = 1
	var iterations float64
	const (
		outstanding = 50 // Number of outstanding entries (more or less) per evaluation.  Could be an entire block though
		target      = 3.0
		Kp          = 2.0
		Ki          = 5.0
		Kd          = 10.0
	)
	var lastDelay int64
	for {
		missing := true
		dbrcs := <-es.MissingDBlockEntries
		dbht := 0
	get10:
		for i := 0; i < 50; { // 1000+ entries
			select {
			case dbrcs2 := <-es.MissingDBlockEntries:
				dbrcs = append(dbrcs, dbrcs2...)
				dbht = dbrcs2[0].DBHeight // keep the directory hieight of the last directory block in the set
				i += dbrcs2[0].NumEntries // Add the entries in this directory block
			default:
				break get10
			}
		}
		pass := 0
		for missing {
			sumTries := 0
			missing = false
			found := int64(0)
			total := int64(len(dbrcs))
			pass++
			for i, rc := range dbrcs {
				if rc == nil {
					continue
				}
				iterations++

				d := delay / ((total - found) + 1)
				if lastDelay != d {
					//	s.LogPrintf("entrysyncing", "Delay %d", d)
					lastDelay = d
				}
				if d < 1000000 {
					d = 1000000
				}
				time.Sleep(time.Duration(d))

				if !has(s, rc.EntryHash) {
					rc.Tries++
					entryRequest := messages.NewMissingData(s, rc.EntryHash)
					entryRequest.SendOut(s, entryRequest)

					missing = true
				} else {
					if rc.Tries == 0 {
						total--
					}
					found++
					dbrcs[i] = nil
					sumTries += rc.Tries
					s.LogPrintf("entrysyncing", "%20s %x dbht %8d found %6d/%6d tries %6d avg %8f delay %d ms QueueLen: %d",
						"Found Entry",
						rc.EntryHash.Bytes()[:6],
						rc.DBHeight,
						found,
						total,
						rc.Tries,
						avgTries,
						delay/1000000,
						len(es.MissingDBlockEntries))
				}
			}
			// Pid Controller to adjust the time I spend between requests so we have just a bit
			// above one request per entry recieved.  Because requests can be lost, and because
			// one request is a hard lower bound (We have to do that, and we can't do less, and
			// any large wait will give us that.)
			//
			// So we will aim for 2 requests per entry, and control against the decaying average requests per entry
			//
			// Time is what we are changing, so we are going to consider our iterations to be time.
			//
			// Standard PID controllers have three components, error, integral of error, and derivitave of error
			// spin.atomicobject.com/2016/06/28/intro-pid-control/

			// Compute the new average Request Rate, which is the decaying average the rate of requests
			if found > 0 {

				navgTries := float64(sumTries) / float64(found)

				// Okay, so we are doing control over a set of measurements.  But if we only have a small number,
				// they are more chaotic than with larger samples.  Plus, just a few measurements shouldn't have
				// the same weight as a bunch of measurements.  So yeah, I should be able to balance it with math,
				// but I used a switch statement that pretty much does the same thing.
				switch {
				case found < 4:
					navgTries = (16*avgTries + navgTries) / 17
				case found < 16:
					navgTries = (8*avgTries + navgTries) / 9
				case found < 64:
					navgTries = (avgTries + 4*navgTries) / 5
				case found < 256:
					navgTries = (avgTries + 16*navgTries) / 17
				default:
					navgTries = (avgTries + 32*navgTries) / 31
				}

				rError := 2.0 - navgTries
				rErrorSum += rError * float64(found)
				derivative_of_error := (rError - LastError) / float64(found)
				LastError = rError

				// The output comes out negative.  I don't fight that, I just flip the sign.  Otherwise, its a
				// standard PID weighting:
				out := (rError*Kp + rErrorSum*Ki + derivative_of_error*Kd) * -1

				// If the output still comes out negative, that doesn't work for a delay.  Floor at zero.  This is also
				// why we target some number other than  1.  We always get at least 1 request per found entry, so we
				// can only move up towards a value that has some room underneath.  If I came up with another way to
				// measure error other than just simple X requests to get 1 entry, then maybe I could do something
				// better.
				if out < 0 {
					out = 0
				}

				// Easier to think in milliseconds and target milliseconds.  But we need nanoseconds for delay.
				// converting here.
				ndelay := int64(out * 1000000) // Convert the output of our control to milliseconds for the next step

				// Logging is fun, helps to tune the constants (Kp Ki, and Kd).  However, what I have seems to work.
				s.LogPrintf("entrysyncing", "PID Kp %8.4f Ki %8.4f Kd %8.4f ::: rError*Kp %f8.4 rErrorSum*Ki %8.4f derivitive_of_error %8.4f out %8.4f",
					Kp, Ki, Kd,
					rError, rErrorSum, derivative_of_error, out)
				s.LogPrintf("entrysyncing", "Delay %10d->%10d avgTries %8.4f->%8.4f",
					delay,
					ndelay,
					avgTries,
					navgTries)

				// Update variables I need for the next round of work.
				avgTries = navgTries
				delay = ndelay
			}
		}
		s.LogPrintf("entrysyncing", "%20s dbht %d",
			"Found Entry", dbht)
		s.EntryDBHeightComplete = uint32(dbht)
		s.EntryBlockDBHeightComplete = uint32(dbht)
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
	go s.RequestAndCollectMissingEntries()

	highestChecked := s.EntryDBHeightComplete
	lookingfor := 0
	for {

		if !s.DBFinished {
			time.Sleep(time.Second / 30)
		}

		highestSaved := s.GetHighestSavedBlk()

		somethingMissing := false
		for scan := highestChecked + 1; scan <= highestSaved; scan++ {
			// Okay, stuff we pull from wherever but there is nothing missing, then update our variables.
			if !somethingMissing && scan > 0 && s.EntryDBHeightComplete < scan-1 {
				s.EntryBlockDBHeightComplete = scan - 1
				s.EntryDBHeightComplete = scan - 1
				s.EntrySyncState.DBHeightBase = int(scan) // The base is the height of the block that might have something missing.
				if scan%100 == 0 {
					//	s.LogPrintf("entrysyncing", "DBHeight Complete %d", scan-1)
				}
			}

			s.EntryBlockDBHeightProcessing = scan
			s.EntryDBHeightProcessing = scan

			db := s.GetDirectoryBlockByHeight(scan)

			// Wait for the database if we have to
			for db == nil {
				time.Sleep(1 * time.Second)
				db = s.GetDirectoryBlockByHeight(scan)
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
						somethingMissing = true
					}
				}
			}
			for cap(s.EntrySyncState.MissingDBlockEntries) < len(s.EntrySyncState.MissingDBlockEntries)+cap(s.EntrySyncState.MissingDBlockEntries)/1000 {
				time.Sleep(time.Second)
			}

			lookingfor += len(entries)

			if len(entries) > 0 {
				//	s.LogPrintf("entrysyncing", "Missing entries total %10d at height %10d directory entries: %10d QueueLen %10d",
				//		lookingfor, scan, len(entries), len(s.EntrySyncState.MissingDBlockEntries))
				var rcs []*ReCheck
				for _, entryHash := range entries {
					rc := new(ReCheck)
					rc.EntryHash = entryHash
					rc.TimeToCheck = time.Now().Unix() + int64(s.DirectoryBlockInSeconds/100) // Don't check again for seconds
					rc.DBHeight = int(scan)
					rc.NumEntries = len(entries)
					rcs = append(rcs, rc)
				}
				s.EntrySyncState.MissingDBlockEntries <- rcs
			}
		}
		highestChecked = highestSaved
	}
}
