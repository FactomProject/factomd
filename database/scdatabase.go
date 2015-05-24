// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package database

import (
	"github.com/FactomProject/simplecoin"
)

type SCDatabase struct {
	ISCDatabase
}

var _ ISCDatabase = (*SCDatabase)(nil)

type IDBKey interface {
    GetBucket() []byte
    GetKey()    []byte
}

type DBKey struct {
	IDBKey
	bucket [simplecoin.ADDRESS_LENGTH]byte
	key    [simplecoin.ADDRESS_LENGTH]byte
}

func (k DBKey) GetBucket() []byte{
    return k.bucket[:]
}

func (k DBKey) GetKey()[]byte {
    return k.key[:]
}

func makeKey(bucket []byte, key []byte) IDBKey {

	if len(bucket) > simplecoin.ADDRESS_LENGTH || len(key) > simplecoin.ADDRESS_LENGTH {
		panic("Key provided to ISCDatabase is too long")
	}

	k := new(DBKey)
	copy(k.bucket[:], bucket)
	copy(k.key[:], key)

	return k
}

func (db *SCDatabase) Get(bucket string, key simplecoin.IHash) (value simplecoin.IBlock) {
	return db.GetRaw([]byte(bucket), key.Bytes())
}

func (db *SCDatabase) GetKey(key IDBKey) (value simplecoin.IBlock) {
	return db.GetRaw(key.GetBucket(),key.GetKey())
}

func (db *SCDatabase) Put(bucket string, key simplecoin.IHash, value simplecoin.IBlock) {
    b := []byte(bucket)
    k := key.Bytes()
    db.PutRaw(b, k, value)
}

func (db *SCDatabase) PutKey(key IDBKey, value simplecoin.IBlock) {
	db.PutRaw(key.GetBucket(), key.GetKey(), value)
}
