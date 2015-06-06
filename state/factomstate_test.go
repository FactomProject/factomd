// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
    "encoding/hex"
    "encoding/binary"
    "fmt"
    sc "github.com/FactomProject/simplecoin"
    "github.com/FactomProject/simplecoin/database"
    "github.com/agl/ed25519"
    "math/rand"
    "testing"
    
)

var _ = hex.EncodeToString
var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Write
var _ = sc.Prtln 

func GetDatabase() database.ISCDatabase {
    
    var bucketList [][]byte
    var instances  map[[sc.ADDRESS_LENGTH]byte]sc.IBlock
    var addinstance = func  (b sc.IBlock){
        key := new ([32]byte)
        copy(key[:],b.GetDBHash().Bytes())
        instances[*key] = b 
    }
    
    bucketList = make([][]byte,5,5)
    
    bucketList[0] = []byte("factoidAddress_balances")
    bucketList[0] = []byte("factoidOrphans_balances")
    bucketList[0] = []byte("factomAddress_balances")
    
    instances = make(map[[sc.ADDRESS_LENGTH]byte]sc.IBlock)
    
    addinstance (new(sc.Address))
    addinstance (new(sc.Hash))
    addinstance (new(sc.InAddress))
    addinstance (new(sc.OutAddress))
    addinstance (new(sc.OutECAddress))
    addinstance (new(sc.RCD_1))
    addinstance (new(sc.RCD_2))
    addinstance (new(sc.Signature))
    addinstance (new(sc.Transaction))
    addinstance (new(FSbalance))
    
    db := new(database.BoltDB)
    db.Clear(bucketList,"/tmp/fs_test.db")
    db.Init(bucketList,instances,"/tmp/fs_test.db")
    
    return db
}

func Test_updating_balances_factomstate (test *testing.T) {
    fs := new(FactomState)
    fs.database = GetDatabase()

}