package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/util"
	"sort"
)

// ProcessABlockBatch inserts the AdminBlock
func (db *Overlay) ProcessABlockBatch(block interfaces.DatabaseBatchable) error {
	return db.ProcessBlockBatch(ADMINBLOCK, ADMINBLOCK_NUMBER, ADMINBLOCK_SECONDARYINDEX, block)
}

func (db *Overlay) ProcessABlockBatchWithoutHead(block interfaces.DatabaseBatchable) error {
	return db.ProcessBlockBatchWithoutHead(ADMINBLOCK, ADMINBLOCK_NUMBER, ADMINBLOCK_SECONDARYINDEX, block)
}

func (db *Overlay) ProcessABlockMultiBatch(block interfaces.DatabaseBatchable) error {
	return db.ProcessBlockMultiBatch(ADMINBLOCK, ADMINBLOCK_NUMBER, ADMINBLOCK_SECONDARYINDEX, block)
}

func (db *Overlay) FetchABlock(hash interfaces.IHash) (interfaces.IAdminBlock, error) {
	block, err := db.FetchABlockByPrimary(hash)
	if err != nil {
		return nil, err
	}
	if block != nil {
		return block, nil
	}
	return db.FetchABlockBySecondary(hash)
}

// FetchABlockByHash gets an admin block by hash from the database.
func (db *Overlay) FetchABlockBySecondary(hash interfaces.IHash) (interfaces.IAdminBlock, error) {
	block, err := db.FetchBlockBySecondaryIndex(ADMINBLOCK_SECONDARYINDEX, ADMINBLOCK, hash, new(adminBlock.AdminBlock))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IAdminBlock), nil
}

// FetchABlockByKeyMR gets an admin block by keyMR from the database.
func (db *Overlay) FetchABlockByPrimary(hash interfaces.IHash) (interfaces.IAdminBlock, error) {
	block, err := db.FetchBlock(ADMINBLOCK, hash, new(adminBlock.AdminBlock))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IAdminBlock), nil
}

// FetchABlockByHeight gets an admin block by height from the database.
func (db *Overlay) FetchABlockByHeight(blockHeight uint32) (interfaces.IAdminBlock, error) {
	block, err := db.FetchBlockByHeight(ADMINBLOCK_NUMBER, ADMINBLOCK, blockHeight, new(adminBlock.AdminBlock))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IAdminBlock), nil
}

// FetchAllABlocks gets all of the admin blocks
func (db *Overlay) FetchAllABlocks() ([]interfaces.IAdminBlock, error) {
	list, err := db.FetchAllBlocksFromBucket(ADMINBLOCK, new(adminBlock.AdminBlock))
	if err != nil {
		return nil, err
	}
	return toABlocksList(list), nil
}

func (db *Overlay) FetchAllABlockKeys() ([]interfaces.IHash, error) {
	return db.FetchAllBlockKeysFromBucket(ADMINBLOCK)
}

func toABlocksList(source []interfaces.BinaryMarshallableAndCopyable) []interfaces.IAdminBlock {
	answer := make([]interfaces.IAdminBlock, len(source))
	for i, v := range source {
		answer[i] = v.(interfaces.IAdminBlock)
	}
	sort.Sort(util.ByABlockIDAccending(answer))
	return answer
}

func (db *Overlay) SaveABlockHead(block interfaces.DatabaseBatchable) error {
	return db.ProcessABlockBatch(block)
}

func (db *Overlay) FetchABlockHead() (interfaces.IAdminBlock, error) {
	blk := adminBlock.NewAdminBlock(nil)
	block, err := db.FetchChainHeadByChainID(ADMINBLOCK, primitives.NewHash(blk.GetChainID().Bytes()), blk)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IAdminBlock), nil
}
