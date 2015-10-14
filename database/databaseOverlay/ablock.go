package databaseOverlay

import (
	. "github.com/FactomProject/factomd/common/interfaces"
)

// ProcessABlockBatch inserts the AdminBlock
func (db *Overlay) ProcessABlockBatch(block DatabaseBatchable) error {
	return db.ProcessBlockBatch([]byte{byte(TBL_AB)}, []byte{byte(TBL_AB_NUM)}, block)
}

// FetchABlockByHash gets an admin block by hash from the database.
func (db *Overlay) FetchABlockByHash(hash IHash, dst BinaryMarshallable) (BinaryMarshallable, error) {
	return db.FetchBlockByHash([]byte{byte(TBL_AB)}, hash, dst)
}

// FetchAllABlocks gets all of the admin blocks
func (db *Overlay) FetchAllABlocks(sample BinaryMarshallableAndCopyable) ([]BinaryMarshallableAndCopyable, error) {
	return db.FetchAllBlocksFromBucket([]byte{byte(TBL_AB)}, sample)
}
