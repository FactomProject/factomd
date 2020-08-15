// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package mapdb

import (
	"sort"
	"sync"

	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/util"
)

type MapDB struct {
	Sem   sync.RWMutex
	Cache map[string]map[string][]byte // Our Cache
}

var _ interfaces.IDatabase = (*MapDB)(nil)

func (MapDB) Close() error {
	return nil
}

func (db *MapDB) ListAllBuckets() ([][]byte, error) {
	if db.Cache == nil {
		db.Sem.Lock()
		db.Cache = map[string]map[string][]byte{}
		db.Sem.Unlock()
	}

	db.Sem.RLock()
	defer db.Sem.RUnlock()

	answer := [][]byte{}
	for k, _ := range db.Cache {
		answer = append(answer, []byte(k))
	}

	return answer, nil
}

// Don't do anything here.
func (db *MapDB) Trim() {
}

func (db *MapDB) createCache(bucket []byte) {
	if db.Cache == nil {
		db.Sem.Lock()
		db.Cache = map[string]map[string][]byte{}
		db.Sem.Unlock()
	}
	db.Sem.RLock()
	_, ok := db.Cache[string(bucket)]
	db.Sem.RUnlock()
	if ok == false {
		db.Sem.Lock()
		db.Cache[string(bucket)] = map[string][]byte{}
		db.Sem.Unlock()
	}
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
	var hex []byte
	var err error

	//defer func() {
	//	messages.LogPrintf("database.txt", "Put(bucket %d[%x], key %d[%x], value %d[%x]", len(bucket), bucket, len(key), key, len(hex), hex)
	//}()

	if db.Cache == nil {
		db.Cache = map[string]map[string][]byte{}
	}
	_, ok := db.Cache[string(bucket)]
	if ok == false {
		db.Cache[string(bucket)] = map[string][]byte{}
	}
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
	var data []byte
	var ok bool
	//defer func() {
	//	messages.LogPrintf("database.txt", "Get(bucket %d[%x], key %d[%x], value %d[%x]", len(bucket), bucket, len(key), key, len(data), data)
	//}()

	db.createCache(bucket)

	db.Sem.RLock()
	defer db.Sem.RUnlock()

	if db.Cache == nil {
		db.Cache = map[string]map[string][]byte{}
	}
	_, ok = db.Cache[string(bucket)]
	if ok == false {
		db.Cache[string(bucket)] = map[string][]byte{}
	}
	data, ok = db.Cache[string(bucket)][string(key)]
	if ok == false {
		return nil, nil
	}
	if data == nil {
		return nil, nil
	}
	_, err := destination.UnmarshalBinaryData(data)
	if err != nil {
		_, err := destination.UnmarshalBinaryData(data)

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
	db.createCache(bucket)

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

func (db *MapDB) GetAll(bucket []byte, sample interfaces.BinaryMarshallableAndCopyable) ([]interfaces.BinaryMarshallableAndCopyable, [][]byte, error) {
	db.createCache(bucket)

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
		return nil, nil, err
	}

	answer := []interfaces.BinaryMarshallableAndCopyable{}
	for _, k := range keys {
		tmp := sample.New()
		v := db.Cache[string(bucket)][string(k)]
		err := tmp.UnmarshalBinary(v)
		if err != nil {
			return nil, nil, err
		}
		answer = append(answer, tmp)
	}
	return answer, keys, nil
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

func (db *MapDB) DoesKeyExist(bucket, key []byte) (bool, error) {
	db.createCache(bucket)

	db.Sem.RLock()
	defer db.Sem.RUnlock()

	if db.Cache == nil {
		db.Cache = map[string]map[string][]byte{}
	}
	_, ok := db.Cache[string(bucket)]
	if ok == false {
		db.Cache[string(bucket)] = map[string][]byte{}
	}
	data, ok := db.Cache[string(bucket)][string(key)]
	if ok == false {
		return false, nil
	}
	if data == nil {
		return false, nil
	}
	if len(data) < 1 {
		return false, nil
	}
	return true, nil
}
