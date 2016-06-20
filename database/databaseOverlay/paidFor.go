package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func (db *Overlay) SavePaidFor(entry, ecEntry interfaces.IHash) error {
	if entry == nil || ecEntry == nil {
		return nil
	}
	batch := []interfaces.Record{}

	batch = append(batch, interfaces.Record{PAID_FOR, entry.Bytes(), ecEntry})

	err := db.DB.PutInBatch(batch)
	if err != nil {
		return err
	}

	return nil
}

func (db *Overlay) SavePaidForMultiFromBlockMultiBatch(block interfaces.IEntryCreditBlock, checkForDuplicateEntries bool) error {
	if block == nil {
		return nil
	}
	batch := []interfaces.Record{}

	for _, entry := range block.GetBody().GetEntries() {
		if entry.ECID() != entryCreditBlock.ECIDChainCommit && entry.ECID() != entryCreditBlock.ECIDEntryCommit {
			continue
		}
		var entryHash interfaces.IHash

		if entry.ECID() == entryCreditBlock.ECIDChainCommit {
			entryHash = entry.(*entryCreditBlock.CommitChain).EntryHash
		}
		if entry.ECID() == entryCreditBlock.ECIDEntryCommit {
			entryHash = entry.(*entryCreditBlock.CommitEntry).EntryHash
		}

		if checkForDuplicateEntries == true {
			loaded, err := db.Get(PAID_FOR, entryHash.Bytes(), primitives.NewZeroHash())
			if err != nil {
				return err
			}
			if loaded != nil {
				continue
			}
		}
		batch = append(batch, interfaces.Record{PAID_FOR, entryHash.Bytes(), entry.Hash()})
	}
	if len(batch) == 0 {
		return nil
	}

	db.PutInMultiBatch(batch)

	return nil
}

func (db *Overlay) SavePaidForMultiFromBlock(block interfaces.IEntryCreditBlock, checkForDuplicateEntries bool) error {
	if block == nil {
		return nil
	}
	batch := []interfaces.Record{}

	for _, entry := range block.GetBody().GetEntries() {
		if entry.ECID() != entryCreditBlock.ECIDChainCommit && entry.ECID() != entryCreditBlock.ECIDEntryCommit {
			continue
		}
		var entryHash interfaces.IHash

		if entry.ECID() == entryCreditBlock.ECIDChainCommit {
			entryHash = entry.(*entryCreditBlock.CommitChain).EntryHash
		}
		if entry.ECID() == entryCreditBlock.ECIDEntryCommit {
			entryHash = entry.(*entryCreditBlock.CommitEntry).EntryHash
		}

		if checkForDuplicateEntries == true {
			loaded, err := db.Get(PAID_FOR, entryHash.Bytes(), primitives.NewZeroHash())
			if err != nil {
				return err
			}
			if loaded != nil {
				continue
			}
		}
		batch = append(batch, interfaces.Record{PAID_FOR, entryHash.Bytes(), entry.Hash()})
	}
	if len(batch) == 0 {
		return nil
	}

	err := db.DB.PutInBatch(batch)
	if err != nil {
		return err
	}

	return nil
}

func (db *Overlay) FetchPaidFor(hash interfaces.IHash) (interfaces.IHash, error) {
	block, err := db.DB.Get(PAID_FOR, hash.Bytes(), new(primitives.Hash))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IHash), nil
}
