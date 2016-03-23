// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package mapdb

import (
	"sort"
	"sync"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/util"
)

type MapDB struct {
	Sem   sync.RWMutex
	Cache map[string]map[string][]byte // Our Cache
}

var _ interfaces.IDatabase = (*MapDB)(nil)

func (MapDB) Close() error {
	return nil
}

func (db *MapDB) Init(bucketList [][]byte) {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	db.Cache = map[string]map[string][]byte{}
	for _, v := range bucketList {
		db.Cache[string(v)] = map[string][]byte{}
	}
}

func (db *MapDB) Put(bucket, key []byte, data interfaces.BinaryMarshallable) error {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	return db.rawPut(bucket, key, data)
}

func (db *MapDB) RawPut(bucket, key []byte, data interfaces.BinaryMarshallable) error {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	return db.rawPut(bucket, key, data)
}

func (db *MapDB) rawPut(bucket, key []byte, data interfaces.BinaryMarshallable) error {
	if db.Cache == nil {
		db.Cache = map[string]map[string][]byte{}
	}
	_, ok := db.Cache[string(bucket)]
	if ok == false {
		db.Cache[string(bucket)] = map[string][]byte{}
	}
	var hex []byte
	var err error
	if data != nil {
		hex, err = data.MarshalBinary()
		if err != nil {
			return err
		}
	}
	db.Cache[string(bucket)][string(key)] = hex
	return nil
}

func (db *MapDB) PutInBatch(records []interfaces.Record) error {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	for _, v := range records {
		err := db.rawPut(v.Bucket, v.Key, v.Data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *MapDB) Get(bucket, key []byte, destination interfaces.BinaryMarshallable) (interfaces.BinaryMarshallable, error) {
	db.Sem.RLock()
	defer db.Sem.RUnlock()

	if db.Cache == nil {
		db.Cache = map[string]map[string][]byte{}
	}
	_, ok := db.Cache[string(bucket)]
	if ok == false {
		db.Cache[string(bucket)] = map[string][]byte{}
	}
	v, ok := db.Cache[string(bucket)][string(key)]
	if ok == false {
		return nil, nil
	}
	if v == nil {
		return nil, nil
	}
	_, err := destination.UnmarshalBinaryData(v)
	if err != nil {
		return nil, err
	}
	return destination, nil
}

func (db *MapDB) Delete(bucket, key []byte) error {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	if db.Cache == nil {
		db.Cache = map[string]map[string][]byte{}
	}
	_, ok := db.Cache[string(bucket)]
	if ok == false {
		db.Cache[string(bucket)] = map[string][]byte{}
	}
	delete(db.Cache[string(bucket)], string(key))
	return nil
}

func (db *MapDB) ListAllKeys(bucket []byte) ([][]byte, error) {
	db.Sem.RLock()
	defer db.Sem.RUnlock()

	if db.Cache == nil {
		db.Cache = map[string]map[string][]byte{}
	}
	_, ok := db.Cache[string(bucket)]
	if ok == false {
		db.Cache[string(bucket)] = map[string][]byte{}
	}
	answer := [][]byte{}
	for k, _ := range db.Cache[string(bucket)] {
		answer = append(answer, []byte(k))
	}

	sort.Sort(util.ByByteArray(answer))

	return answer, nil
}

func (db *MapDB) GetAll(bucket []byte, sample interfaces.BinaryMarshallableAndCopyable) ([]interfaces.BinaryMarshallableAndCopyable, error) {
	db.Sem.RLock()
	defer db.Sem.RUnlock()

	if db.Cache == nil {
		db.Cache = map[string]map[string][]byte{}
	}
	_, ok := db.Cache[string(bucket)]
	if ok == false {
		db.Cache[string(bucket)] = map[string][]byte{}
	}

	keys, err := db.ListAllKeys(bucket)
	if err != nil {
		return nil, err
	}

	answer := []interfaces.BinaryMarshallableAndCopyable{}
	for _, k := range keys {
		tmp := sample.New()
		v := db.Cache[string(bucket)][string(k)]
		err := tmp.UnmarshalBinary(v)
		if err != nil {
			return nil, err
		}
		answer = append(answer, tmp)
	}
	return answer, nil
}

func (db *MapDB) Clear(bucket []byte) error {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	if db.Cache == nil {
		db.Cache = map[string]map[string][]byte{}
	}
	delete(db.Cache, string(bucket))
	return nil
}
