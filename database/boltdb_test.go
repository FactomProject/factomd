// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package database

import (
    "encoding/binary"
    "fmt"
    sc "github.com/FactomProject/simplecoin"
    "github.com/agl/ed25519"
    "math/rand"
    "testing"
)

var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Read

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

func cp (a sc.IHash) [sc.ADDRESS_LENGTH]byte {
    r := new([sc.ADDRESS_LENGTH]byte)
    copy(r[:],a.Bytes())
    return *r
}

func Test_bolt_init(t *testing.T) {
    db := new(BoltDB)

    bucketList := make([][]byte,5,5)
    
    bucketList[0] = []byte("one")
    bucketList[1] = []byte("two")
    bucketList[2] = []byte("three")
    bucketList[3] = []byte("four")
    bucketList[4] = []byte("five")
    
    instances := make(map[[sc.ADDRESS_LENGTH]byte]sc.IBlock)
    {   
        var a sc.IBlock
        a = new(sc.Address); instances[cp(a.GetDBHash())] = a 
        a = new(sc.Hash); instances[cp(a.GetDBHash())] = a 
        a = new(sc.InAddress); instances[cp(a.GetDBHash())] = a 
        a = new(sc.OutAddress); instances[cp(a.GetDBHash())] = a 
        a = new(sc.OutECAddress); instances[cp(a.GetDBHash())] = a 
        a = new(sc.RCD_1); instances[cp(a.GetDBHash())] = a 
        a = new(sc.RCD_2); instances[cp(a.GetDBHash())] = a 
        a = new(sc.Signature); instances[cp(a.GetDBHash())] = a 
        a = new(sc.Transaction); instances[cp(a.GetDBHash())] = a 
    }
    db.Init(bucketList, instances)
    a := new(sc.Address)
    a.SetBytes(sc.Sha([]byte("I came, I saw")).Bytes())
    db.Put("one",sc.Sha([]byte("one")),a)
    r := db.Get("one",sc.Sha([]byte("one")))
    
    if !a.IsEqual(r) { t.Fail() }
}
    







