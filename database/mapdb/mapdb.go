// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package database

import (
	"bytes"
	"fmt"
	. "github.com/FactomProject/factomd/common/interfaces"
)

type MapDB struct {
	cache map[[]byte]map[[]byte][]byte // Our Cache
}

var _ IDatabase = (*MapDB)(nil)

func (MapDB) Close() error {
	return nil
}

func (db *MapDB) Init(bucketList [][]byte) {
	db.cache = map[[]byte]map[[]byte][]byte{}
	for _, v := range bucketList {
		db.cache[v] = map[[]byte][]byte{}
	}
}

func (db *MapDB) Put(bucket, key []byte, data BinaryMarshallable) error {
	_, ok := db.cache[bucket]
	if ok == false {
		db.cache[bucket] = map[[]byte][]byte{}
	}
	hex, err := value.MarshalBinary()
	if err != nil {
		return err
	}
	db.cache[bucket][key] = hex
	return nil
}

func (db *MapDB) Get(bucket, key []byte, destination BinaryMarshallable) (BinaryMarshallable, error) {
	_, ok := db.cache[bucket]
	if ok == false {
		db.cache[bucket] = map[[]byte][]byte{}
	}
	v, ok := db.cache[bucket][key]
	if ok == false {
		return nil, nil
	}
	_, err := destination.UnmarshalBinaryData(v)
	if err != nil {
		return nil, err
	}
	return destination, nil
}

func (db *MapDB) Delete(bucket, key []byte) error {
	_, ok := db.cache[bucket]
	if ok == false {
		db.cache[bucket] = map[[]byte][]byte{}
	}
	delete(db.cache[bucket], key)
	return nil
}

func (db *MapDB) ListAllKeys(bucket []byte) ([][]byte, error) {
	_, ok := db.cache[bucket]
	if ok == false {
		db.cache[bucket] = map[[]byte][]byte{}
	}
	answer := [][]byte{}
	for k, _ := range db.cache[bucket] {
		answer = append(answer, k)
	}
	return asnwer, nil
}

func (db *MapDB) Clear(bucket []byte) error {
	delete(db.cache, bucket)
	return nil
}
