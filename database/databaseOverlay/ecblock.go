package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/util"
	"sort"
)

// ProcessECBlockBatch inserts the ECBlock and update all it's cbentries in DB
func (db *Overlay) ProcessECBlockBatch(block interfaces.IEntryCreditBlock, checkForDuplicateEntries bool) error {
	err := db.ProcessBlockBatch([]byte{byte(ENTRYCREDITBLOCK)},
		[]byte{byte(ENTRYCREDITBLOCK_NUMBER)},
		[]byte{byte(ENTRYCREDITBLOCK_KEYMR)}, block)
	if err != nil {
		return err
	}
	err = db.SaveIncludedInMultiFromBlock(block, false)
	if err != nil {
		return err
	}
	return db.SavePaidForMultiFromBlock(block, checkForDuplicateEntries)
}

func (db *Overlay) ProcessECBlockMultiBatch(block interfaces.IEntryCreditBlock, checkForDuplicateEntries bool) error {
	err := db.ProcessBlockMultiBatch([]byte{byte(ENTRYCREDITBLOCK)},
		[]byte{byte(ENTRYCREDITBLOCK_NUMBER)},
		[]byte{byte(ENTRYCREDITBLOCK_KEYMR)}, block)
	if err != nil {
		return err
	}
	err = db.SaveIncludedInMultiFromBlockMultiBatch(block, true)
	if err != nil {
		return err
	}
	return db.SavePaidForMultiFromBlockMultiBatch(block, checkForDuplicateEntries)
}

// FetchECBlockByHeaderHash gets an Entry Credit block by hash from the database.
func (db *Overlay) FetchECBlockByHeaderHash(hash interfaces.IHash) (interfaces.IEntryCreditBlock, error) {
	block, err := db.FetchBlockBySecondaryIndex([]byte{byte(ENTRYCREDITBLOCK_KEYMR)}, []byte{byte(ENTRYCREDITBLOCK)}, hash, entryCreditBlock.NewECBlock())
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IEntryCreditBlock), nil
}

// FetchECBlockByHash gets an Entry Credit block by hash from the database.
func (db *Overlay) FetchECBlockByHash(hash interfaces.IHash) (interfaces.IEntryCreditBlock, error) {
	block, err := db.FetchBlock([]byte{byte(ENTRYCREDITBLOCK)}, hash, entryCreditBlock.NewECBlock())
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IEntryCreditBlock), nil
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
	sort.Sort(util.ByECBlockIDAccending(answer))
	return answer
}

func (db *Overlay) SaveECBlockHead(block interfaces.IEntryCreditBlock, checkForDuplicateEntries bool) error {
	return db.ProcessECBlockBatch(block, checkForDuplicateEntries)
}

func (db *Overlay) FetchECBlockHead() (interfaces.IEntryCreditBlock, error) {
	blk := entryCreditBlock.NewECBlock()
	block, err := db.FetchChainHeadByChainID([]byte{byte(ENTRYCREDITBLOCK)}, primitives.NewHash(blk.GetChainID().Bytes()), blk)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.IEntryCreditBlock), nil
}
