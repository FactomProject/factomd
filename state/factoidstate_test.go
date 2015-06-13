// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
    "encoding/hex"
    "encoding/binary"
    "fmt"
    fct "github.com/FactomProject/factoid"
    "github.com/FactomProject/factoid/database"
    "github.com/agl/ed25519"
    "math/rand"
    "testing"
    
)

var _ = hex.EncodeToString
var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Write
var _ = fct.Prtln 

func GetDatabase() database.IFDatabase {
    
    var bucketList [][]byte
    var instances  map[[fct.ADDRESS_LENGTH]byte]fct.IBlock
    var addinstance = func  (b fct.IBlock){
        key := new ([32]byte)
        copy(key[:],b.GetDBHash().Bytes())
        instances[*key] = b 
    }
    
    bucketList = make([][]byte,5,5)
    
    bucketList[0] = []byte("factoidAddress_balances")
    bucketList[0] = []byte("factoidOrphans_balances")
    bucketList[0] = []byte("factomAddress_balances")
    
    instances = make(map[[fct.ADDRESS_LENGTH]byte]fct.IBlock)
    
    addinstance (new(fct.Address))
    addinstance (new(fct.Hash))
    addinstance (new(fct.InAddress))
    addinstance (new(fct.OutAddress))
    addinstance (new(fct.OutECAddress))
    addinstance (new(fct.RCD_1))
    addinstance (new(fct.RCD_2))
    addinstance (new(fct.Signature))
    addinstance (new(fct.Transaction))
    addinstance (new(FSbalance))
    
    db := new(database.BoltDB)

    db.Init(bucketList,instances,"/tmp/fs_test.db")
    
    return db
}

func Test_updating_balances_FactoidState (test *testing.T) {
    fs := new(FactoidState)
    fs.database = GetDatabase()

}