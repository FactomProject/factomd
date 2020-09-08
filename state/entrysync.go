package state

import (
	"sync"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
)

var (
	// EntrySyncMax is the maximum amount of entries to request concurrently
	EntrySyncMax = 750 // good-ish value
	// EntrySyncMaxEBlocks is the maximum amount of eblocks to process concurrently.
	// The value only applies if that range of eblocks has fewer than EntrySyncMax entries.
	EntrySyncMaxEBlocks = 200
	// EntrySyncRetry dictates after what period to retry an unanswered request
	EntrySyncRetry = time.Second * 10
	// EntrySyncWarning is the number of failed requests before warning in the console.
	EntrySyncWarning = 16
)

// EntrySync is responsible for sending the requests to fetch missing entries to the network.
// (aka 2nd pass sync)
// The general algorithms are as follows:
//
// Sync:
// Start with the last known complete height
// 1. Load the DBlock and iterate over all EBlock hashes
// 2. For each EBlock, iterate over the Entry hashes
// 3. If the DB has the entry, skip, otherwise add the hash to the ask queue
// 4. Collate all missing entry hashes for each EBlock and add it to eblock queue
// 5. Inform state we are processing up to that eblock
//
// Ask:
// 1. Look at the first N records of the ask queue
// 2. If the entry exists, remove it from queue
// 3. If it's the first time or time threshold passed, send a request to the network
//
// Check:
// 1. Look at the first record of the eblock queue
// 2. Go through all missing records and check if they exist
// 3. If there are still missing records, wait and go back to 1
// 4. If the eblock is finished, inform the state we have all the entries up to that eblock
type EntrySync struct {
	s *State

	position uint32

	askMtx sync.RWMutex
	askMap map[[32]byte]bool // entry hash => ask
	asks   []*entrySyncAsk   // chronological

	eblocks chan *entrySyncEBlock

	closer chan interface{}
}

// an item in the ask queue
type entrySyncAsk struct {
	hash  interfaces.IHash
	time  time.Time
	tries int
}

// an item in the eblock queue
type entrySyncEBlock struct {
	height  uint32
	missing []interfaces.IHash
}

// NewEntrySync creates a new EntrySync
func NewEntrySync(s *State) *EntrySync {
	es := new(EntrySync)
	es.s = s
	es.closer = make(chan interface{}, 1)
	es.askMap = make(map[[32]byte]bool)
	es.eblocks = make(chan *entrySyncEBlock, EntrySyncMaxEBlocks)
	return es
}

// Stop initiates shutdown and will stop all associated goroutines
func (es *EntrySync) Stop() {
	close(es.closer)
	close(es.eblocks)
}

// check routine, for description see EntrySync comment
func (es *EntrySync) check() {
	for eb := range es.eblocks {
		es.s.SetEntryBlockDBHeightProcessing(eb.height)

		for {
			select {
			case <-es.closer:
				return
			default:
			}

			for i := 0; i < len(eb.missing); i++ {
				if es.has(eb.missing[i]) {
					eb.missing[i] = eb.missing[len(eb.missing)-1]
					eb.missing = eb.missing[:len(eb.missing)-1]
					i--
				}
			}

			// eblock is complete
			if len(eb.missing) == 0 {
				es.s.SetEntryBlockDBHeightComplete(eb.height)
				break
			}

			time.Sleep(time.Second)
		}
	}
}

// has checks if the hash exists in the database
func (es *EntrySync) has(hash interfaces.IHash) bool {
	if has, err := es.s.DB.DoesKeyExist(databaseOverlay.ENTRY, hash.Bytes()); err != nil {
		panic(err)
	} else {
		return has
	}
}

// ask routine, for description see EntrySync comment
func (es *EntrySync) ask() {
	for {
		select {
		case <-es.closer:
			return
		default:
		}

		es.askMtx.Lock()
		// check the first N entries in the list
		for i := 0; i < len(es.asks) && i < EntrySyncMax; i++ {
			ask := es.asks[i]

			// remove asks that are fulfilled
			if es.has(ask.hash) {
				es.asks = append(es.asks[:i], es.asks[i+1:]...)
				i-- // we removed one, index stays the same
				delete(es.askMap, ask.hash.Fixed())
				continue
			}

			// re-ask
			if time.Since(ask.time) > EntrySyncRetry {
				request := messages.NewMissingData(primitives.NewTimestampNow(), ask.hash)
				request.SendOut(es.s, request)
				ask.time = time.Now()
				ask.tries++

				if ask.tries%EntrySyncWarning == 0 {
					es.s.Logger.WithField("tries", ask.tries).WithField("hash", ask.hash.String()).Warnf("Unable to retrieve entry from network")
				}
			}
		}
		es.askMtx.Unlock()

		time.Sleep(time.Second)
	}
}

// returns the maximum possible height we have eblocks for
func (es *EntrySync) syncMax() uint32 {
	end := es.s.GetHighestSavedBlk()
	if es.s.DBStates.ProcessHeight < end {
		end = es.s.DBStates.ProcessHeight
	}
	return end
}

// DelayedStart waits for the database to be finished loading before starting the syncing process
func (es *EntrySync) DelayedStart() {
	// delay the start of entrysync until
	for !es.s.DBFinished {
		time.Sleep(time.Second)
	}

	es.SyncHeight()
}

// SyncHeight starts the Sync, Ask, and Check routines.
func (es *EntrySync) SyncHeight() {
	go es.ask()
	go es.check()

	es.position = es.s.EntryBlockDBHeightComplete
	// genesis block edge case since EntryDBHeightComplete can't be -1
	if es.position > 0 {
		es.position++
	}

	for {
		select {
		case <-es.closer:
			return
		default:
		}
		// block not available yet
		if es.position > es.syncMax() {
			time.Sleep(time.Second)
			continue
		}

		db := es.s.GetDirectoryBlockByHeight(es.position)
		if db == nil { // block not saved yet
			time.Sleep(time.Second)
			continue
		}

		es.syncDBlock(db)
		es.position++
	}
}

// extract eblocks from the dblock
func (es *EntrySync) syncDBlock(db interfaces.IDirectoryBlock) {
	eblocks := db.GetEntryHashes()[3:] // skip f/c/a-block
	if len(eblocks) > 0 {
		for _, keymr := range eblocks {
			for !es.syncEBlock(db.GetDatabaseHeight(), keymr, db.GetTimestamp()) {
				time.Sleep(time.Second)
			}
		}
	} else {
		es.syncNoEBlock(db.GetDatabaseHeight())
	}
}

// for dblocks with no eblocks, we still need to advance the complete counter
func (es *EntrySync) syncNoEBlock(height uint32) {
	ebsync := new(entrySyncEBlock)
	ebsync.height = height
	es.eblocks <- ebsync
}

// process a single eblock, will call syncEntryHash() on each entry, and add an eblock to the queue
func (es *EntrySync) syncEBlock(height uint32, keymr interfaces.IHash, ts interfaces.Timestamp) bool {
	eblock, err := es.s.DB.FetchEBlock(keymr)
	if err != nil { // database corrupt
		panic(err)
	}

	if eblock == nil {
		return false
	}

	ebsync := new(entrySyncEBlock)
	ebsync.height = height

	for _, entryHash := range eblock.GetEntryHashes() {
		if entryHash.IsMinuteMarker() {
			continue
		}

		// see state.UpdateState()
		update := new(EntryUpdate)
		update.Hash = entryHash
		update.Timestamp = ts
		es.s.UpdateEntryHash <- update

		if es.has(entryHash) {
			continue
		}
		es.syncEntryHash(entryHash)
		ebsync.missing = append(ebsync.missing, entryHash)
	}

	es.eblocks <- ebsync

	return true
}

// process a single entry hash. check for duplicate asks, otherwise add to queue
func (es *EntrySync) syncEntryHash(hash interfaces.IHash) {
	es.askMtx.Lock()
	defer es.askMtx.Unlock()

	if _, ok := es.askMap[hash.Fixed()]; ok {
		return
	}

	ask := new(entrySyncAsk)
	ask.hash = hash
	ask.time = time.Time{} // zero time, never asked

	es.askMap[hash.Fixed()] = true
	es.asks = append(es.asks, ask) // add to end of queue
}

// AskedFor returns true if the entry sync routine asked for a specific hash
// and that entry has not yet been found in the database.
// Returns false otherwise.
func (es *EntrySync) AskedFor(hash interfaces.IHash) bool {
	es.askMtx.RLock()
	defer es.askMtx.RUnlock()
	return es.askMap[hash.Fixed()]
}

// Positions returns the DBlock EntrySync is adding next
func (es *EntrySync) Position() uint32 {
	return es.position
}
