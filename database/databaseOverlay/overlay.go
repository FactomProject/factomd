// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay

import (
	"encoding/binary"
	"github.com/FactomProject/factomd/btcd/wire"
	. "github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
)

const (
	dbVersion     int = 2
	dbMaxTransCnt     = 20000
	dbMaxTransMem     = 64 * 1024 * 1024 // 64 MB
)

// the "table" prefix
const (

	// Directory Block
	TBL_DB uint8 = iota
	TBL_DB_NUM
	TBL_DB_MR
	TBL_DB_INFO

	// Admin Block
	TBL_AB //4
	TBL_AB_NUM

	TBL_SC
	TBL_SC_NUM

	// Entry Credit Block
	TBL_CB //8
	TBL_CB_NUM
	TBL_CB_MR

	// Entry Chain
	TBL_CHAIN_HASH //11

	// The latest Block MR for chains including special chains
	TBL_CHAIN_HEAD

	// Entry Block
	TBL_EB //13
	TBL_EB_CHAIN_NUM
	TBL_EB_MR

	//Entry
	TBL_ENTRY
)

// the process status in db
const (
	STATUS_IN_QUEUE uint8 = iota
	STATUS_PROCESSED
)

var currentChainType uint32 = 1

var isLookupDB bool = true // to be put in property file

type Overlay struct {
	// leveldb pieces
	DB IDatabase

	lastDirBlkShaCached bool
	lastDirBlkSha       IHash
	lastDirBlkHeight    int64
}

func (db *Overlay) Close() (err error) {
	return db.DB.Close()
}

func NewOverlay(db IDatabase) *Overlay {
	answer := new(Overlay)
	answer.DB = db

	answer.lastDirBlkHeight = -1

	return answer
}

func (db *Overlay) FetchBlockByHeight(heightBucket []byte, blockBucket []byte, blockHeight uint32, dst DatabaseBatchable) (DatabaseBatchable, error) {
	index, err := db.FetchBlockIndexByHeight(heightBucket, blockHeight)
	if err != nil {
		return nil, err
	}
	if index == nil {
		return nil, nil
	}
	return db.FetchBlock(blockBucket, index, dst)
}

func (db *Overlay) FetchBlockIndexByHeight(bucket []byte, blockHeight uint32) (IHash, error) {
	key := make([]byte, 4)
	binary.BigEndian.PutUint32(key, blockHeight)

	block, err := db.DB.Get(bucket, key, new(Hash))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(*Hash), nil
}

func (db *Overlay) FetchPrimaryIndexBySecondaryIndex(bucket []byte, key IHash) (IHash, error) {
	block, err := db.DB.Get(bucket, key.Bytes(), new(Hash))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(*Hash), nil
}

func (db *Overlay) FetchBlockBySecondaryIndex(secondaryIndexBucket, blockBucket []byte, index IHash, dst DatabaseBatchable) (DatabaseBatchable, error) {
	hash, err := db.FetchPrimaryIndexBySecondaryIndex(secondaryIndexBucket, index)
	if err != nil {
		return nil, err
	}
	if hash == nil {
		return nil, nil
	}
	return db.FetchBlock(blockBucket, hash, dst)
}

func (db *Overlay) FetchBlock(bucket []byte, key IHash, dst DatabaseBatchable) (DatabaseBatchable, error) {
	block, err := db.DB.Get(bucket, key.Bytes(), dst)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(DatabaseBatchable), nil
}

func (db *Overlay) FetchAllBlocksFromBucket(bucket []byte, sample BinaryMarshallableAndCopyable) ([]BinaryMarshallableAndCopyable, error) {
	answer, err := db.DB.GetAll(bucket, sample)
	if err != nil {
		return nil, err
	}
	return answer, nil
}

type DatabaseBatchable interface {
	BinaryMarshallable
	GetDatabaseHeight() uint32

	DatabasePrimaryIndex() IHash   //block.KeyMR()
	DatabaseSecondaryIndex() IHash //block.GetHash()

	GetChainID() []byte
}

func (db *Overlay) Insert(bucket []byte, entry DatabaseBatchable) error {
	err := db.DB.Put(bucket, entry.DatabasePrimaryIndex().Bytes(), entry)
	if err != nil {
		return err
	}
	return nil
}

func (db *Overlay) ProcessBlockBatch(blockBucket, numberBucket, secondaryIndexBucket []byte, block DatabaseBatchable) error {
	if block == nil {
		return nil
	}

	batch := []Record{}

	batch = append(batch, Record{blockBucket, block.DatabasePrimaryIndex().Bytes(), block})

	if numberBucket != nil {
		bytes := make([]byte, 4)
		binary.BigEndian.PutUint32(bytes, block.GetDatabaseHeight())
		batch = append(batch, Record{numberBucket, bytes, block.DatabasePrimaryIndex()})
	}

	if secondaryIndexBucket != nil {
		batch = append(batch, Record{secondaryIndexBucket, block.DatabaseSecondaryIndex().Bytes(), block.DatabasePrimaryIndex()})
	}

	batch = append(batch, Record{[]byte{TBL_CHAIN_HEAD}, block.GetChainID(), block.DatabasePrimaryIndex()})

	err := db.DB.PutInBatch(batch)
	if err != nil {
		return err
	}

	return nil
}

// FetchHeadMRByChainID gets an index of the highest block from the database.
func (db *Overlay) FetchHeadIndexByChainID(chainID IHash) (IHash, error) {
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

// AllShas is a special value that can be used as the final sha when requesting
// a range of shas by height to request them all.
const AllShas = int64(^uint64(0) >> 1)

func (db *Overlay) FetchBlockIndexesInHeightRange(numberBucket []byte, startHeight, endHeight int64) ([]IHash, error) {
	//TODO: deprecate AllShas
	var endidx int64
	if endHeight == AllShas {
		endidx = startHeight + wire.MaxBlocksPerMsg
	} else {
		endidx = endHeight
	}

	shalist := make([]IHash, 0, endidx-startHeight)
	for height := startHeight; height < endidx; height++ {
		dbhash, err := db.FetchBlockIndexByHeight(numberBucket, uint32(height))
		if err != nil {
			return nil, err
		}
		if dbhash == nil {
			break
		}

		shalist = append(shalist, dbhash)
	}

	return shalist, nil
}
