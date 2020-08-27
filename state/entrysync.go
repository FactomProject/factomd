package state

import (
	"sync"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
)

const (
	EntrySyncMax     = 750 // good-ish value
	EntrySyncRetry   = time.Second * 10
	EntrySyncWarning = 16
)

// TODO comment
type EntrySync struct {
	s *State

	askMtx sync.RWMutex
	askMap map[[32]byte]bool // entry hash => ask
	asks   []*entrySyncAsk   // chronological

	ebMtx   sync.Mutex
	eblocks []*entrySyncEBlock

	closer chan interface{}
}

type entrySyncAsk struct {
	hash  interfaces.IHash
	time  time.Time
	tries int
}

type entrySyncEBlock struct {
	height  uint32
	missing []interfaces.IHash
}

func NewEntrySync(s *State) *EntrySync {
	es := new(EntrySync)
	es.s = s
	es.closer = make(chan interface{}, 1)
	es.askMap = make(map[[32]byte]bool)
	return es
}

func (es *EntrySync) Stop() {
	close(es.closer)
}

func (es *EntrySync) check() {
	for {
		select {
		case <-es.closer:
			return
		default:
		}

		if len(es.eblocks) == 0 { // outside of mutex but should be fine
			time.Sleep(time.Second)
			continue
		}

		es.ebMtx.Lock()
		eb := es.eblocks[0]

		stillmissing := make([]interfaces.IHash, 0, len(eb.missing))
		for _, entryhash := range eb.missing {
			if !es.has(entryhash) {
				stillmissing = append(stillmissing, entryhash)
			}
		}

		if len(stillmissing) > 0 {
			eb.missing = stillmissing
		} else { // eblock is complete
			es.eblocks = es.eblocks[1:]
			es.s.SetEntryBlockDBHeightComplete(eb.height)
		}
		es.ebMtx.Unlock()

		if len(stillmissing) > 0 { // waiting on requests
			time.Sleep(time.Second)
		}
	}
}

func (es *EntrySync) has(hash interfaces.IHash) bool {
	if has, err := es.s.DB.DoesKeyExist(databaseOverlay.ENTRY, hash.Bytes()); err != nil {
		panic(err)
	} else {
		return has
	}
}

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

func (es *EntrySync) syncMax() uint32 {
	end := es.s.GetHighestSavedBlk()
	if es.s.DBStates.ProcessHeight < end {
		end = es.s.DBStates.ProcessHeight
	}
	return end
}

func (es *EntrySync) SyncHeight() {
	go es.ask()
	go es.check()

	position := es.s.EntryDBHeightComplete + 1
	for {
		select {
		case <-es.closer:
			return
		default:
		}

		// nothing to do
		if position == es.syncMax() {
			time.Sleep(time.Second)
			continue
		}

		db := es.s.GetDirectoryBlockByHeight(position)
		if db == nil { // block not saved yet
			time.Sleep(time.Second)
			continue
		}

		if es.s.DBFinished { // throttle syncing
			time.Sleep(time.Millisecond * 125)
		}

		for _, keymr := range db.GetEntryHashes()[3:] { // skip the first 3
			for !es.syncEBlock(position, keymr, db.GetTimestamp()) {
				time.Sleep(time.Second)
			}
		}

		es.s.SetEntryBlockDBHeightProcessing(position)
		position++
	}
}

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

	es.ebMtx.Lock()
	es.eblocks = append(es.eblocks, ebsync)
	es.ebMtx.Unlock()

	return true
}

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

func (es *EntrySync) AskedFor(hash interfaces.IHash) bool {
	es.askMtx.RLock()
	defer es.askMtx.RUnlock()
	return es.askMap[hash.Fixed()]
}
