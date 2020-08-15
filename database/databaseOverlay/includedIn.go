package databaseOverlay

import (
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
)

func (db *Overlay) SaveIncludedIn(entry, block interfaces.IHash) error {
	if entry == nil || block == nil {
		return nil
	}
	batch := []interfaces.Record{}

	batch = append(batch, interfaces.Record{INCLUDED_IN, entry.Bytes(), block})

	err := db.DB.PutInBatch(batch)
	if err != nil {
		return err
	}

	return nil
}

func (db *Overlay) SaveIncludedInMultiFromBlockMultiBatch(block interfaces.DatabaseBlockWithEntries, checkForDuplicateEntries bool) error {
	entries := block.GetEntryHashes()
	entries = append(entries, block.GetEntrySigHashes()...)
	hash := block.DatabasePrimaryIndex()

	return db.SaveIncludedInMultiMultiBatch(entries, hash, checkForDuplicateEntries)
}

func (db *Overlay) SaveIncludedInMultiFromBlock(block interfaces.DatabaseBlockWithEntries, checkForDuplicateEntries bool) error {
	entries := block.GetEntryHashes()
	entries = append(entries, block.GetEntrySigHashes()...)
	hash := block.DatabasePrimaryIndex()

	return db.SaveIncludedInMulti(entries, hash, checkForDuplicateEntries)
}

func (db *Overlay) SaveIncludedInMultiMultiBatch(entries []interfaces.IHash, block interfaces.IHash, checkForDuplicateEntries bool) error {
	if entries == nil || block == nil {
		return nil
	}
	batch := []interfaces.Record{}

	for _, entry := range entries {
		if entry.IsMinuteMarker() == true {
			continue
		}
		if checkForDuplicateEntries == true {
			exists, err := db.DoesKeyExist(INCLUDED_IN, entry.Bytes())
			if err != nil {
				return err
			}
			if exists == true {
				continue
			}
		}
		batch = append(batch, interfaces.Record{INCLUDED_IN, entry.Bytes(), block})
	}

	db.PutInMultiBatch(batch)

	return nil
}

func (db *Overlay) SaveIncludedInMulti(entries []interfaces.IHash, block interfaces.IHash, checkForDuplicateEntries bool) error {
	if entries == nil || block == nil {
		return nil
	}
	batch := []interfaces.Record{}

	for _, entry := range entries {
		if checkForDuplicateEntries == true {
			exists, err := db.DoesKeyExist(INCLUDED_IN, entry.Bytes())
			if err != nil {
				return err
			}
			if exists == true {
				continue
			}
		}
		batch = append(batch, interfaces.Record{INCLUDED_IN, entry.Bytes(), block})
	}

	err := db.DB.PutInBatch(batch)
	if err != nil {
		return err
	}

	return nil
}

func (db *Overlay) FetchIncludedIn(hash interfaces.IHash) (interfaces.IHash, error) {
	block, err := db.DB.Get(INCLUDED_IN, hash.Bytes(), new(primitives.Hash))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IHash), nil
}
