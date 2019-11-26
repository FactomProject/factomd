package databaseOverlay

import (
	"sort"

	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/util"
)

// ProcessDirBlockInfoBatch inserts the dirblock info block
func (db *Overlay) ProcessDirBlockInfoBatch(block interfaces.IDirBlockInfo) error {
	var err error
	if block.GetBTCConfirmed() == true {
		err := db.Delete(DIRBLOCKINFO_UNCONFIRMED, block.DatabasePrimaryIndex().Bytes())
		if err != nil {
			return err
		}
		err = db.ProcessBlockBatchWithoutHead(DIRBLOCKINFO, DIRBLOCKINFO_NUMBER, DIRBLOCKINFO_SECONDARYINDEX, block)
	} else {
		err = db.ProcessBlockBatchWithoutHead(DIRBLOCKINFO_UNCONFIRMED, DIRBLOCKINFO_NUMBER, DIRBLOCKINFO_SECONDARYINDEX, block)
	}
	if err != nil {
		db.parentState.GetEvents().EmitDBAnchorEvent(block)
	}
	return err
}

func (db *Overlay) ProcessDirBlockInfoMultiBatch(block interfaces.IDirBlockInfo) error {
	var err error = nil
	if block.GetBTCConfirmed() == true {
		err = db.Delete(DIRBLOCKINFO_UNCONFIRMED, block.DatabasePrimaryIndex().Bytes())
		if err != nil {
			return err
		}
		err = db.ProcessBlockMultiBatchWithoutHead(DIRBLOCKINFO, DIRBLOCKINFO_NUMBER, DIRBLOCKINFO_SECONDARYINDEX, block)
	} else {
		err = db.ProcessBlockMultiBatchWithoutHead(DIRBLOCKINFO_UNCONFIRMED, DIRBLOCKINFO_NUMBER, DIRBLOCKINFO_SECONDARYINDEX, block)
	}
	if err != nil {
		db.parentState.GetEvents().EmitDBAnchorEvent(block)
	}
	return err
}

// FetchDirBlockInfoByHash gets a dirblock info block by hash from the database.
func (db *Overlay) FetchDirBlockInfoByHash(hash interfaces.IHash) (interfaces.IDirBlockInfo, error) {
	block, err := db.FetchBlockBySecondaryIndex(DIRBLOCKINFO_SECONDARYINDEX, DIRBLOCKINFO_UNCONFIRMED, hash, dbInfo.NewDirBlockInfo())
	if err != nil {
		return nil, err
	}
	if block == nil {
		block, err = db.FetchBlockBySecondaryIndex(DIRBLOCKINFO_SECONDARYINDEX, DIRBLOCKINFO, hash, dbInfo.NewDirBlockInfo())
		if err != nil {
			return nil, err
		}
		if block == nil {
			return nil, nil
		}
	}
	return block.(interfaces.IDirBlockInfo), nil
}

// FetchDirBlockInfoByKeyMR gets a dirblock info block by keyMR from the database.
func (db *Overlay) FetchDirBlockInfoByKeyMR(hash interfaces.IHash) (interfaces.IDirBlockInfo, error) {
	block, err := db.FetchBlock(DIRBLOCKINFO_UNCONFIRMED, hash, dbInfo.NewDirBlockInfo())
	if err != nil {
		return nil, err
	}
	if block == nil {
		block, err = db.FetchBlock(DIRBLOCKINFO, hash, dbInfo.NewDirBlockInfo())
		if err != nil {
			return nil, err
		}
		if block == nil {
			return nil, nil
		}
	}
	return block.(interfaces.IDirBlockInfo), nil
}

// FetchAllConfirmedDirBlockInfos gets all of the confiemed dirblock info blocks
func (db *Overlay) FetchAllConfirmedDirBlockInfos() ([]interfaces.IDirBlockInfo, error) {
	list, err := db.FetchAllBlocksFromBucket(DIRBLOCKINFO, dbInfo.NewDirBlockInfo())
	if err != nil {
		return nil, err
	}
	return toDirBlockInfosList(list), nil
}

// FetchAllUnconfirmedDirBlockInfos gets all of the unconfirmed dirblock info blocks
func (db *Overlay) FetchAllUnconfirmedDirBlockInfos() ([]interfaces.IDirBlockInfo, error) {
	list, err := db.FetchAllBlocksFromBucket(DIRBLOCKINFO_UNCONFIRMED, dbInfo.NewDirBlockInfo())
	if err != nil {
		return nil, err
	}
	return toDirBlockInfosList(list), nil
}

// FetchAllDirBlockInfos gets all of the dirblock info blocks
func (db *Overlay) FetchAllDirBlockInfos() ([]interfaces.IDirBlockInfo, error) {
	unconfirmed, err := db.FetchAllUnconfirmedDirBlockInfos()
	if err != nil {
		return nil, err
	}
	confirmed, err := db.FetchAllConfirmedDirBlockInfos()
	if err != nil {
		return nil, err
	}
	all := append(unconfirmed, confirmed...)
	sort.Sort(util.ByDirBlockInfoIDAscending(all))
	return all, nil
}

func toDirBlockInfosList(source []interfaces.BinaryMarshallableAndCopyable) []interfaces.IDirBlockInfo {
	answer := make([]interfaces.IDirBlockInfo, len(source))
	for i, v := range source {
		answer[i] = v.(interfaces.IDirBlockInfo)
	}
	sort.Sort(util.ByDirBlockInfoIDAscending(answer))
	return answer
}

func (db *Overlay) SaveDirBlockInfo(block interfaces.IDirBlockInfo) error {
	return db.ProcessDirBlockInfoBatch(block)
}
