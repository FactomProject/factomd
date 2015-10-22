package hybridDB

import (
	"github.com/FactomProject/factomd/common/interfaces"

	"github.com/FactomProject/factomd/database/boltdb"
	"github.com/FactomProject/factomd/database/leveldb"
	"github.com/FactomProject/factomd/database/mapdb"
)

type HybridDB struct {
	temporaryStorage  interfaces.IDatabase
	persistentStorage interfaces.IDatabase
}

var _ interfaces.IDatabase = (*HybridDB)(nil)

func (db *HybridDB) Close() error {
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
	return db.persistentStorage.ListAllKeys(bucket)
}

func (db *HybridDB) GetAll(bucket []byte, sample interfaces.BinaryMarshallableAndCopyable) ([]interfaces.BinaryMarshallableAndCopyable, error) {
	return db.persistentStorage.GetAll(bucket, sample)
}

func (db *HybridDB) Clear(bucket []byte) error {
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
