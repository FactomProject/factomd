// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay

import (
	/*"encoding/binary"
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
/*
type tTxInsertData struct {
	txsha   IHash
	blockid int64
	txoff   int
	txlen   int
	usedbuf []byte
}*/

type Overlay struct {
	// leveldb pieces
	DB IDatabase
	
	lastDirBlkShaCached bool
	lastDirBlkSha       IHash
	lastDirBlkHeight    int64
}

func NewOverlay(db IDatabase) *Overlay{
	answer:=new(Overlay)
	answer.DB = db
	return answer
}

/*
var CurrentDBVersion int32 = 1

func OpenLevelDB(dbpath string, create bool) (pbdb database.Db, err error) {
	return openDB(dbpath, create)
}

func openDB(dbpath string, create bool) (pbdb database.Db, err error) {
	var db Overlay
	var tlDb *leveldb.DB
	var dbversion int32

	defer func() {
		if err == nil {
			db.lDb = tlDb

			// Initialize db
			db.lastDirBlkHeight = -1

			pbdb = &db
		}
	}()

	if create == true {
		err = os.MkdirAll(dbpath, 0750)
		if err != nil {
			log.Println("mkdir failed %v %v", dbpath, err)
			return
		}
	} else {
		_, err = os.Stat(dbpath)
		if err != nil {
			return
		}
	}

	needVersionFile := false
	verfile := dbpath + ".ver"
	fi, ferr := os.Open(verfile)
	if ferr == nil {
		defer fi.Close()

		ferr = binary.Read(fi, binary.BigEndian, &dbversion)
		if ferr != nil {
			dbversion = ^0
		}
	} else {
		if create == true {
			needVersionFile = true
			dbversion = CurrentDBVersion
		}
	}

	opts := &opt.Options{
		Compression: opt.NoCompression,
	}

	switch dbversion {
	case 0:
		opts = &opt.Options{}
	case 1:
		// uses defaults from above
	default:
		err = fmt.Errorf("unsupported db version %v", dbversion)
		return
	}

	tlDb, err = leveldb.OpenFile(dbpath, opts)
	if err != nil {
		return
	}

	// If we opened the database successfully on 'create'
	// update the
	if needVersionFile {
		fo, ferr := os.Create(verfile)
		if ferr != nil {
			// TODO(design) close and delete database?
			err = ferr
			return
		}
		defer fo.Close()
		err = binary.Write(fo, binary.BigEndian, dbversion)
		if err != nil {
			return
		}
	}

	return
}

func (db *Overlay) close() error {
	return db.lDb.Close()
}

// Sync verifies that the database is coherent on disk,
// and no outstanding transactions are in flight.
func (db *Overlay) Sync() error {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	// while specified by the API, does nothing
	// however does grab lock to verify it does not return until other operations are complete.
	return nil
}

// Close cleanly shuts down database, syncing all data.
func (db *Overlay) Close() error {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	return db.close()
}

func int64ToKey(keyint int64) []byte {
	key := strconv.FormatInt(keyint, 10)
	return []byte(key)
}

func shaBlkToKey(sha IHash) []byte {
	shaB := sha.Bytes()
	return shaB
}

func shaTxToKey(sha IHash) []byte {
	shaB := sha.Bytes()
	shaB = append(shaB, "tx"...)
	return shaB
}

func shaSpentTxToKey(sha IHash) []byte {
	shaB := sha.Bytes()
	shaB = append(shaB, "sx"...)
	return shaB
}

func (db *Overlay) lBatch() *leveldb.Batch {
	if db.lbatch == nil {
		db.lbatch = new(leveldb.Batch)
	}
	return db.lbatch
}

func (db *Overlay) RollbackClose() error {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	return db.close()
}
*/