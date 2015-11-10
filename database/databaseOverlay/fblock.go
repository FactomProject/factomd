package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/factoid/block"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
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

func (db *Overlay) FetchFactoidBlockHead() (interfaces.IFBlock, error) {
	fblock := new(block.FBlock)
	block, err := db.FetchChainHeadByChainID([]byte{byte(FACTOIDBLOCK)}, primitives.NewHash(fblock.GetChainID()), fblock)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IFBlock), nil
}

func (db *Overlay) SaveFactoidBlockHead(fblock interfaces.DatabaseBatchable) error {
	return db.ProcessFBlockBatch(fblock)
}
