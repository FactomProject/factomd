package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/util"
	"sort"
)

// ProcessDirBlockInfoBatch inserts the dirblock info block
func (db *Overlay) ProcessDirBlockInfoBatch(block interfaces.DatabaseBatchable) error {
	return db.ProcessBlockBatch([]byte{byte(DIRBLOCKINFO)}, []byte{byte(DIRBLOCKINFO_NUMBER)}, []byte{byte(DIRBLOCKINFO_KEYMR)}, block)
}

// FetchDirBlockInfoByHash gets a dirblock info block by hash from the database.
func (db *Overlay) FetchDirBlockInfoByHash(hash interfaces.IHash) (*dbInfo.DirBlockInfo, error) {
	block, err := db.FetchBlockBySecondaryIndex([]byte{byte(DIRBLOCKINFO_KEYMR)}, []byte{byte(DIRBLOCKINFO)}, hash, dbInfo.NewDirBlockInfo())
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(*dbInfo.DirBlockInfo), nil
}

// FetchDirBlockInfoByKeyMR gets a dirblock info block by keyMR from the database.
func (db *Overlay) FetchDirBlockInfoByKeyMR(hash interfaces.IHash) (*dbInfo.DirBlockInfo, error) {
	block, err := db.FetchBlock([]byte{byte(DIRBLOCKINFO)}, hash, dbInfo.NewDirBlockInfo())
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(*dbInfo.DirBlockInfo), nil
}

// FetchAllDirBlockInfos gets all of the dirblock info blocks
func (db *Overlay) FetchAllDirBlockInfos() ([]*dbInfo.DirBlockInfo, error) {
	list, err := db.FetchAllBlocksFromBucket([]byte{byte(DIRBLOCKINFO)}, dbInfo.NewDirBlockInfo())
	if err != nil {
		return nil, err
	}
	return toDirBlockInfosList(list), nil
}

func toDirBlockInfosList(source []interfaces.BinaryMarshallableAndCopyable) []*dbInfo.DirBlockInfo {
	answer := make([]*dbInfo.DirBlockInfo, len(source))
	for i, v := range source {
		answer[i] = v.(*dbInfo.DirBlockInfo)
	}
	sort.Sort(util.ByDirBlockInfoIDAccending(answer))
	return answer
}

func (db *Overlay) SaveDirBlockInfoHead(block interfaces.DatabaseBatchable) error {
	return db.ProcessDirBlockInfoBatch(block)
}

func (db *Overlay) FetchDirBlockInfoHead() (*dbInfo.DirBlockInfo, error) {
	blk := dbInfo.NewDirBlockInfo()
	block, err := db.FetchChainHeadByChainID([]byte{byte(DIRBLOCKINFO)}, primitives.NewHash(blk.GetChainID()), blk)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(*dbInfo.DirBlockInfo), nil
}
