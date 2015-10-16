package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

func (db *Overlay) ProcessFBlockBatch(block interfaces.DatabaseBatchable) error {
	return db.ProcessBlockBatch([]byte{byte(TBL_SC)}, []byte{byte(TBL_SC_NUM)}, nil, block)
}

func (db *Overlay) FetchFBlockByHash(hash interfaces.IHash, dst interfaces.DatabaseBatchable) (interfaces.DatabaseBatchable, error) {
	return db.FetchBlock([]byte{byte(TBL_SC)}, hash, dst)
}

func (db *Overlay) FetchAllFBlocks(sample interfaces.BinaryMarshallableAndCopyable) ([]interfaces.BinaryMarshallableAndCopyable, error) {
	return db.FetchAllBlocksFromBucket([]byte{byte(TBL_SC)}, sample)
}
