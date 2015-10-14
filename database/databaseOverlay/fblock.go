package databaseOverlay

import (
	. "github.com/FactomProject/factomd/common/interfaces"
)

func (db *Overlay) ProcessFBlockBatch(block DatabaseBatchable) error {
	return db.ProcessBlockBatch([]byte{byte(TBL_SC)}, []byte{byte(TBL_SC_NUM)}, block)
}

func (db *Overlay) FetchFBlockByHash(hash IHash, dst BinaryMarshallable) (BinaryMarshallable, error) {
	return db.FetchBlockByHash([]byte{byte(TBL_SC)}, hash, dst)
}

func (db *Overlay) FetchAllFBlocks(sample BinaryMarshallableAndCopyable) ([]BinaryMarshallableAndCopyable, error) {
	return db.FetchAllBlocksFromBucket([]byte{byte(TBL_SC)}, sample)
}
