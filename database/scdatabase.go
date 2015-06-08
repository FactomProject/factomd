// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package database

import (
	"github.com/FactomProject/factoid"
)

type FDatabase struct {
	IFDatabase
	backer IFDatabase          // We can have backing databases.  For now this will be nil
	persist IFDatabase         // We do need LevelDB or Bolt.  It would go here.
}

var _ IFDatabase = (*FDatabase)(nil)

type IDBKey interface {
    GetBucket() []byte
    GetKey()    []byte
}

type DBKey struct {
	IDBKey
	bucket [factoid.ADDRESS_LENGTH]byte
	key    [factoid.ADDRESS_LENGTH]byte
}

// A Backer database allows the implementation of a least recently
// used cache to purge data from memory.
func (db *FDatabase) SetBacker(b IFDatabase) {
    db.backer = b
}
func (db FDatabase) GetBacker() IFDatabase{
    return db.backer
}
// A Persist database is needed to persist writes.  This is where 
// one can hook up a LevelDB or Bolt database.
func (db *FDatabase) SetPersist(p IFDatabase){
    db.persist = p
}
func (db FDatabase) GetPersist() IFDatabase{
    return db.persist
} 

func (k DBKey) GetBucket() []byte{
    return k.bucket[:]
}

func (k DBKey) GetKey()[]byte {
    return k.key[:]
}

func makeKey(bucket []byte, key []byte) IDBKey {

	if len(bucket) > factoid.ADDRESS_LENGTH || len(key) > factoid.ADDRESS_LENGTH {
		panic("Key provided to IFDatabase is too long")
	}

	k := new(DBKey)
	copy(k.bucket[:], bucket)
	copy(k.key[:], key)

	return k
}


