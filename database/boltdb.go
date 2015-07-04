// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package database

import (
    "fmt"
    "bytes"
    "encoding/binary"
    "encoding/hex"
    fct "github.com/FactomProject/factoid"
    
    "github.com/boltdb/bolt"
)

var _ = hex.EncodeToString
// This database stores and retrieves IBlock instances.  To do that, it
// needs a list of buckets that the using function wants, so it can make sure
// all those buckets exist.  (Avoids checking and building buckets in every 
// write).  
//
// It also needs a map of a hash to a IBlock instance.  To support this, 
// every block needs to be able to give the database a Hash for its type.
// This has to match the reverse, where looking up the hash gives the 
// database the type for the hash.  This way, the database can marshal
// and unmarshal IBlocks for storage in the database.  And since the IBlocks
// can provide the hash, we don't need two maps.  Just the Hash to the
// IBlock.
//
// Lastly it needs a filename with a full path.  If none is specified, it will
// use "/tmp/bolt_my.db".  Not the best idea to let this code default.
//
type BoltDB struct {
	FDatabase
    
    db          *bolt.DB                        // Pointer to the bolt db
    instances   map[[32]byte]fct.IBlock  // Maps a hash to an instance of an IBlock
    filename    string                          // location to write the db
}

var _ IFDatabase = (*BoltDB)(nil)

func (bdb BoltDB) GetKeysValues(bucket []byte) (keys [][]byte, values []fct.IBlock) {
    keys = make([][]byte,0,32)
    values = make([]fct.IBlock,0,32)
    bdb.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(bucket))
        if b == nil {
            fmt.Println("bucket '",bucket,"' not found")
        }else{
            b.ForEach(func(k, v []byte) error {
                keys = append(keys,k)
                instance := bdb.GetInstance(v)
                values = append(values,instance)
                return nil
            })
        }
        return nil
    })
    return
}

func (b BoltDB) String() string {
    txt,err := b.MarshalText()
    if err != nil {return "<error>" }
    return string(txt)
}

func (d *BoltDB) Clear(bucketList [][]byte) {
    
    tdb, err := bolt.Open(d.filename, 0600, nil)
    
    if err != nil {
        panic("Database "+d.filename+" was not found, and could not be created.")
    }
    defer tdb.Close()

    for _,bucket := range bucketList {
        tdb.Update(func(tx *bolt.Tx) error {
            err := tx.DeleteBucket(bucket)
            if err != nil {
                return fmt.Errorf("No bucket: %s", err)
            }
            return nil
        })
    }
}       
// We have to make accomadation for many Init functions.  But what we really
// want here is:
//
//      Init(bucketList [][]byte, instances map[[32]byte]IBlock, filename string)
//
func (d *BoltDB) Init(a ...interface{}) {
    d.doNotCache = make(map[string][]byte,5)
    d.doNotPersist = make(map[string][]byte,5)
    
    bucketList := a[0].([][]byte)
    instances  := a[1].(map[[32]byte]fct.IBlock)
    if(len(a)<3) {
        d.filename = "/tmp/bolt_my.db"
    }else{
        d.filename = a[2].(string)   
    }
    
    tdb, err := bolt.Open(d.filename, 0600, nil)
    if err != nil {
        panic("Database was not found, and could not be created.")
    }
    
    d.db = tdb
    
    
    for _,bucket := range bucketList {
        d.db.Update(func(tx *bolt.Tx) error {
            _, err := tx.CreateBucketIfNotExists(bucket)
            if err != nil {
                return fmt.Errorf("create bucket: %s", err)
            }
            return nil
        })
    }
    
    d.instances = instances
}

func (d *BoltDB) Close() {
    d.db.Close()
}

func (d *BoltDB) GetRaw(bucket []byte, key []byte) (value fct.IBlock) {
    var v []byte
    d.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket(bucket)
        v1 := b.Get(key)
        if v1 == nil { return nil }
        v = make([]byte,len(v1))
        copy(v,v1)
        return nil
    })
    if v == nil || len(v)<32 {      // If the value is undefined, return nil
        return nil
    }
    
    return d.GetInstance(v)
    
}

// Return the instance, properly unmarshaled, given the entry in the database, which is 
// the hash for the Instance (vv) followed by the source from which to unmarshal (v)
func (d *BoltDB) GetInstance(v []byte) fct.IBlock {

    var vv[32]byte
    copy(vv[:],v[:32])
    v=v[32:]
    
    var instance fct.IBlock = d.instances[vv]
    if instance == nil {
        vp := fct.NewHash(vv[:])
        fct.Prtln("Object hash: ",vp)
        panic("This should not happen.  Object stored in the database has no IBlock instance")
    }
    
    r := instance.GetNewInstance()
    if r == nil {
        panic("An IBlock has failed to implement GetNewInstance()")
    }
    
    datalen, v := binary.BigEndian.Uint32(v[0:4]), v[4:]
    if len(v) != int(datalen) {
        fct.Prtln("Lengths don't match.  Expected ",datalen," and got ",len(v))
        panic("Data not returned properly")
    }
    err := r.UnmarshalBinary(v)
    if err != nil {
        panic("This should not happen.  IBlock failed to unmarshal.")
    }
    
    return r
}


func (d *BoltDB) PutRaw(bucket []byte, key []byte, value fct.IBlock) {
    var out bytes.Buffer
    hash := value.GetDBHash()
    out.Write(hash.Bytes())
    data, err := value.MarshalBinary()
    binary.Write(&out, binary.BigEndian, uint32(len(data)))
    out.Write(data)
    
    if err != nil {
        panic("This should not happen.  Failed to marshal IBlock for BoltDB")
    }
    d.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket(bucket)
        err := b.Put(key, out.Bytes())
    return err
    })
    
}

// We don't care if delete works or not.  If the key isn't there, that's ok
func (d *BoltDB) DeleteKey(bucket []byte, key []byte) {
    d.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket(bucket)
        b.Delete(key)
        return nil
    })
    
}

func (db *BoltDB) Get(bucket string, key fct.IHash) (value fct.IBlock) {
    return db.GetRaw([]byte(bucket), key.Bytes())
}

func (db *BoltDB) GetKey(key IDBKey) (value fct.IBlock) {
    return db.GetRaw(key.GetBucket(),key.GetKey())
}

func (db *BoltDB) Put(bucket string, key fct.IHash, value fct.IBlock) {
    b := []byte(bucket)
    k := key.Bytes()
    db.PutRaw(b, k, value)
}

func (db *BoltDB) PutKey(key IDBKey, value fct.IBlock) {
    db.PutRaw(key.GetBucket(), key.GetKey(), value)
}