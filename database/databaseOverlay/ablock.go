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
	return db.ProcessBlockBatch([]byte{byte(ADMINBLOCK)}, []byte{byte(ADMINBLOCK_NUMBER)}, []byte{byte(ADMINBLOCK_KEYMR)}, block)
}

func (db *Overlay) ProcessABlockMultiBatch(block interfaces.DatabaseBatchable) error {
	return db.ProcessBlockMultiBatch([]byte{byte(ADMINBLOCK)}, []byte{byte(ADMINBLOCK_NUMBER)}, []byte{byte(ADMINBLOCK_KEYMR)}, block)
}

// FetchABlockByHash gets an admin block by hash from the database.
func (db *Overlay) FetchABlockByHash(hash interfaces.IHash) (interfaces.IAdminBlock, error) {
	block, err := db.FetchBlockBySecondaryIndex([]byte{byte(ADMINBLOCK_KEYMR)}, []byte{byte(ADMINBLOCK)}, hash, new(adminBlock.AdminBlock))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IAdminBlock), nil
}

// FetchABlockByKeyMR gets an admin block by keyMR from the database.
func (db *Overlay) FetchABlockByKeyMR(hash interfaces.IHash) (interfaces.IAdminBlock, error) {
	block, err := db.FetchBlock([]byte{byte(ADMINBLOCK)}, hash, new(adminBlock.AdminBlock))
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
	list, err := db.FetchAllBlocksFromBucket([]byte{byte(ADMINBLOCK)}, new(adminBlock.AdminBlock))
	if err != nil {
		return nil, err
	}
	return toABlocksList(list), nil
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
	blk := adminBlock.NewAdminBlock()
	block, err := db.FetchChainHeadByChainID([]byte{byte(ADMINBLOCK)}, primitives.NewHash(blk.GetChainID().Bytes()), blk)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IAdminBlock), nil
}
