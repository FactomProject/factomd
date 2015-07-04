// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package database

import (
	"github.com/FactomProject/factoid"
)

/***********************************************
 * IFDatabase
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
type IFDatabase interface {
	factoid.IBlock

	// Clear removes all the specified buckets from the database.
	// This allows us to cleanly rebuild databases, or in the case
	// of testing, ensure a particular database state.
	Clear(bucketList [][]byte)
	
	// Users must call Init() prior to using the database.
	Init(a ...interface{})
    
    // Users should defer a call to Close()
    Close()
    
    // The Get methods return an entry, or nil if it does not yet
	// exist.  No errors are thrown.
	Get(bucket string, key factoid.IHash) factoid.IBlock
	GetRaw(bucket []byte, key []byte) factoid.IBlock
	GetKey(key IDBKey) factoid.IBlock

	// Put places the value in the database.  No errors are thrown.
	Put(bucket string, key factoid.IHash, value factoid.IBlock)
	PutRaw(bucket []byte, key []byte, value factoid.IBlock)
	PutKey(key IDBKey, value factoid.IBlock)
    DeleteKey(bucket []byte, key[]byte)
    
    // A Backer database allows the implementation of a least recently
    // used cache to purge data from memory.
    SetBacker(db IFDatabase)     
    GetBacker() IFDatabase
    // A Persist database is needed to persist writes.  This is where 
    // one can hook up a LevelDB or Bolt database.
    SetPersist(db IFDatabase)
    GetPersist() IFDatabase
    // This is a bucket that holds small or limited information that
    // does not have to go to disk.
    DoNotPersist(bucket string )
    // This bucket is written to disk.  We do not cache into memory 
    // because the stuff in it is too big, and we don't need fast 
    // access to it.
    DoNotCache(bucket string )
    // Get a list of the keys and values in a bucket
    GetKeysValues(bucket []byte) (keys [][]byte, values []factoid.IBlock)
}


