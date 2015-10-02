// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay

import (
	/*"bytes"
	"encoding/binary"
	"errors"
	"github.com/FactomProject/factomd/btcd/wire"*/
	//. "github.com/FactomProject/factomd/common/DirectoryBlock"
	//. "github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
)
/*
// FetchDBEntriesFromQueue gets all of the dbentries that have not been processed
/*func (db *Overlay) FetchDBEntriesFromQueue(startTime *[]byte) (dbentries []*DBEntry, err error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var fromkey []byte = []byte{byte(TBL_EB_QUEUE)} // Table Name (1 bytes)
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
*//*
// ProcessDBlockBatche inserts the DBlock and update all it's dbentries in DB
func (db *Overlay) ProcessDBlockBatch(dblock *DirectoryBlock) error {

	if dblock != nil {
		if db.lbatch == nil {
			db.lbatch = new(leveldb.Batch)
		}

		defer db.lbatch.Reset()

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

		// Insert the binary directory block
		var key []byte = []byte{byte(TBL_DB)}
		key = append(key, dblock.DBHash.Bytes()...)
		db.lbatch.Put(key, binaryDblock)

		// Insert block height cross reference
		var dbNumkey []byte = []byte{byte(TBL_DB_NUM)}
		var buf bytes.Buffer
		binary.Write(&buf, binary.BigEndian, dblock.Header.DBHeight)
		dbNumkey = append(dbNumkey, buf.Bytes()...)
		db.lbatch.Put(dbNumkey, dblock.DBHash.Bytes())

		// Insert the directory block merkle root cross reference
		key = []byte{byte(TBL_DB_MR)}
		key = append(key, dblock.KeyMR.Bytes()...)
		binaryDBHash, _ := dblock.DBHash.MarshalBinary()
		db.lbatch.Put(key, binaryDBHash)

		// Update the chain head reference
		key = []byte{byte(TBL_CHAIN_HEAD)}
		key = append(key, D_CHAINID...)
		db.lbatch.Put(key, dblock.KeyMR.Bytes())

		err = db.lDb.Write(db.lbatch, db.wo)
		if err != nil {
			return err
		}

		// Update DirBlock Height cache
		db.lastDirBlkHeight = int64(dblock.Header.DBHeight)
		db.lastDirBlkSha, _ = NewShaHash(dblock.DBHash.Bytes())
		db.lastDirBlkShaCached = true

	}
	return nil
}*/

// UpdateBlockHeightCache updates the dir block height cache in db
func (db *Overlay) UpdateBlockHeightCache(dirBlkHeigh uint32, dirBlkHash IHash) error {
	// Update DirBlock Height cache
	db.lastDirBlkHeight = int64(dirBlkHeigh)
	db.lastDirBlkSha, _ = NewShaHash(dirBlkHash.Bytes())
	db.lastDirBlkShaCached = true
	return nil
}
/*
// FetchBlockHeightCache returns the hash and block height of the most recent
func (db *Overlay) FetchBlockHeightCache() (sha IHash, height int64, err error) {
	return db.lastDirBlkSha, db.lastDirBlkHeight, nil
}
// FetchHeightRange looks up a range of blocks by the start and ending
// heights.  Fetch is inclusive of the start height and exclusive of the
// ending height. To fetch all hashes from the start height until no
// more are present, use the special id `AllShas'.
func (db *Overlay) FetchHeightRange(startHeight, endHeight int64) (rshalist []IHash, err error) {

	var endidx int64
	if endHeight == database.AllShas {
		endidx = startHeight + wire.MaxBlocksPerMsg
	} else {
		endidx = endHeight
	}

	shalist := make([]IHash, 0, endidx-startHeight)
	for height := startHeight; height < endidx; height++ {
		// TODO(drahn) fix blkFile from height

		dbhash, lerr := db.FetchDBHashByHeight(uint32(height))
		if lerr != nil || dbhash == nil {
			break
		}

		shalist = append(shalist, dbhash)
	}

	if err != nil {
		return
	}
	//log.Tracef("FetchIdxRange idx %v %v returned %v shas err %v", startHeight, endHeight, len(shalist), err)

	return shalist, nil
}

// FetchBlockHeightBySha returns the block height for the given hash.  This is
// part of the database.Db interface implementation.
func (db *Overlay) FetchBlockHeightBySha(sha IHash) (int64, error) {

	dblk, _ := db.FetchDBlockByHash(sha)

	var height int64 = -1
	if dblk != nil {
		height = int64(dblk.Header.DBHeight)
	}

	return height, nil
}

// Insert the Directory Block meta data into db
func (db *Overlay) InsertDirBlockInfo(dirBlockInfo *DirBlockInfo) (err error) {
	if dirBlockInfo.BTCTxHash == nil {
		return
	}

	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	if db.lbatch == nil {
		db.lbatch = new(leveldb.Batch)
	}
	defer db.lbatch.Reset()

	var key []byte = []byte{byte(TBL_DB_INFO)} // Table Name (1 bytes)
	key = append(key, dirBlockInfo.DBHash.Bytes()...)
	binaryDirBlockInfo, _ := dirBlockInfo.MarshalBinary()
	db.lbatch.Put(key, binaryDirBlockInfo)

	err = db.lDb.Write(db.lbatch, db.wo)
	if err != nil {
		return err
	}

	return nil
}

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
}

// FetchDBlock gets an entry by hash from the database.
func (db *Overlay) FetchDBlockByHash(dBlockHash IHash) (*DirectoryBlock, error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var key []byte = []byte{byte(TBL_DB)}
	key = append(key, dBlockHash.Bytes()...)
	data, _ := db.lDb.Get(key, db.ro)

	dBlock := NewDBlock()
	if data == nil {
		return nil, errors.New("DBlock not found for Hash: " + dBlockHash.String())
	} else {
		_, err := dBlock.UnmarshalBinaryData(data)
		if err != nil {
			return nil, err
		}
	}

	dBlock.DBHash = dBlockHash

	return dBlock, nil
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
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var key []byte = []byte{byte(TBL_DB_NUM)}
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, dBlockHeight)
	key = append(key, buf.Bytes()...)
	data, err := db.lDb.Get(key, db.ro)
	if err != nil {
		return nil, err
	}

	dBlockHash := NewZeroHash()
	_, err = dBlockHash.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}

	return dBlockHash, nil
}

// FetchDBHashByMR gets a DBHash by MR from the database.
func (db *Overlay) FetchDBHashByMR(dBMR IHash) (IHash, error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var key []byte = []byte{byte(TBL_DB_MR)}
	key = append(key, dBMR.Bytes()...)
	data, err := db.lDb.Get(key, db.ro)
	if err != nil {
		return nil, err
	}

	dBlockHash := NewZeroHash()
	_, err = dBlockHash.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}

	return dBlockHash, nil
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
func (db *Overlay) FetchHeadMRByChainID(chainID IHash) (blkMR IHash, err error) {
	if chainID == nil {
		return nil, nil
	}

	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var key []byte = []byte{byte(TBL_CHAIN_HEAD)}
	key = append(key, chainID.Bytes()...)
	data, err := db.lDb.Get(key, db.ro)
	if err != nil {
		return nil, err
	}

	blkMR = NewZeroHash()
	_, err = blkMR.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}

	return blkMR, nil
}

// FetchAllDBlocks gets all of the fbInfo
func (db *Overlay) FetchAllDBlocks() (dBlocks []DirectoryBlock, err error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var fromkey []byte = []byte{byte(TBL_DB)}   // Table Name (1 bytes)						// Timestamp  (8 bytes)
	var tokey []byte = []byte{byte(TBL_DB + 1)} // Table Name (1 bytes)

	dBlockSlice := make([]DirectoryBlock, 0, 10)

	iter := db.lDb.NewIterator(&util.Range{Start: fromkey, Limit: tokey}, db.ro)

	for iter.Next() {
		var dBlock DirectoryBlock
		_, err := dBlock.UnmarshalBinaryData(iter.Value())
		if err != nil {
			return nil, err
		}
		//TODO: to be optimized??
		dBlock.DBHash = Sha(iter.Value())

		dBlockSlice = append(dBlockSlice, dBlock)

	}
	iter.Release()
	err = iter.Error()

	return dBlockSlice, nil
}

// FetchAllDirBlockInfo gets all of the dirBlockInfo
func (db *Overlay) FetchAllDirBlockInfo() (dirBlockInfoMap map[string]*DirBlockInfo, err error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var fromkey []byte = []byte{byte(TBL_DB_INFO)}   // Table Name (1 bytes)
	var tokey []byte = []byte{byte(TBL_DB_INFO + 1)} // Table Name (1 bytes)

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
}

// FetchAllUnconfirmedDirBlockInfo gets all of the dirBlockInfos that have BTC Anchor confirmation
func (db *Overlay) FetchAllUnconfirmedDirBlockInfo() (dirBlockInfoMap map[string]*DirBlockInfo, err error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var fromkey []byte = []byte{byte(TBL_DB_INFO)}   // Table Name (1 bytes)
	var tokey []byte = []byte{byte(TBL_DB_INFO + 1)} // Table Name (1 bytes)

	dirBlockInfoMap = make(map[string]*DirBlockInfo)

	iter := db.lDb.NewIterator(&util.Range{Start: fromkey, Limit: tokey}, db.ro)

	for iter.Next() {
		dBInfo := new(DirBlockInfo)

		// The last byte stores the confirmation flag
		if iter.Value()[len(iter.Value())-1] == 0 {
			_, err := dBInfo.UnmarshalBinaryData(iter.Value())
			if err != nil {
				return dirBlockInfoMap, err
			}
			dirBlockInfoMap[dBInfo.DBMerkleRoot.String()] = dBInfo
		}
	}
	iter.Release()
	err = iter.Error()
	return dirBlockInfoMap, err
}
*/