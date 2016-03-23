package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/factoid/block"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/util"
	"sort"
)

func (db *Overlay) ProcessFBlockBatch(block interfaces.DatabaseBlockWithEntries) error {
	err := db.ProcessBlockBatch([]byte{byte(FACTOIDBLOCK)}, []byte{byte(FACTOIDBLOCK_NUMBER)}, []byte{byte(FACTOIDBLOCK_KEYMR)}, block)
	if err != nil {
		return err
	}
	return db.SaveIncludedInMultiFromBlock(block)
}

func (db *Overlay) ProcessFBlockMultiBatch(block interfaces.DatabaseBlockWithEntries) error {
	err := db.ProcessBlockMultiBatch([]byte{byte(FACTOIDBLOCK)}, []byte{byte(FACTOIDBLOCK_NUMBER)}, []byte{byte(FACTOIDBLOCK_KEYMR)}, block)
	if err != nil {
		return err
	}
	return db.SaveIncludedInMultiFromBlockMultiBatch(block)
}

func (db *Overlay) FetchFBlockByHash(hash interfaces.IHash) (interfaces.IFBlock, error) {
	block, err := db.FetchBlockBySecondaryIndex([]byte{byte(FACTOIDBLOCK_KEYMR)}, []byte{byte(FACTOIDBLOCK)}, hash, new(block.FBlock))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IFBlock), nil
}

func (db *Overlay) FetchFBlockByKeyMR(keyMR interfaces.IHash) (interfaces.IFBlock, error) {
	block, err := db.FetchBlock([]byte{byte(FACTOIDBLOCK)}, keyMR, new(block.FBlock))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IFBlock), nil
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
	sort.Sort(util.ByFBlockIDAccending(answer))
	return answer
}

func (db *Overlay) SaveFactoidBlockHead(fblock interfaces.DatabaseBlockWithEntries) error {
	return db.ProcessFBlockBatch(fblock)
}

func (db *Overlay) FetchFactoidBlockHead() (interfaces.IFBlock, error) {
	blk := new(block.FBlock)
	block, err := db.FetchChainHeadByChainID([]byte{byte(FACTOIDBLOCK)}, primitives.NewHash(blk.GetChainID().Bytes()), blk)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IFBlock), nil
}
