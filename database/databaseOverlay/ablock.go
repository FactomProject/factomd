package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

// ProcessABlockBatch inserts the AdminBlock
func (db *Overlay) ProcessABlockBatch(block interfaces.DatabaseBatchable) error {
	return db.ProcessBlockBatch([]byte{byte(TBL_AB)}, []byte{byte(TBL_AB_NUM)}, nil, block)
}

// FetchABlockByHash gets an admin block by hash from the database.
func (db *Overlay) FetchABlockByHash(hash interfaces.IHash, dst interfaces.DatabaseBatchable) (interfaces.DatabaseBatchable, error) {
	return db.FetchBlock([]byte{byte(TBL_AB)}, hash, dst)
}

// FetchAllABlocks gets all of the admin blocks
func (db *Overlay) FetchAllABlocks(sample interfaces.BinaryMarshallableAndCopyable) ([]interfaces.BinaryMarshallableAndCopyable, error) {
	return db.FetchAllBlocksFromBucket([]byte{byte(TBL_AB)}, sample)
}

func (db *Overlay) SaveABlockHead(block interfaces.DatabaseBatchable) error {
	return db.ProcessABlockBatch(block)
}
