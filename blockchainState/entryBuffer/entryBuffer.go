// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryBuffer

import (
	"sync"
	"time"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
)

var EntryExpirationDuration time.Duration = time.Hour

type StoredData struct {
	Data []byte
	Date time.Time
}

type EntryBuffer struct {
	Mutex sync.RWMutex

	//Keeps track of entries that have been requested
	//So we can forward them straight into the database
	RequestedEntries map[string]string
	//Keeps track of entries received from messages
	//That are not part of a block yet
	//To be either saved in the database, or deleted once they expire
	TempEntries map[string]*StoredData
}

func (eb *EntryBuffer) AddRequestedEntry(h interfaces.IHash) {
	eb.Mutex.Lock()
	defer eb.Mutex.Unlock()

	eb.RequestedEntries[h.String()] = "y"
}

func (eb *EntryBuffer) StoreEntry(e interfaces.IEntry) error {
	sd := new(StoredData)
	sd.Date = time.Now()
	b, err := e.MarshalBinary()
	if err != nil {
		return err
	}
	sd.Data = b
	eb.Mutex.Lock()
	defer eb.Mutex.Unlock()
	eb.TempEntries[e.GetHash().String()] = sd
	return nil
}

func (eb *EntryBuffer) ClearExpiredEntries() {
	eb.Mutex.Lock()
	defer eb.Mutex.Unlock()

	for k, v := range eb.TempEntries {
		if time.Since(v.Date) > EntryExpirationDuration {
			delete(eb.TempEntries, k)
		}
	}
}

func (eb *EntryBuffer) GetEntry(h interfaces.IHash) (interfaces.IEntry, error) {
	eb.Mutex.RLock()
	defer eb.Mutex.RUnlock()

	b, ok := eb.TempEntries[h.String()]
	if ok == false {
		return nil, nil
	}
	e := new(entryBlock.Entry)
	err := e.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	return e, nil
}
