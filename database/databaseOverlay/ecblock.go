package databaseOverlay

import (
	. "github.com/FactomProject/factomd/common/interfaces"
)

// ProcessECBlockBatche inserts the ECBlock and update all it's cbentries in DB
func (db *Overlay) ProcessECBlockBatch(block DatabaseBatchable) error {
	return db.ProcessBlockBatch([]byte{byte(TBL_CB)}, nil, nil, block)
}

// FetchECBlockByHash gets an Entry Credit block by hash from the database.
func (db *Overlay) FetchECBlockByHash(hash IHash, dst DatabaseBatchable) (DatabaseBatchable, error) {
	return db.FetchBlock([]byte{byte(TBL_CB)}, hash, dst)
}

// FetchAllECBlocks gets all of the entry credit blocks
func (db *Overlay) FetchAllECBlocks(sample BinaryMarshallableAndCopyable) ([]BinaryMarshallableAndCopyable, error) {
	return db.FetchAllBlocksFromBucket([]byte{byte(TBL_CB)}, sample)
}
