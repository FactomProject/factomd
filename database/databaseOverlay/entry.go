package databaseOverlay

import (
	. "github.com/FactomProject/factomd/common/interfaces"
)

// InsertEntry inserts an entry
func (db *Overlay) InsertEntry(entry DatabaseBatchable) error {
	return db.Insert([]byte{byte(TBL_ENTRY)}, entry)
}

// FetchEntry gets an entry by hash from the database.
func (db *Overlay) FetchEntryByHash(hash IHash, dst DatabaseBatchable) (DatabaseBatchable, error) {
	return db.FetchBlock([]byte{byte(TBL_ENTRY)}, hash, dst)
}
