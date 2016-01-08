package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func (db *Overlay) SaveIncludedIn(entry, block interfaces.IHash) error {
	if entry == nil || block == nil {
		return nil
	}
	batch := []interfaces.Record{}

	batch = append(batch, interfaces.Record{[]byte{INCLUDED_IN}, entry.Bytes(), block})

	err := db.DB.PutInBatch(batch)
	if err != nil {
		return err
	}

	return nil
}

func (db *Overlay) SaveIncludedInMultiFromBlock(block interfaces.DatabaseBlockWithEntries) error {
	entries := block.GetEntryHashes()
	hash := block.DatabasePrimaryIndex()

	return db.SaveIncludedInMulti(entries, hash)
}

func (db *Overlay) SaveIncludedInMulti(entries []interfaces.IHash, block interfaces.IHash) error {
	if entries == nil || block == nil {
		return nil
	}
	batch := []interfaces.Record{}

	for _, entry := range entries {
		batch = append(batch, interfaces.Record{[]byte{INCLUDED_IN}, entry.Bytes(), block})
	}

	err := db.DB.PutInBatch(batch)
	if err != nil {
		return err
	}

	return nil
}

func (db *Overlay) LoadIncludedIn(hash interfaces.IHash) (interfaces.IHash, error) {
	block, err := db.DB.Get([]byte{INCLUDED_IN}, hash.Bytes(), new(primitives.Hash))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IHash), nil
}
