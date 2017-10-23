package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	//"github.com/FactomProject/factomd/log"
	//"github.com/FactomProject/factomd/util"
	//"sort"
	"strings"
)

// ProcessEBlockBatche inserts the EBlock and update all it's ebentries in DB
func (db *Overlay) ProcessEBlockBatch(eblock interfaces.DatabaseBlockWithEntries, checkForDuplicateEntries bool) error {
	//Each chain has its own number bucket, otherwise we would have conflicts
	err := db.ProcessBlockBatchWithoutHead(ENTRYBLOCK, nil, ENTRYBLOCK_SECONDARYINDEX, eblock)
	if err != nil {
		return err
	}
	return db.SaveIncludedInMultiFromBlock(eblock, checkForDuplicateEntries)
}

func (db *Overlay) ProcessEBlockMultiBatch(eblock interfaces.DatabaseBlockWithEntries, checkForDuplicateEntries bool) error {
	//Each chain has its own number bucket, otherwise we would have conflicts
	err := db.ProcessBlockMultiBatchWithoutHead(ENTRYBLOCK, nil, ENTRYBLOCK_SECONDARYINDEX, eblock)
	if err != nil {
		return err
	}
	return db.SaveIncludedInMultiFromBlockMultiBatch(eblock, checkForDuplicateEntries)
}

func (db *Overlay) FetchEBlock(hash interfaces.IHash) (interfaces.IEntryBlock, error) {
	block, err := db.FetchEBlockByPrimary(hash)
	if err != nil {
		return nil, err
	}
	if block != nil {
		return block, nil
	}
	return db.FetchEBlockBySecondary(hash)
}

// FetchEBlockByHash gets an entry block by merkle root from the database.
func (db *Overlay) FetchEBlockBySecondary(hash interfaces.IHash) (interfaces.IEntryBlock, error) {
	block, err := db.FetchBlockBySecondaryIndex(ENTRYBLOCK_SECONDARYINDEX, ENTRYBLOCK, hash, entryBlock.NewEBlock())
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IEntryBlock), nil
}

// FetchEBlockByKeyMR gets an entry by hash from the database.
func (db *Overlay) FetchEBlockByPrimary(hash interfaces.IHash) (interfaces.IEntryBlock, error) {
	block, err := db.FetchBlock(ENTRYBLOCK, hash, entryBlock.NewEBlock())
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IEntryBlock), nil
}

// FetchEBKeyMRByHash gets an entry by hash from the database.
func (db *Overlay) FetchEBKeyMRByHash(hash interfaces.IHash) (interfaces.IHash, error) {
	return db.FetchPrimaryIndexBySecondaryIndex(ENTRYBLOCK_SECONDARYINDEX, hash)
}

// FetchAllEBlocksByChain gets all of the blocks by chain id
func (db *Overlay) FetchAllEBlocksByChain(chainID interfaces.IHash) ([]interfaces.IEntryBlock, error) {
	bucket := append(ENTRYBLOCK_CHAIN_NUMBER, chainID.Bytes()...)
	keyList, err := db.FetchAllBlocksFromBucket(bucket, new(primitives.Hash))
	if err != nil {
		return nil, err
	}

	list := make([]interfaces.IEntryBlock, len(keyList))

	for i, v := range keyList {
		block, err := db.FetchEBlock(v.(interfaces.IHash))
		if err != nil {
			return nil, err
		}
		list[i] = block
	}

	return list, nil
}

func (db *Overlay) SaveEBlock(block interfaces.DatabaseBlockWithEntries, checkForDuplicateEntries bool) error {
	return db.ProcessEBlockBatch(block, checkForDuplicateEntries)
}

func (db *Overlay) FetchEBlockHead(chainID interfaces.IHash) (interfaces.IEntryBlock, error) {
	block, err := db.FetchChainHeadByChainID(ENTRYBLOCK, chainID, entryBlock.NewEBlock())
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(*entryBlock.EBlock), nil
}

func (db *Overlay) FetchAllEBlockChainIDs() ([]interfaces.IHash, error) {
	ids, err := db.FetchAllBlockKeysFromBucket(CHAIN_HEAD)
	if err != nil {
		return nil, err
	}
	entries := []interfaces.IHash{}
	for _, h := range ids {
		str := h.String()
		if strings.Contains(str, "000000000000000000000000000000000000000000000000000000000000000") {
			//skipping basic blocks
			continue
		}
		entries = append(entries, h)
	}
	return entries, nil
}
