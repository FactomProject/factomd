package hybridDB

import (
	"sync"

	"github.com/PaulSnow/factom2d/common/interfaces"

	"github.com/PaulSnow/factom2d/database/boltdb"
	"github.com/PaulSnow/factom2d/database/leveldb"
	"github.com/PaulSnow/factom2d/database/mapdb"
)

type HybridDB struct {
	Sem               sync.RWMutex
	temporaryStorage  interfaces.IDatabase
	persistentStorage interfaces.IDatabase
}

var _ interfaces.IDatabase = (*HybridDB)(nil)

func (db *HybridDB) ListAllBuckets() ([][]byte, error) {
	db.Sem.RLock()
	defer db.Sem.RUnlock()
	return db.persistentStorage.ListAllBuckets()
}

func (db *HybridDB) Trim() {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	m := new(mapdb.MapDB)
	m.Init(nil)
	db.temporaryStorage = m
}

func (db *HybridDB) Close() error {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	err := db.temporaryStorage.Close()
	if err != nil {
		return err
	}
	err = db.persistentStorage.Close()
	return err
}

func NewLevelMapHybridDB(filename string, create bool) (*HybridDB, error) {
	answer := new(HybridDB)

	m := new(mapdb.MapDB)
	m.Init(nil)
	answer.temporaryStorage = m

	b, err := leveldb.NewLevelDB(filename, create)
	if err != nil {
		return nil, err
	}
	answer.persistentStorage = b

	return answer, nil
}

func NewBoltMapHybridDB(bucketList [][]byte, filename string) *HybridDB {
	answer := new(HybridDB)

	m := new(mapdb.MapDB)
	m.Init(bucketList)
	answer.temporaryStorage = m

	b := new(boltdb.BoltDB)
	b.Init(bucketList, filename)
	answer.persistentStorage = b

	return answer
}

func (db *HybridDB) Put(bucket, key []byte, data interfaces.BinaryMarshallable) error {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	err := db.persistentStorage.Put(bucket, key, data)
	if err != nil {
		return err
	}

	err = db.temporaryStorage.Put(bucket, key, data)
	if err != nil {
		return err
	}

	return nil
}

func (db *HybridDB) PutInBatch(records []interfaces.Record) error {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	err := db.persistentStorage.PutInBatch(records)
	if err != nil {
		return err
	}
	err = db.temporaryStorage.PutInBatch(records)
	if err != nil {
		return err
	}
	return nil
}

func (db *HybridDB) Get(bucket, key []byte, destination interfaces.BinaryMarshallable) (interfaces.BinaryMarshallable, error) {
	db.Sem.RLock()
	defer db.Sem.RUnlock()

	answer, err := db.temporaryStorage.Get(bucket, key, destination)
	if err != nil {
		return nil, err
	}
	if answer != nil {
		return answer, nil
	}

	answer, err = db.persistentStorage.Get(bucket, key, destination)
	if err != nil {
		return nil, err
	}
	db.temporaryStorage.Put(bucket, key, answer) //storing the data for later re-fetching

	return answer, nil
}

func (db *HybridDB) Delete(bucket, key []byte) error {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	err := db.persistentStorage.Delete(bucket, key)
	if err != nil {
		return err
	}

	err = db.temporaryStorage.Delete(bucket, key)
	if err != nil {
		return err
	}
	return nil
}

func (db *HybridDB) ListAllKeys(bucket []byte) ([][]byte, error) {
	db.Sem.RLock()
	defer db.Sem.RUnlock()

	return db.persistentStorage.ListAllKeys(bucket)
}

func (db *HybridDB) GetAll(bucket []byte, sample interfaces.BinaryMarshallableAndCopyable) ([]interfaces.BinaryMarshallableAndCopyable, [][]byte, error) {
	db.Sem.RLock()
	defer db.Sem.RUnlock()

	return db.persistentStorage.GetAll(bucket, sample)
}

func (db *HybridDB) Clear(bucket []byte) error {
	db.Sem.Lock()
	defer db.Sem.Unlock()

	err := db.persistentStorage.Clear(bucket)
	if err != nil {
		return err
	}

	err = db.temporaryStorage.Clear(bucket)
	if err != nil {
		return err
	}
	return nil
}

func (db *HybridDB) DoesKeyExist(bucket, key []byte) (bool, error) {
	db.Sem.RLock()
	defer db.Sem.RUnlock()
	exist, err := db.temporaryStorage.DoesKeyExist(bucket, key)
	if err != nil || exist == false {
		return db.persistentStorage.DoesKeyExist(bucket, key)
	}
	return exist, nil
}
