package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
)

// InsertEntry inserts an entry
func (db *Overlay) InsertEntry(entry interfaces.IEBEntry) error {
	if entry == nil {
		return nil
	}

	//Entries are saved in buckets represented by their chainID for easy batch loading
	//They are also indexed in ENTRY bucket by their hash that points to their chainID
	//So they can be loaded in two load operations without needing to know their chainID

	batch := []interfaces.Record{}
	batch = append(batch, interfaces.Record{entry.GetChainID(), entry.DatabasePrimaryIndex().Bytes(), entry})
	batch = append(batch, interfaces.Record{[]byte{byte(ENTRY)}, entry.DatabasePrimaryIndex().Bytes(), entry.GetChainIDHash()})

	return db.PutInBatch(batch)
}

// FetchEntry gets an entry by hash from the database.
func (db *Overlay) FetchEntryByHash(hash interfaces.IHash) (interfaces.IEBEntry, error) {
	chainID, err := db.FetchPrimaryIndexBySecondaryIndex([]byte{byte(ENTRY)}, hash)
	if err != nil {
		return nil, err
	}
	if chainID == nil {
		return nil, nil
	}

	entry, err := db.FetchBlock(chainID.Bytes(), hash, entryBlock.NewEntry())
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	return entry.(interfaces.IEBEntry), nil
}

func (db *Overlay) FetchAllEntriesByChainID(chainID interfaces.IHash) ([]interfaces.IEBEntry, error) {
	list, err := db.FetchAllBlocksFromBucket(chainID.Bytes(), entryBlock.NewEntry())
	if err != nil {
		return nil, err
	}
	return toEntryList(list), nil
}

func toEntryList(source []interfaces.BinaryMarshallableAndCopyable) []interfaces.IEBEntry {
	answer := make([]interfaces.IEBEntry, len(source))
	for i, v := range source {
		answer[i] = v.(interfaces.IEBEntry)
	}
	return answer
}

// *************************************************
//TODO fix wsapi.go when Entry is updated. line 144
