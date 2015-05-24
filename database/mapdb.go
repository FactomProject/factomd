// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package database

import (
	"github.com/FactomProject/simplecoin"
)

type MapDB struct {
	SCDatabase

	backer ISCDatabase          // We can have backing databases.  For now this will be nil
    persist ISCDatabase         // We do need LevelDB or Bolt.  It would go here.
    
	cache  map[DBKey](simplecoin.IBlock) // Our Cache
}

var _ ISCDatabase = (*MapDB)(nil)

func (db *MapDB) Init(a ...interface{}) {
	db.cache = make(map[DBKey](simplecoin.IBlock), 100)
}

func (db *MapDB) GetRaw(bucket []byte, key []byte) (value simplecoin.IBlock) {
    dbkey := makeKey(bucket,key).(*DBKey)
    value = db.cache[*dbkey]
    if value == nil && db.backer != nil {
        value = db.backer.GetKey(dbkey)
        if value != nil {
            db.PutKey(dbkey, value)  // Put this value in our cache
        }
    }
    return value
}

func (db *MapDB) PutRaw(bucket []byte, key []byte, value simplecoin.IBlock) {
    dbkey := makeKey(bucket, key).(*DBKey)
    db.cache[*dbkey] = value
    if db.persist != nil {
        db.persist.PutRaw(bucket,key,value)
    }
}

func (db *MapDB) Get(bucket string, key simplecoin.IHash) (value simplecoin.IBlock) {
    return db.GetRaw([]byte(bucket), key.Bytes())
}

func (db *MapDB) GetKey(key IDBKey) (value simplecoin.IBlock) {
    return db.GetRaw(key.GetBucket(),key.GetKey())
}

func (db *MapDB) Put(bucket string, key simplecoin.IHash, value simplecoin.IBlock) {
    b := []byte(bucket)
    k := key.Bytes()
    db.PutRaw(b, k, value)
}

func (db *MapDB) PutKey(key IDBKey, value simplecoin.IBlock) {
    db.PutRaw(key.GetBucket(), key.GetKey(), value)
}
