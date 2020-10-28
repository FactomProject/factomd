package databaseOverlay

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func (db *Overlay) SaveKeyValueStore(kvs interfaces.BinaryMarshallable, key []byte) error {
	if kvs == nil || key == nil {
		return nil
	}
	batch := []interfaces.Record{}

	batch = append(batch, interfaces.Record{Bucket: KEY_VALUE_STORE, Key: key, Data: kvs})

	err := db.DB.PutInBatch(batch)
	if err != nil {
		return err
	}

	return nil
}

func (db *Overlay) FetchKeyValueStore(key []byte, dst interfaces.BinaryMarshallable) (interfaces.BinaryMarshallable, error) {
	block, err := db.DB.Get(KEY_VALUE_STORE, key, dst)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(interfaces.BinaryMarshallable), nil
}

var DatabaseEntryHeightKey = []byte("DatabaseEntryHeight")

func (db *Overlay) SaveDatabaseEntryHeight(height uint32) error {
	buf := primitives.NewBuffer(nil)
	buf.PushUInt32(height)
	bs := new(primitives.ByteSlice)
	bs.Bytes = buf.DeepCopyBytes()

	return db.SaveKeyValueStore(bs, DatabaseEntryHeightKey)
}

func (db *Overlay) FetchDatabaseEntryHeight() (uint32, error) {
	bs := new(primitives.ByteSlice)
	_, err := db.FetchKeyValueStore(DatabaseEntryHeightKey, bs)
	if err != nil {
		return 0, err
	}
	buf := primitives.NewBuffer(bs.Bytes)
	height, err := buf.PopUInt32()
	if err != nil {
		return 0, err
	}
	return height, nil
}
