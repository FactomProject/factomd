package databaseOverlay

import (
	. "github.com/FactomProject/factomd/common/EntryBlock"
	. "github.com/FactomProject/factomd/common/interfaces"
)

// InsertEntry inserts an entry
func (db *Overlay) InsertEntry(entry IEntry) error {
	bucket := []byte{byte(TBL_ENTRY)}
	key := entry.Hash().Bytes()
	err := db.DB.Put(bucket, key, entry)
	if err != nil {
		return err
	}
	return nil
}

// FetchEntry gets an entry by hash from the database.
func (db *Overlay) FetchEntryByHash(entrySha IHash) (IEntry, error) {
	bucket := []byte{byte(TBL_ENTRY)}
	key := entrySha.Bytes()

	entry, err := db.DB.Get(bucket, key, new(Entry))
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}
	return entry.(IEntry), nil
}

/*
// Initialize External ID map for explorer search
func (db *Overlay) InitializeExternalIDMap() (extIDMap map[string]bool, err error) {

	var fromkey []byte = []byte{byte(TBL_ENTRY)} // Table Name (1 bytes)

	var tokey []byte = []byte{byte(TBL_ENTRY + 1)} // Table Name (1 bytes)

	extIDMap = make(map[string]bool)

	iter := db.lDb.NewIterator(&util.Range{Start: fromkey, Limit: tokey}, db.ro)

	for iter.Next() {
		entry := new(Entry)
		_, err := entry.UnmarshalBinaryData(iter.Value())
		if err != nil {
			return nil, err
		}
		if entry.ExtIDs != nil {
			for i := 0; i < len(entry.ExtIDs); i++ {
				mapKey := string(iter.Key()[1:])
				mapKey = mapKey + strings.ToLower(string(entry.ExtIDs[i]))
				extIDMap[mapKey] = true
			}
		}

	}
	iter.Release()
	err = iter.Error()

	return extIDMap, nil
}
*/
