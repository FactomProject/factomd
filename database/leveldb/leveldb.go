// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package boltdb

import (
	"encoding/hex"
	"fmt"
	. "github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/goleveldb/leveldb"
	"github.com/FactomProject/goleveldb/leveldb/util"
	"sync"
)

type LevelDB struct {
	// lock preventing multiple entry
	dbLock sync.Mutex
	lDB    *leveldb.DB
	lbatch *leveldb.Batch
	ro     *opt.ReadOptions
}

var _ IDatabase = (*LevelDB)(nil)

func (db *LevelDB) Delete(bucket []byte, key []byte) error {
	//TODO: do
	return nil
}

func (db *LevelDB) Close() error {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	return db.lDB.Close()
}

func (db *LevelDB) Get(bucket []byte, key []byte, destination BinaryMarshallable) (BinaryMarshallable, error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	ldbKey := append(bucket, key)
	data, err := db.lDb.Get(ldbKey, db.ro)
	if err != nil {
		return nil, err
	}

	_, err = destination.UnmarshalBinaryData(v)
	if err != nil {
		return nil, err
	}

	return destination, nil
}

func (db *LevelDB) Put(bucket []byte, key []byte, data BinaryMarshallable) error {
	if db.lbatch == nil {
		db.lbatch = new(leveldb.Batch)
	}

	defer db.lbatch.Reset()

	ldbKey := append(bucket, key)
	db.lbatch.Put(ldbKey, data.Bytes())

	err = db.lDB.Write(db.lbatch, db.wo)
	if err != nil {
		log.Println("batch failed %v\n", err)
		return err
	}
}

func (db *LevelDB) PutInBatch(records []Record) error {
	if db.lbatch == nil {
		db.lbatch = new(leveldb.Batch)
	}

	defer db.lbatch.Reset()

	for _, v := range records {
		ldbKey := append(v.Bucket, v.Key)
		db.lbatch.Put(ldbKey, v.Data.Bytes())
	}

	err = db.lDB.Write(db.lbatch, db.wo)
	if err != nil {
		log.Println("batch failed %v\n", err)
		return err
	}
	return nil
}

func (db *LevelDB) Clear(bucket []byte) error {
	//TODO: add
	return nil
}

func (db *LevelDB) ListAllKeys(bucket []byte) (keys [][]byte, err error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var fromKey []byte = bucket[:]
	var toKey []byte = bucket[:]
	toKey = addOneToByteArray(toKey)

	iter := db.lDb.NewIterator(&util.Range{Start: fromKey, Limit: toKey}, db.ro)

	var answer [][]byte

	for iter.Next() {
		key := iter.Key()
		answer = append(answer, key[len(bucket):])
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		return err
	}

	return answer, nil
}

func NewLevelDB(filename string, create bool) (*IDatabase, error) {
	db := new(LevelDB)
	var err error

	var tlDB *leveldb.DB

	if create == true {
		err = os.MkdirAll(filename, 0750)
		if err != nil {
			log.Println("mkdir failed %v %v", filename, err)
			return nil, err
		}
	} else {
		_, err = os.Stat(filename)
		if err != nil {
			return nil, err
		}
	}

	opts := &opt.Options{
		Compression: opt.NoCompression,
	}

	tlDB, err = leveldb.OpenFile(filename, opts)
	if err != nil {
		return nil, err
	}
	db.lDB = tlDB

	return db, nil
}

// Internal db use only
func addOneToByteArray(input []byte) (output []byte) {
	if input == nil {
		return []byte{byte(1)}
	}
	output = make([]byte, len(input))
	copy(output, input)
	for i := len(input); i > 0; i-- {
		if output[i-1] <= 255 {
			output[i-1] = output[i-1] + 1
			break
		}
	}
	return output
}
