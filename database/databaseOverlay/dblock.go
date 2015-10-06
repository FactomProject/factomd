// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay

import (
	/*"bytes"
	"errors"*/
	//TODO: remove dependency on /wire
	"github.com/FactomProject/factomd/btcd/wire"
	. "github.com/FactomProject/factomd/common/DirectoryBlock"
	. "github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"

	"encoding/binary"
)

// AllShas is a special value that can be used as the final sha when requesting
// a range of shas by height to request them all.
const AllShas = int64(^uint64(0) >> 1)

/*
// FetchDBEntriesFromQueue gets all of the dbentries that have not been processed
/*func (db *Overlay) FetchDBEntriesFromQueue(startTime *[]byte) (dbentries []*DBEntry, err error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var fromkey []byte = []byte{byte(TBL_EB_QUEUE)}
	fromkey = append(fromkey, *startTime...)        // Timestamp  (8 bytes)

	var tokey []byte = []byte{byte(TBL_EB_QUEUE)} // Table Name (4 bytes)
	binaryTimestamp := make([]byte, 8)
	binary.BigEndian.PutUint64(binaryTimestamp, uint64(time.Now().Unix()))
	tokey = append(tokey, binaryTimestamp...) // Timestamp  (8 bytes)

	fbEntrySlice := make([]*DBEntry, 0, 10)

	iter := db.lDb.NewIterator(&util.Range{Start: fromkey, Limit: tokey}, db.ro)

	for iter.Next() {
		if bytes.Equal(iter.Value(), []byte{byte(STATUS_IN_QUEUE)}) {
			key := make([]byte, len(iter.Key()))
			copy(key, iter.Key())
			dbEntry := new(DBEntry)

			dbEntry.SetTimestamp(key[1:9]) // Timestamp (8 bytes)
			cid := key[9:41]
			dbEntry.ChainID = new(Hash)
			dbEntry.ChainID.Bytes = cid // Chain id (32 bytes)
			dbEntry.SetHash(key[41:73]) // Entry Hash (32 bytes)

			fbEntrySlice = append(fbEntrySlice, dbEntry)
		}
	}
	iter.Release()
	err = iter.Error()

	return fbEntrySlice, nil
}
*/

// ProcessDBlockBatche inserts the DBlock and update all it's dbentries in DB
func (db *Overlay) ProcessDBlockBatch(dblock *DirectoryBlock) error {
	if dblock == nil {
		return nil
	}

	binaryDblock, err := dblock.MarshalBinary()
	if err != nil {
		return err
	}

	if dblock.DBHash == nil {
		dblock.DBHash = Sha(binaryDblock)
	}
	if dblock.KeyMR == nil {
		dblock.BuildKeyMerkleRoot()
	}

	batch := []Record{}

	// Insert the binary directory block
	batch = append(batch, Record{[]byte{byte(TBL_DB)}, dblock.DBHash.Bytes(), dblock})

	// Insert block height cross reference
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, dblock.Header.DBHeight)
	batch = append(batch, Record{[]byte{byte(TBL_DB_NUM)}, bytes, dblock.DBHash})

	// Insert the directory block merkle root cross reference
	batch = append(batch, Record{[]byte{byte(TBL_DB_MR)}, dblock.KeyMR.Bytes(), dblock.DBHash})

	// Update the chain head reference
	batch = append(batch, Record{[]byte{byte(TBL_CHAIN_HEAD)}, D_CHAINID, dblock.KeyMR})

	err = db.DB.PutInBatch(batch)
	if err != nil {
		return err
	}
	// Update DirBlock Height cache
	db.lastDirBlkHeight = int64(dblock.Header.DBHeight)
	db.lastDirBlkSha, _ = NewShaHash(dblock.DBHash.Bytes())
	db.lastDirBlkShaCached = true

	return nil
}

// UpdateBlockHeightCache updates the dir block height cache in db
func (db *Overlay) UpdateBlockHeightCache(dirBlkHeigh uint32, dirBlkHash IHash) error {
	// Update DirBlock Height cache
	db.lastDirBlkHeight = int64(dirBlkHeigh)
	db.lastDirBlkSha, _ = NewShaHash(dirBlkHash.Bytes())
	db.lastDirBlkShaCached = true
	return nil
}

// FetchBlockHeightCache returns the hash and block height of the most recent
func (db *Overlay) FetchBlockHeightCache() (sha IHash, height int64, err error) {
	return db.lastDirBlkSha, db.lastDirBlkHeight, nil
}

// FetchHeightRange looks up a range of blocks by the start and ending
// heights.  Fetch is inclusive of the start height and exclusive of the
// ending height. To fetch all hashes from the start height until no
// more are present, use the special id `AllShas'.
func (db *Overlay) FetchHeightRange(startHeight, endHeight int64) ([]IHash, error) {
	var endidx int64
	if endHeight == AllShas {
		endidx = startHeight + wire.MaxBlocksPerMsg
	} else {
		endidx = endHeight
	}

	shalist := make([]IHash, 0, endidx-startHeight)
	for height := startHeight; height < endidx; height++ {
		// TODO(drahn) fix blkFile from height

		dbhash, err := db.FetchDBHashByHeight(uint32(height))
		if err != nil {
			return nil, err
		}
		if dbhash == nil {
			break
		}

		shalist = append(shalist, dbhash)
	}
	//log.Tracef("FetchIdxRange idx %v %v returned %v shas err %v", startHeight, endHeight, len(shalist), err)

	return shalist, nil
}

// FetchBlockHeightBySha returns the block height for the given hash.  This is
// part of the database.Db interface implementation.
func (db *Overlay) FetchBlockHeightBySha(sha IHash) (int64, error) {
	dblk, err := db.FetchDBlockByHash(sha)
	if err != nil {
		return -1, err
	}

	var height int64 = -1
	if dblk != nil {
		height = int64(dblk.Header.DBHeight)
	}

	return height, nil
}

// Insert the Directory Block meta data into db
func (db *Overlay) InsertDirBlockInfo(dirBlockInfo *DirBlockInfo) error {
	if dirBlockInfo.BTCTxHash == nil {
		return nil
	}

	bucket := []byte{byte(TBL_DB_INFO)}
	key := dirBlockInfo.DBHash.Bytes()

	err := db.DB.Put(bucket, key, dirBlockInfo)
	if err != nil {
		return err
	}

	return nil
}

/*
// FetchDirBlockInfoByHash gets an DirBlockInfo obj
func (db *Overlay) FetchDirBlockInfoByHash(dbHash IHash) (dirBlockInfo *DirBlockInfo, err error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var key []byte = []byte{byte(TBL_DB_INFO)}
	key = append(key, dbHash.Bytes()...)
	data, err := db.lDb.Get(key, db.ro)

	if data != nil {
		dirBlockInfo = new(DirBlockInfo)
		_, err := dirBlockInfo.UnmarshalBinaryData(data)
		if err != nil {
			return nil, err
		}
	}

	return dirBlockInfo, nil
}*/

// FetchDBlock gets an entry by hash from the database.
func (db *Overlay) FetchDBlockByHash(dBlockHash IHash) (*DirectoryBlock, error) {
	bucket := []byte{byte(TBL_DB)}
	key := dBlockHash.Bytes()

	block, err := db.DB.Get(bucket, key, new(DirectoryBlock))
	if err != nil {
		return nil, err
	}
	return block.(*DirectoryBlock), nil
}

// FetchDBlockByHeight gets an directory block by height from the database.
func (db *Overlay) FetchDBlockByHeight(dBlockHeight uint32) (dBlock *DirectoryBlock, err error) {
	dBlockHash, err := db.FetchDBHashByHeight(dBlockHeight)
	if err != nil {
		return nil, err
	}

	if dBlockHash != nil {
		dBlock, err = db.FetchDBlockByHash(dBlockHash)
		if err != nil {
			return nil, err
		}
	}

	return dBlock, nil
}

// FetchDBHashByHeight gets a dBlockHash from the database.
func (db *Overlay) FetchDBHashByHeight(dBlockHeight uint32) (IHash, error) {
	bucket := []byte{byte(TBL_DB_NUM)}
	key := make([]byte, 4)
	binary.BigEndian.PutUint32(key, dBlockHeight)

	block, err := db.DB.Get(bucket, key, new(Hash))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(*Hash), nil
}

// FetchDBHashByMR gets a DBHash by MR from the database.
func (db *Overlay) FetchDBHashByMR(dBMR IHash) (IHash, error) {
	bucket := []byte{byte(TBL_DB_MR)}
	key := dBMR.Bytes()

	block, err := db.DB.Get(bucket, key, new(Hash))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(*Hash), nil
}

// FetchDBlockByMR gets a directory block by merkle root from the database.
func (db *Overlay) FetchDBlockByMR(dBMR IHash) (*DirectoryBlock, error) {
	dBlockHash, err := db.FetchDBHashByMR(dBMR)
	if err != nil {
		return nil, err
	}

	dBlock, err := db.FetchDBlockByHash(dBlockHash)
	if err != nil {
		return dBlock, err
	}

	return dBlock, nil
}

// FetchHeadMRByChainID gets a MR of the highest block from the database.
func (db *Overlay) FetchHeadMRByChainID(chainID IHash) (IHash, error) {
	if chainID == nil {
		return nil, nil
	}

	bucket := []byte{byte(TBL_CHAIN_HEAD)}
	key := chainID.Bytes()

	block, err := db.DB.Get(bucket, key, new(Hash))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(*Hash), nil
}

// FetchAllDBlocks gets all of the fbInfo
func (db *Overlay) FetchAllDBlocks() (dBlocks []*DirectoryBlock, err error) {
	bucket := []byte{byte(TBL_DB)}

	list, err := db.DB.GetAll(bucket, new(DirectoryBlock))
	if err != nil {
		return nil, err
	}
	answer := make([]*DirectoryBlock, len(list))
	for i, v := range list {
		answer[i] = v.(*DirectoryBlock)
	}
	return answer, nil
}

/*
// FetchAllDirBlockInfo gets all of the dirBlockInfo
func (db *Overlay) FetchAllDirBlockInfo() (dirBlockInfoMap map[string]*DirBlockInfo, err error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var fromkey []byte = []byte{byte(TBL_DB_INFO)}
	var tokey []byte = []byte{byte(TBL_DB_INFO + 1)}

	dirBlockInfoMap = make(map[string]*DirBlockInfo)

	iter := db.lDb.NewIterator(&util.Range{Start: fromkey, Limit: tokey}, db.ro)

	for iter.Next() {
		dBInfo := new(DirBlockInfo)
		_, err := dBInfo.UnmarshalBinaryData(iter.Value())
		if err != nil {
			return nil, err
		}
		dirBlockInfoMap[dBInfo.DBMerkleRoot.String()] = dBInfo
	}
	iter.Release()
	err = iter.Error()
	return dirBlockInfoMap, err
}*/

// FetchAllUnconfirmedDirBlockInfo gets all of the dirBlockInfos that have BTC Anchor confirmation
func (db *Overlay) FetchAllUnconfirmedDirBlockInfo() (dirBlockInfoMap map[string]*DirBlockInfo, err error) {
	bucket := []byte{byte(TBL_DB_INFO)}

	all, err := db.DB.GetAll(bucket, new(DirBlockInfo))
	if err != nil {
		return nil, err
	}

	dirBlockInfoMap = make(map[string]*DirBlockInfo)

	for _, v := range all {
		dBInfo := v.(*DirBlockInfo)
		if dBInfo.BTCConfirmed == false {
			dirBlockInfoMap[dBInfo.DBMerkleRoot.String()] = dBInfo
		}
	}

	return dirBlockInfoMap, nil
}
