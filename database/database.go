// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package database

import ()

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
    Get(bucket string,key IHash) []byte
    GetRaw(bucket []byte, key []byte) []byte 
    GetKey(key DBKey)
    Put(bucket string,key IHash, value []byte)
    PutRaw(bucket []byte, key []byte, value []byte)
    PutKey(key DBKey, value []byte)
}

type DBKey struct {
    bucket [ADDRESS_LENGTH] byte
    key    [ADDRESS_LENGTH] byte
}

type SCDatabase struct {
    ISCDatabase                    
    
    backer ISCDatabase             // We can have backing databases.  For now this will be nil
	cache map[key]([]byte)         // Our Cache
    
}

func (db *FDatabase) GetRaw(bucket []byte, key []byte) (value []byte) {
    
    if(len(bucket) > ADDRESS_LENGTH || len(key) > ADDRESS_LENGTH ){
        panic("Key provided to ISCDatabase is too long")
    }
    
    k := new(DBKey)
    copy(k.bucket[:],bucket)
    copy(k.key[:],key)
    
    return db.getBytes(k)
}

func (db *FDatabase) Get(bucket string, key IHash) (value []byte) {
    return db.GetRaw([]byte(string),key.Bytes())
}

// Get the value out of hour hash.  If we don't have it, look and see if we 
// have a backer ISCDatabase that does.  
//
// Otherwise return a nil.
func (db *FDatabase) getBytes(key DBKey) (value []byte) {
    value = cache[key]
    if value == nil && backer != nil {
        return backer.Get(key)
    }
	return value
}
