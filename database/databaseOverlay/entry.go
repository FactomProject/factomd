package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
)

// InsertEntry inserts an entry
func (db *Overlay) InsertEntry(entry interfaces.DatabaseBatchable) error {
	return db.Insert([]byte{byte(ENTRY)}, entry)
}

// FetchEntry gets an entry by hash from the database.
func (db *Overlay) FetchEntryByHash(hash interfaces.IHash) (interfaces.IEBEntry, error) {
	entry, err := db.FetchBlock([]byte{byte(ENTRY)}, hash, entryBlock.NewEntry())
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}
	return entry.(interfaces.IEBEntry), nil
}

// *************************************************
//TODO fix wsapi.go when Entry is updated. line 144
