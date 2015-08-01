// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package database

import (
    "fmt"
    "bytes"
	fct "github.com/FactomProject/factoid"
)

var _ = bytes.Compare

type MapDB struct {
	FDatabase
	cache  map[DBKey](fct.IBlock) // Our Cache
}

var _ IFDatabase = (*MapDB)(nil)
var _ = fmt.Println


func (MapDB) Close() {}

func (MapDB) IsEqual(fct.IBlock) []fct.IBlock {
    return nil
}

func (MapDB) GetNewInstance() fct.IBlock {
    return new(MapDB)
}

func (MapDB) MarshalBinary() ([]byte, error) {
    return nil,nil
}

func (MapDB) UnmarshalBinary([]byte) error {
    return nil
}

func (MapDB) UnmarshalBinaryData([]byte) ([]byte, error) {
    return nil,nil
}


func (b MapDB) GetKeysValues(bucket []byte) (keys [][]byte, values []fct.IBlock) {

    if b.GetPersist() == nil || b.doNotPersist[string(bucket)] != nil {    
        keys = make([][]byte,0,32)
        values = make([]fct.IBlock,0,32)

        for dbKey,v:= range b.cache {
            b := dbKey.GetBucket();
            for b[len(b)-1] == 0 { b = b[:len(b)-1] }
            if bytes.Compare(b,bucket)==0 {
                keys = append(keys, dbKey.GetKey())
                values = append(values,v)
            }
        }
        return
    }
    return b.GetPersist().GetKeysValues(bucket)
}

func (b MapDB) String() string {
    txt,err := b.MarshalText()
    if err != nil {return "<error>" }
    return string(txt)
}

func (db *MapDB) Init(a ...interface{}) {
    db.cache = make(map[DBKey](fct.IBlock), 100)
    db.doNotCache = make(map[string][]byte,5)
    db.doNotPersist = make(map[string][]byte,5)
}

func (db *MapDB) GetRaw(bucket []byte, key []byte) (value fct.IBlock) {
    dbkey := makeKey(bucket,key).(*DBKey)
    value = db.cache[*dbkey]
    if value == nil && db.GetBacker() != nil {
        value = db.GetBacker().GetRaw(bucket,key)
        if value != nil && db.doNotCache[string(bucket)] == nil { 
            db.cache[*dbkey]=value  // Put this value in our cache
        }
    }
    if value == nil && db.GetPersist() != nil {
        value = db.GetPersist().GetRaw(bucket,key)
        if value != nil && db.doNotCache[string(bucket)] == nil { 
            db.cache[*dbkey]=value  // Put this value in our cache
        }
    }
    return value
}

func (db *MapDB) PutRaw(bucket []byte, key []byte, value fct.IBlock) {
    dbkey := makeKey(bucket, key).(*DBKey)
    db.cache[*dbkey] = value
    if db.doNotPersist[string(bucket)] != nil { return }
    if db.GetPersist() != nil {
        db.GetPersist().PutRaw(bucket,key,value)
    }
}

func (db *MapDB) DeleteKey(bucket []byte, key []byte) {
    dbkey := makeKey(bucket,key).(*DBKey)
    db.cache[*dbkey] = nil
    if db.doNotPersist[string(bucket)] != nil && db.GetPersist() != nil {
        db.GetPersist().DeleteKey(bucket,key)
    }
    if db.GetBacker() != nil {
        db.GetBacker().DeleteKey(bucket,key)
    }
}

func (db *MapDB) Get(bucket string, key fct.IHash) (value fct.IBlock) {
    return db.GetRaw([]byte(bucket), key.Bytes())
}

func (db *MapDB) GetKey(key IDBKey) (value fct.IBlock) {
    return db.GetRaw(key.GetBucket(),key.GetKey())
}

func (db *MapDB) Put(bucket string, key fct.IHash, value fct.IBlock) {
    b := []byte(bucket)
    k := key.Bytes()
    db.PutRaw(b, k, value)
}

func (db *MapDB) PutKey(key IDBKey, value fct.IBlock) {
    db.PutRaw(key.GetBucket(), key.GetKey(), value)
}
