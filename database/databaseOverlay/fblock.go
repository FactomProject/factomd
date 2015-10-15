package databaseOverlay

import (
	. "github.com/FactomProject/factomd/common/interfaces"
)

func (db *Overlay) ProcessFBlockBatch(block DatabaseBatchable) error {
	return db.ProcessBlockBatch([]byte{byte(TBL_SC)}, []byte{byte(TBL_SC_NUM)}, nil, block)
}

func (db *Overlay) FetchFBlockByHash(hash IHash, dst DatabaseBatchable) (DatabaseBatchable, error) {
	return db.FetchBlock([]byte{byte(TBL_SC)}, hash, dst)
}

func (db *Overlay) FetchAllFBlocks(sample BinaryMarshallableAndCopyable) ([]BinaryMarshallableAndCopyable, error) {
	return db.FetchAllBlocksFromBucket([]byte{byte(TBL_SC)}, sample)
}
