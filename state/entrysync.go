package state

import (
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"go.uber.org/ratelimit"
)

// TODO rename
type EntrySyncNew struct {
	s     *State
	limit ratelimit.Limiter

	closer chan interface{}
}

func NewEntrySync(s *State, requestsPerSecond int) *EntrySyncNew {
	es := new(EntrySyncNew)
	es.s = s
	es.limit = ratelimit.New(requestsPerSecond)
	return es
}

func (es *EntrySyncNew) Start() {
	select {
	case <-es.closer:
	default:
		panic("EntrySync already running")
	}
	es.closer = make(chan interface{}, 1)
	go es.syncHeight()
	go es.processReplies()
}

func (es *EntrySyncNew) syncMax() uint32 {
	end := es.s.GetHighestSavedBlk()
	if es.s.DBStates.ProcessHeight < end {
		end = es.s.DBStates.ProcessHeight
	}
	return end
}

func (es *EntrySyncNew) syncHeight() {
	position := es.s.EntryDBHeightComplete + 1

	for {
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

		for _, keymr := range db.GetEntryHashes() {
			for !es.syncEBlock(keymr, db.GetTimestamp()) {
				time.Sleep(time.Second)
			}
		}
	}
}

func (es *EntrySyncNew) syncEBlock(keymr interfaces.IHash, ts interfaces.Timestamp) bool {
	eblock, err := es.s.DB.FetchEBlock(keymr)
	if err != nil { // database corrupt
		panic(err)
	}

	if eblock == nil {
		return false
	}

	for _, entryHash := range eblock.GetEntryHashes() {
		if entryHash.IsMinuteMarker() {
			continue
		}

		// see state.UpdateState()
		update := new(EntryUpdate)
		update.Hash = entryHash
		update.Timestamp = ts
		es.s.UpdateEntryHash <- update

		es.syncEntryHash(entryHash)
	}

	return true
}

func (es *EntrySyncNew) syncEntryHash(hash interfaces.IHash) {

}

func (es *EntrySyncNew) processReplies() {

}
