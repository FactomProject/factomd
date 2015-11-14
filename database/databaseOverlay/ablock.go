package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/interfaces"
)

// ProcessABlockBatch inserts the AdminBlock
func (db *Overlay) ProcessABlockBatch(block interfaces.DatabaseBatchable) error {
	return db.ProcessBlockBatch([]byte{byte(ADMINBLOCK)}, []byte{byte(ADMINBLOCK_NUMBER)}, []byte{byte(ADMINBLOCK_KEYMR)}, block)
}

// FetchABlockByHash gets an admin block by hash from the database.
func (db *Overlay) FetchABlockByHash(hash interfaces.IHash) (interfaces.IAdminBlock, error) {
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
	return answer
}

func (db *Overlay) SaveABlockHead(block interfaces.DatabaseBatchable) error {
	return db.ProcessABlockBatch(block)
}
