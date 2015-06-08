// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package database

import (
	sc "github.com/FactomProject/simplecoin"
)

type MapDB struct {
	SCDatabase
    doNotPersist  map[string] []byte  
	cache  map[DBKey](sc.IBlock) // Our Cache
}

var _ ISCDatabase = (*MapDB)(nil)

func (m MapDB) DoNotPersist(bucket string) {
    m.doNotPersist[bucket]= []byte(bucket)
}

func (b MapDB) String() string {
    txt,err := b.MarshalText()
    if err != nil {return "<error>" }
    return string(txt)
}

func (db *MapDB) Init(a ...interface{}) {
	db.cache = make(map[DBKey](sc.IBlock), 100)
    db.doNotPersist = make(map[string][]byte,5)
}

func (db *MapDB) GetRaw(bucket []byte, key []byte) (value sc.IBlock) {
    dbkey := makeKey(bucket,key).(*DBKey)
    value = db.cache[*dbkey]
    if value == nil && db.GetBacker() != nil {
        value = db.GetBacker().GetRaw(bucket,key)
        if value != nil {
            db.cache[*dbkey]=value  // Put this value in our cache
        }
    }
    return value
}

func (db *MapDB) PutRaw(bucket []byte, key []byte, value sc.IBlock) {
    dbkey := makeKey(bucket, key).(*DBKey)
    db.cache[*dbkey] = value
    if db.doNotPersist[string(bucket)] != nil { return }
    if db.GetPersist() != nil {
        db.GetPersist().PutRaw(bucket,key,value)
    }
}

func (db *MapDB) Get(bucket string, key sc.IHash) (value sc.IBlock) {
    return db.GetRaw([]byte(bucket), key.Bytes())
}

func (db *MapDB) GetKey(key IDBKey) (value sc.IBlock) {
    return db.GetRaw(key.GetBucket(),key.GetKey())
}

func (db *MapDB) Put(bucket string, key sc.IHash, value sc.IBlock) {
    b := []byte(bucket)
    k := key.Bytes()
    db.PutRaw(b, k, value)
}

func (db *MapDB) PutKey(key IDBKey, value sc.IBlock) {
    db.PutRaw(key.GetBucket(), key.GetKey(), value)
}
