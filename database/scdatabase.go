// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package database

import (
	"github.com/FactomProject/simplecoin"
)

type SCDatabase struct {
	ISCDatabase
	backer ISCDatabase          // We can have backing databases.  For now this will be nil
	persist ISCDatabase         // We do need LevelDB or Bolt.  It would go here.
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

// A Backer database allows the implementation of a least recently
// used cache to purge data from memory.
func (db *SCDatabase) SetBacker(b ISCDatabase) {
    db.backer = b
}
func (db SCDatabase) GetBacker() ISCDatabase{
    return db.backer
}
// A Persist database is needed to persist writes.  This is where 
// one can hook up a LevelDB or Bolt database.
func (db *SCDatabase) SetPersist(p ISCDatabase){
    db.persist = p
}
func (db SCDatabase) GetPersist() ISCDatabase{
    return db.persist
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


