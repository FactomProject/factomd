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
func (db *Overlay) FetchECBlockByHash(hash interfaces.IHash, dst interfaces.DatabaseBatchable) (interfaces.DatabaseBatchable, error) {
	return db.FetchBlock([]byte{byte(ENTRYCREDITBLOCK)}, hash, dst)
}

// FetchAllECBlocks gets all of the entry credit blocks
func (db *Overlay) FetchAllECBlocks(sample interfaces.BinaryMarshallableAndCopyable) ([]interfaces.BinaryMarshallableAndCopyable, error) {
	return db.FetchAllBlocksFromBucket([]byte{byte(ENTRYCREDITBLOCK)}, sample)
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
