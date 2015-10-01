package hybridDB

import (
	. "github.com/FactomProject/factomd/common/interfaces"

	"github.com/FactomProject/factomd/database/boltdb"
	"github.com/FactomProject/factomd/database/mapdb"
)

type HybridDB struct {
	temporaryStorage  IDatabase
	persistentStorage IDatabase
}

var _ IDatabase = (*HybridDB)(nil)

func (db *HybridDB) Close() error {
	err := db.temporaryStorage.Close()
	if err != nil {
		return err
	}
	err = db.persistentStorage.Close()
	return err
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

func (db *HybridDB) Put(bucket, key []byte, data BinaryMarshallable) error {
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

func (db *HybridDB) Get(bucket, key []byte, destination BinaryMarshallable) (BinaryMarshallable, error) {
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
