package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	//"github.com/FactomProject/factomd/util"
	//"sort"
)

// ProcessEBlockBatche inserts the EBlock and update all it's ebentries in DB
func (db *Overlay) ProcessEBlockBatch(eblock interfaces.DatabaseBatchable) error {
	return db.ProcessBlockBatch([]byte{byte(ENTRYBLOCK)}, []byte{byte(ENTRYBLOCK_CHAIN_NUMBER)}, []byte{byte(ENTRYBLOCK_KEYMR)}, eblock)
}

// FetchEBlockByMR gets an entry block by merkle root from the database.
func (db *Overlay) FetchEBlockByHash(hash interfaces.IHash) (interfaces.IEntryBlock, error) {
	block, err := db.FetchBlockBySecondaryIndex([]byte{byte(ENTRYBLOCK_KEYMR)}, []byte{byte(ENTRYBLOCK)}, hash, entryBlock.NewEBlock())
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IEntryBlock), nil
}

// FetchEntryBlock gets an entry by hash from the database.
func (db *Overlay) FetchEBlockByKeyMR(hash interfaces.IHash) (interfaces.IEntryBlock, error) {
	block, err := db.FetchBlock([]byte{byte(ENTRYBLOCK)}, hash, entryBlock.NewEBlock())
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IEntryBlock), nil
}

// FetchEBHashByMR gets an entry by hash from the database.
func (db *Overlay) FetchEBKeyMRByHash(hash interfaces.IHash) (interfaces.IHash, error) {
	return db.FetchPrimaryIndexBySecondaryIndex([]byte{byte(ENTRYBLOCK_KEYMR)}, hash)
}

/*
// InsertChain inserts the newly created chain into db
func (db *Overlay) InsertChain(chain *EChain) error {
	bucket := []byte{byte(ENTRYCHAIN)}
	key := chain.ChainID.Bytes()
	err := db.DB.Put(bucket, key, chain)
	if err != nil {
		return err
	}
	return nil
}

// FetchAllChains get all of the cahins
func (db *Overlay) FetchAllChains() (chains []*EChain, err error) {
	bucket := []byte{byte(ENTRYCHAIN)}

	list, err := db.DB.GetAll(bucket, new(primitives.Hash))
	if err != nil {
		return nil, err
	}
	answer := make([]*EChain, len(list))
	for i, v := range list {
		answer[i] = v.(*EChain)
	}
	return answer, nil
}*/

// FetchAllEBlocksByChain gets all of the blocks by chain id
func (db *Overlay) FetchAllEBlocksByChain(chainID interfaces.IHash) ([]interfaces.IEntryBlock, error) {
	bucket := append([]byte{byte(ENTRYBLOCK_CHAIN_NUMBER)}, chainID.Bytes()...)
	list, err := db.FetchAllBlocksFromBucket(bucket, entryBlock.NewEBlock())
	if err != nil {
		return nil, err
	}
	return toEBlocksList(list), nil
}

func toEBlocksList(source []interfaces.BinaryMarshallableAndCopyable) []interfaces.IEntryBlock {
	answer := make([]interfaces.IEntryBlock, len(source))
	for i, v := range source {
		answer[i] = v.(interfaces.IEntryBlock)
	}
	//sort.Sort(util.ByEBlockIDAccending(answer))
	return answer
}

func (db *Overlay) SaveEBlockHead(block interfaces.DatabaseBatchable) error {
	return db.ProcessEBlockBatch(block)
}

func (db *Overlay) FetchEBlockHead(chainID interfaces.IHash) (interfaces.IEntryBlock, error) {
	block, err := db.FetchChainHeadByChainID([]byte{byte(ENTRYBLOCK)}, chainID, entryBlock.NewEBlock())
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(*entryBlock.EBlock), nil
}
