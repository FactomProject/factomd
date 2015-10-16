package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

// ProcessECBlockBatche inserts the ECBlock and update all it's cbentries in DB
func (db *Overlay) ProcessECBlockBatch(block interfaces.DatabaseBatchable) error {
	return db.ProcessBlockBatch([]byte{byte(TBL_CB)}, nil, nil, block)
}

// FetchECBlockByHash gets an Entry Credit block by hash from the database.
func (db *Overlay) FetchECBlockByHash(hash interfaces.IHash, dst interfaces.DatabaseBatchable) (interfaces.DatabaseBatchable, error) {
	return db.FetchBlock([]byte{byte(TBL_CB)}, hash, dst)
}

// FetchAllECBlocks gets all of the entry credit blocks
func (db *Overlay) FetchAllECBlocks(sample interfaces.BinaryMarshallableAndCopyable) ([]interfaces.BinaryMarshallableAndCopyable, error) {
	return db.FetchAllBlocksFromBucket([]byte{byte(TBL_CB)}, sample)
}
