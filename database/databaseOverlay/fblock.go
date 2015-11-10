package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

func (db *Overlay) ProcessFBlockBatch(block interfaces.DatabaseBatchable) error {
	return db.ProcessBlockBatch([]byte{byte(FACTOIDBLOCK)}, []byte{byte(FACTOIDBLOCK_NUMBER)}, nil, block)
}

func (db *Overlay) FetchFBlockByHash(hash interfaces.IHash, dst interfaces.DatabaseBatchable) (interfaces.DatabaseBatchable, error) {
	return db.FetchBlock([]byte{byte(FACTOIDBLOCK)}, hash, dst)
}

func (db *Overlay) FetchAllFBlocks(sample interfaces.BinaryMarshallableAndCopyable) ([]interfaces.BinaryMarshallableAndCopyable, error) {
	return db.FetchAllBlocksFromBucket([]byte{byte(FACTOIDBLOCK)}, sample)
}

func (db *Overlay) SaveFactoidBlockHead(fblock interfaces.DatabaseBatchable) error {
	return db.ProcessFBlockBatch(fblock)
}
