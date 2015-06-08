// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
    "encoding/hex"
    "encoding/binary"
    "fmt"
    ftc "github.com/FactomProject/factoid"
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
var _ = ftc.Prtln 

func GetDatabase() database.IFDatabase {
    
    var bucketList [][]byte
    var instances  map[[ftc.ADDRESS_LENGTH]byte]ftc.IBlock
    var addinstance = func  (b ftc.IBlock){
        key := new ([32]byte)
        copy(key[:],b.GetDBHash().Bytes())
        instances[*key] = b 
    }
    
    bucketList = make([][]byte,5,5)
    
    bucketList[0] = []byte("factoidAddress_balances")
    bucketList[0] = []byte("factoidOrphans_balances")
    bucketList[0] = []byte("factomAddress_balances")
    
    instances = make(map[[ftc.ADDRESS_LENGTH]byte]ftc.IBlock)
    
    addinstance (new(ftc.Address))
    addinstance (new(ftc.Hash))
    addinstance (new(ftc.InAddress))
    addinstance (new(ftc.OutAddress))
    addinstance (new(ftc.OutECAddress))
    addinstance (new(ftc.RCD_1))
    addinstance (new(ftc.RCD_2))
    addinstance (new(ftc.Signature))
    addinstance (new(ftc.Transaction))
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