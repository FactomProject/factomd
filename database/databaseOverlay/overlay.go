// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay

import (
	"encoding/binary"
	/*
		"fmt"
		"log"
		"os"
		"strconv"
		"sync"

		"github.com/FactomProject/factomd/database"*/

	. "github.com/FactomProject/factomd/common/interfaces"
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

func (db *Overlay) FetchBlockByHash(bucket []byte, key IHash, dst BinaryMarshallable) (BinaryMarshallable, error) {
	block, err := db.DB.Get(bucket, key.Bytes(), dst)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block, nil
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
	GetDBHeight() uint32
	GetHash() IHash //block.GetHash().Bytes()
	GetChainID() []byte
}

func (db *Overlay) ProcessBlockBatch(blockBucket, numberBucket []byte, block DatabaseBatchable) error {
	if block == nil {
		return nil
	}

	batch := []Record{}

	batch = append(batch, Record{blockBucket, block.GetHash().Bytes(), block})

	// Insert the sc block number cross reference
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, block.GetDBHeight())
	batch = append(batch, Record{numberBucket, bytes, block.GetHash()})

	batch = append(batch, Record{[]byte{TBL_CHAIN_HEAD}, block.GetChainID(), block.GetHash()})

	err := db.DB.PutInBatch(batch)
	if err != nil {
		return err
	}

	return nil
}
