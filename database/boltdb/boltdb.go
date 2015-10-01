// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package boltdb

import (
	"encoding/hex"
	"fmt"
	. "github.com/FactomProject/factomd/common/interfaces"

	"github.com/boltdb/bolt"
)

var _ = hex.EncodeToString

// This database stores and retrieves IBlock instances.  To do that, it
// needs a list of buckets that the using function wants, so it can make sure
// all those buckets exist.  (Avoids checking and building buckets in every
// write).
//
// It also needs a map of a hash to a IBlock instance.  To support this,
// every block needs to be able to give the database a Hash for its type.
// This has to match the reverse, where looking up the hash gives the
// database the type for the hash.  This way, the database can marshal
// and unmarshal IBlocks for storage in the database.  And since the IBlocks
// can provide the hash, we don't need two maps.  Just the Hash to the
// IBlock.
//
// Lastly it needs a filename with a full path.  If none is specified, it will
// use "/tmp/bolt_my.db".  Not the best idea to let this code default.
//
type BoltDB struct {
	db *bolt.DB // Pointer to the bolt db
}

var _ IDatabase = (*BoltDB)(nil)

/***************************************
 *       Methods
 ***************************************/

// We don't care if delete works or not.  If the key isn't there, that's ok
func (d *BoltDB) Delete(bucket []byte, key []byte) error {
	d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		b.Delete(key)
		return nil
	})
	return nil
}

func (d *BoltDB) Close() error {
	d.db.Close()
	return nil
}

func (d *BoltDB) Get(bucket []byte, key []byte, destination BinaryMarshallable) (BinaryMarshallable, error) {
	var v []byte
	d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		v = b.Get(key)
		if v == nil {
			return nil
		}
		return nil
	})
	if v == nil || len(v) < 32 { // If the value is undefined, return nil
		return nil, nil
	}

	_, err := destination.UnmarshalBinaryData(v)
	if err != nil {
		return nil, err
	}
	return destination, nil
}

func (d *BoltDB) Put(bucket []byte, key []byte, data BinaryMarshallable) error {
	hex, err := data.MarshalBinary()
	if err != nil {
		return err
	}
	err = d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		err := b.Put(key, hex)
		return err
	})
	return err
}

func (db *BoltDB) PutInBatch(records []Record) error {
	//TODO: put in actual batch if possible
	for _, v := range records {
		err := db.Put(v.Bucket, v.Key, v.Data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *BoltDB) Clear(bucket []byte) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucket)
		if err != nil {
			return fmt.Errorf("No bucket: %s", err)
		}
		return nil
	})
	return err
}

func (bdb *BoltDB) ListAllKeys(bucket []byte) (keys [][]byte, err error) {
	keys = make([][]byte, 0, 32)
	bdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			fmt.Println("bucket '", bucket, "' not found")
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

// We have to make accomadation for many Init functions.  But what we really
// want here is:
//
//      Init(bucketList [][]byte, filename string)
//
func (d *BoltDB) Init(bucketList [][]byte, filename string) {

	if d.db == nil {
		if filename == "" {
			filename = "/tmp/bolt_my.db"
		}

		tdb, err := bolt.Open(filename, 0600, nil)
		if err != nil {
			panic("Database was not found, and could not be created.")
		}

		d.db = tdb
	}

	for _, bucket := range bucketList {
		d.db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists(bucket)
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}
			return nil
		})
	}
}
