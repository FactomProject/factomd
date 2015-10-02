package databaseOverlay

import (
	. "github.com/FactomProject/factomd/common/EntryBlock"
	//. "github.com/FactomProject/factomd/common/interfaces"
)



// InsertEntry inserts an entry
func (db *Overlay) InsertEntry(entry *Entry) error {
	bucket := []byte{byte(TBL_ENTRY)}
	key := entry.Hash().Bytes()
	err:=db.DB.Put(bucket, key, entry)
	if err!=nil {
		return err
	}
	return nil
}

/*
// FetchEntry gets an entry by hash from the database.
func (db *Overlay) FetchEntryByHash(entrySha IHash) (entry *Entry, err error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var key []byte = []byte{byte(TBL_ENTRY)}
	key = append(key, entrySha.Bytes()...)
	data, err := db.lDb.Get(key, db.ro)

	if data != nil {
		entry = new(Entry)
		_, err := entry.UnmarshalBinaryData(data)
		if err != nil {
			return nil, err
		}
	}
	return entry, nil
}

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