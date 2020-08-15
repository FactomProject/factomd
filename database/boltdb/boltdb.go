// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package boltdb

import (
	"fmt"
	"sync"

	"os"
	"path/filepath"

	"github.com/FactomProject/bolt"
	"github.com/PaulSnow/factom2d/common/interfaces"
)

// This database stores and retrieves interfaces.IBlock instances.  To do that, it
// needs a list of buckets that the using function wants, so it can make sure
// all those buckets exist.  (Avoids checking and building buckets in every
// write).
//
// It also needs a map of a hash to a interfaces.IBlock instance.  To support this,
// every block needs to be able to give the database a Hash for its type.
// This has to match the reverse, where looking up the hash gives the
// database the type for the hash.  This way, the database can marshal
// and unmarshal interfaces.IBlocks for storage in the database.  And since the interfaces.IBlocks
// can provide the hash, we don't need two maps.  Just the Hash to the
// interfaces.IBlock.
//
// Lastly it needs a filename with a full path.  If none is specified, it will
// use "/tmp/bolt_my.db".  Not the best idea to let this code default.
//
type BoltDB struct {
	Sem sync.RWMutex
	db  *bolt.DB // Pointer to the bolt db
}

var _ interfaces.IDatabase = (*BoltDB)(nil)

func NewBoltDB(bucketList [][]byte, filename string) *BoltDB {
	db := new(BoltDB)
	db.Init(bucketList, filename)
	return db
}

func NewAndCreateBoltDB(bucketList [][]byte, filename string) *BoltDB {
	err := os.MkdirAll(filepath.Dir(filename), 0750)
	if err != nil {
		if err != nil {
			panic("Database could not be created, " + err.Error())
		}
	}
	return NewBoltDB(bucketList, filename)
}

/***************************************
 *       Methods
 ***************************************/

func (db *BoltDB) ListAllBuckets() ([][]byte, error) {
	db.Sem.RLock()
	defer db.Sem.RUnlock()

	answer := [][]byte{}
	db.db.View(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		k, _ := c.First()
		for {
			if k == nil {
				break
			}
			answer = append(answer, k)
			k, _ = c.Next()
		}
		return nil
	})

	return answer, nil
}

// We don't care if delete works or not.  If the key isn't there, that's ok
func (db *BoltDB) Delete(bucket []byte, key []byte) error {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	db.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		b := tx.Bucket(bucket)
		b.Delete(key)
		return nil
	})
	return nil
}

// Can't trim a real database
func (db *BoltDB) Trim() {
}

func (db *BoltDB) Close() error {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	db.db.Close()
	return nil
}

func (db *BoltDB) Get(bucket []byte, key []byte, destination interfaces.BinaryMarshallable) (interfaces.BinaryMarshallable, error) {
	db.Sem.RLock()
	defer db.Sem.RUnlock()

	var v []byte
	db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		v = b.Get(key)
		if v == nil {
			return nil
		}
		return nil
	})
	if v == nil { // If the value is undefined, return nil
		return nil, nil
	}

	_, err := destination.UnmarshalBinaryData(v)
	if err != nil {
		return nil, err
	}
	return destination, nil
}

func (db *BoltDB) Put(bucket []byte, key []byte, data interfaces.BinaryMarshallable) error {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	hex, err := data.MarshalBinary()
	if err != nil {
		return err
	}
	err = db.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		b := tx.Bucket(bucket)
		err = b.Put(key, hex)
		return err
	})
	return err
}

func (db *BoltDB) PutInBatch(records []interfaces.Record) error {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	err := db.db.Batch(func(tx *bolt.Tx) error {
		for _, v := range records {
			_, err := tx.CreateBucketIfNotExists(v.Bucket)
			if err != nil {
				return err
			}
			b := tx.Bucket(v.Bucket)
			hex, err := v.Data.MarshalBinary()
			if err != nil {
				return err
			}
			err = b.Put(v.Key, hex)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *BoltDB) Clear(bucket []byte) error {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	err := db.db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucket)
		if err != nil {
			return fmt.Errorf("No bucket: %s", err)
		}
		return nil
	})
	return err
}

func (db *BoltDB) ListAllKeys(bucket []byte) (keys [][]byte, err error) {
	db.Sem.RLock()
	defer db.Sem.RUnlock()

	keys = make([][]byte, 0, 32)
	db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			//fmt.Println("bucket 0x" + hex.EncodeToString(bucket) + " not found")
		} else {
			b.ForEach(func(k, v []byte) error {
				keys = append(keys, k)
				return nil
			})
		}
		return nil
	})
	return
}

func (db *BoltDB) GetAll(bucket []byte, sample interfaces.BinaryMarshallableAndCopyable) ([]interfaces.BinaryMarshallableAndCopyable, [][]byte, error) {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	answer := []interfaces.BinaryMarshallableAndCopyable{}
	keys := [][]byte{}
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			//fmt.Println("bucket 0x" + hex.EncodeToString(bucket) + " not found")
		} else {
			b.ForEach(func(k, v []byte) error {
				tmp := sample.New()
				err := tmp.UnmarshalBinary(v)
				if err != nil {
					return err
				}
				keys = append(keys, k)
				answer = append(answer, tmp)
				return nil
			})
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return answer, keys, nil
}

// We have to make accommodation for many Init functions.  But what we really
// want here is:
//
//      Init(bucketList [][]byte, filename string)
//
func (db *BoltDB) Init(bucketList [][]byte, filename string) {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	if db.db == nil {
		if filename == "" {
			filename = "/tmp/bolt_my.db"
		}

		tdb, err := bolt.Open(filename, 0600, nil)
		if err != nil {
			panic("Database was not found, and could not be created.")
		}

		db.db = tdb
	}

	for _, bucket := range bucketList {
		db.db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists(bucket)
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}
			return nil
		})
	}
}

func (db *BoltDB) DoesKeyExist(bucket, key []byte) (bool, error) {
	db.Sem.RLock()
	defer db.Sem.RUnlock()

	var v []byte
	db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		v = b.Get(key)
		if v == nil {
			return nil
		}
		return nil
	})
	if v == nil { // If the value is undefined, return nil
		return false, nil
	}

	return true, nil
}
