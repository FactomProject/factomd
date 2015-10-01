// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package mapdb

import (
	. "github.com/FactomProject/factomd/common/interfaces"
)

type MapDB struct {
	cache map[string]map[string][]byte // Our Cache
}

var _ IDatabase = (*MapDB)(nil)

func (MapDB) Close() error {
	return nil
}

func (db *MapDB) Init(bucketList [][]byte) {
	db.cache = map[string]map[string][]byte{}
	for _, v := range bucketList {
		db.cache[string(v)] = map[string][]byte{}
	}
}

func (db *MapDB) Put(bucket, key []byte, data BinaryMarshallable) error {
	_, ok := db.cache[string(bucket)]
	if ok == false {
		db.cache[string(bucket)] = map[string][]byte{}
	}
	hex, err := data.MarshalBinary()
	if err != nil {
		return err
	}
	db.cache[string(bucket)][string(key)] = hex
	return nil
}

func (db *MapDB) PutInBatch(records []Record) error {
	for _, v := range records {
		err := db.Put(v.Bucket, v.Key, v.Data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *MapDB) Get(bucket, key []byte, destination BinaryMarshallable) (BinaryMarshallable, error) {
	_, ok := db.cache[string(bucket)]
	if ok == false {
		db.cache[string(bucket)] = map[string][]byte{}
	}
	v, ok := db.cache[string(bucket)][string(key)]
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
	_, ok := db.cache[string(bucket)]
	if ok == false {
		db.cache[string(bucket)] = map[string][]byte{}
	}
	delete(db.cache[string(bucket)], string(key))
	return nil
}

func (db *MapDB) ListAllKeys(bucket []byte) ([][]byte, error) {
	_, ok := db.cache[string(bucket)]
	if ok == false {
		db.cache[string(bucket)] = map[string][]byte{}
	}
	answer := [][]byte{}
	for k, _ := range db.cache[string(bucket)] {
		answer = append(answer, []byte(k))
	}
	return answer, nil
}

func (db *MapDB) Clear(bucket []byte) error {
	delete(db.cache, string(bucket))
	return nil
}
