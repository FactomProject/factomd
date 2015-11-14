package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/factoid/block"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func (db *Overlay) ProcessFBlockBatch(block interfaces.DatabaseBatchable) error {
	return db.ProcessBlockBatch([]byte{byte(FACTOIDBLOCK)}, []byte{byte(FACTOIDBLOCK_NUMBER)}, []byte{byte(FACTOIDBLOCK_KEYMR)}, block)
}

func (db *Overlay) FetchFBlockByHash(hash interfaces.IHash) (interfaces.DatabaseBatchable, error) {
	return db.FetchBlock([]byte{byte(FACTOIDBLOCK)}, hash, new(block.FBlock))
}

func (db *Overlay) FetchAllFBlocks() ([]interfaces.IFBlock, error) {
	list, err := db.FetchAllBlocksFromBucket([]byte{byte(FACTOIDBLOCK)}, new(block.FBlock))
	if err != nil {
		return nil, err
	}
	return toFactoidList(list), nil
}

func toFactoidList(source []interfaces.BinaryMarshallableAndCopyable) []interfaces.IFBlock {
	answer := make([]interfaces.IFBlock, len(source))
	for i, v := range source {
		answer[i] = v.(interfaces.IFBlock)
	}
	return answer
}

func (db *Overlay) SaveFactoidBlockHead(fblock interfaces.DatabaseBatchable) error {
	return db.ProcessFBlockBatch(fblock)
}

func (db *Overlay) FetchFactoidBlockHead() (interfaces.IFBlock, error) {
	blk := new(block.FBlock)
	block, err := db.FetchChainHeadByChainID([]byte{byte(FACTOIDBLOCK)}, primitives.NewHash(blk.GetChainID()), blk)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IFBlock), nil
}
