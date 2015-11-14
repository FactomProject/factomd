package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// ProcessECBlockBatche inserts the ECBlock and update all it's cbentries in DB
func (db *Overlay) ProcessECBlockBatch(block interfaces.DatabaseBatchable) error {
	return db.ProcessBlockBatch([]byte{byte(ENTRYCREDITBLOCK)}, []byte{byte(ENTRYCREDITBLOCK_NUMBER)}, []byte{byte(ENTRYCREDITBLOCK_KEYMR)}, block)
}

// FetchECBlockByHash gets an Entry Credit block by hash from the database.
func (db *Overlay) FetchECBlockByHash(hash interfaces.IHash) (interfaces.DatabaseBatchable, error) {
	return db.FetchBlock([]byte{byte(ENTRYCREDITBLOCK)}, hash, entryCreditBlock.NewECBlock())
}

// FetchAllECBlocks gets all of the entry credit blocks
func (db *Overlay) FetchAllECBlocks() ([]interfaces.IEntryCreditBlock, error) {
	list, err := db.FetchAllBlocksFromBucket([]byte{byte(ENTRYCREDITBLOCK)}, entryCreditBlock.NewECBlock())
	if err != nil {
		return nil, err
	}
	return toECBlocksList(list), nil
}

func toECBlocksList(source []interfaces.BinaryMarshallableAndCopyable) []interfaces.IEntryCreditBlock {
	answer := make([]interfaces.IEntryCreditBlock, len(source))
	for i, v := range source {
		answer[i] = v.(interfaces.IEntryCreditBlock)
	}
	return answer
}

func (db *Overlay) SaveECBlockHead(block interfaces.DatabaseBatchable) error {
	return db.ProcessECBlockBatch(block)
}

func (db *Overlay) FetchECBlockHead() (interfaces.IEntryCreditBlock, error) {
	blk := entryCreditBlock.NewECBlock()
	block, err := db.FetchChainHeadByChainID([]byte{byte(ENTRYCREDITBLOCK)}, primitives.NewHash(blk.GetChainID()), blk)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IEntryCreditBlock), nil
}
