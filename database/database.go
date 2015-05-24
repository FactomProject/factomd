// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package database

import (
	"github.com/FactomProject/simplecoin"
)

/***********************************************
 * ISCDatabase
 *
 * Modeling this simple database structure after the architecture of most
 * key value stores.  The user gets to choose a "bucket" where they get a
 * key value from.  They also supply a key, which will generally be a hash,
 * but by no means must it be a hash.
 *
 * Right now, I am providing a Get, and a Put, each of which provide strings
 * for the bucket spec, as well as a Raw form (that is limited to 32 bytes).
 * Bucket specifications are limited to 32 bytes, even as strings.
 *
 * This isn't intended to be "real" but to provide a database like interface
 * that could become real, or could just be used for testing.
 *
 * Values are not limited here.  Factom limits most things to 10k
 ************************************************/
type ISCDatabase interface {
	simplecoin.IBlock

	// Users must call Init() prior to using the database.
	Init()
    
    // The Get methods return an entry, or nil if it does not yet
	// exist.  No errors are thrown.
	Get(bucket string, key simplecoin.IHash) simplecoin.IBlock
	GetRaw(bucket []byte, key []byte) simplecoin.IBlock
	GetKey(key IDBKey) simplecoin.IBlock

	// Put places the value in the database.  No errors are thrown.
	Put(bucket string, key simplecoin.IHash, value simplecoin.IBlock)
	PutRaw(bucket []byte, key []byte, value simplecoin.IBlock)
	PutKey(key IDBKey, value simplecoin.IBlock)

    // A Backer database allows the implementation of a least recently
    // used cache to purge data from memory.
    SetBacker(db ISCDatabase)            
    // A Persist database is needed to persist writes.  This is where 
    // one can hook up a LevelDB or Bolt database.
    SetPersist(db ISCDatabase)
}

type SCDatabase struct {
	ISCDatabase

	backer ISCDatabase          // We can have backing databases.  For now this will be nil
    persist ISCDatabase         // We do need LevelDB or Bolt.  It would go here.
    
	cache  map[DBKey](simplecoin.IBlock) // Our Cache
}

var _ ISCDatabase = (*SCDatabase)(nil)

type IDBKey interface {
}

type DBKey struct {
	IDBKey
	bucket [simplecoin.ADDRESS_LENGTH]byte
	key    [simplecoin.ADDRESS_LENGTH]byte
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

func (db *SCDatabase) Init() {
	db.cache = make(map[DBKey](simplecoin.IBlock), 100)
}

func (db *SCDatabase) GetRaw(bucket []byte, key []byte) (value simplecoin.IBlock) {
	return db.GetKey(makeKey(bucket, key))
}

func (db *SCDatabase) Get(bucket string, key simplecoin.IHash) (value simplecoin.IBlock) {
	return db.GetRaw([]byte(bucket), key.Bytes())
}

// Get the value out of hour hash.  If we don't have it, look and see if we
// have a backer ISCDatabase that does.
//
// Otherwise return a nil.
func (db *SCDatabase) GetKey(key IDBKey) (value simplecoin.IBlock) {
	value = db.cache[*key.(*DBKey)]
	if value == nil && db.backer != nil {
		return db.backer.GetKey(key)
	}
	return value
}

func (db *SCDatabase) Put(bucket string, key simplecoin.IHash, value simplecoin.IBlock) {
	db.PutKey(makeKey([]byte(bucket), key.Bytes()), value)
}

func (db *SCDatabase) PutRaw(bucket []byte, key []byte, value simplecoin.IBlock) {
	db.PutKey(makeKey(bucket, key), value)
}

func (db *SCDatabase) PutKey(key IDBKey, value simplecoin.IBlock) {
	db.cache[*key.(*DBKey)] = value
}
