// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// This package exists to isolate the fact that we need to reference
// almost all the objects within the Transaction package inorder to
// properly configure the database.   This code can be called by 
// system level code (which is isolated from the other Transaction
// packages) without causing import loops.
package stateinit

import (
    "fmt"
    ftc "github.com/FactomProject/factoid"
    "github.com/FactomProject/factoid/block"
    "github.com/FactomProject/factoid/database"
    "github.com/FactomProject/factoid/state"
)

var _ = fmt.Printf

func GetDatabase() database.IFDatabase {
    
    var bucketList [][]byte
    var instances  map[[ftc.ADDRESS_LENGTH]byte]ftc.IBlock
    
    bucketList = make([][]byte,0,5)
    
    bucketList = append(bucketList,[]byte(ftc.DB_FACTOID_BLOCKS))
    bucketList = append(bucketList,[]byte(ftc.DB_F_BALANCES))
    bucketList = append(bucketList,[]byte(ftc.DB_EC_BALANCES))    
    
    instances = make(map[[ftc.ADDRESS_LENGTH]byte]ftc.IBlock)
    
    var addinstance = func  (b ftc.IBlock){
        key := new ([32]byte)
        copy(key[:],b.GetDBHash().Bytes())
        instances[*key] = b 
    }
    addinstance (new(ftc.Address))
    addinstance (new(ftc.Hash))
    addinstance (new(ftc.InAddress))
    addinstance (new(ftc.OutAddress))
    addinstance (new(ftc.OutECAddress))
    addinstance (new(ftc.RCD_1))
    addinstance (new(ftc.RCD_2))
    addinstance (new(ftc.Signature))
    addinstance (new(ftc.Transaction))
    addinstance (new(block.FBlock))
    addinstance (new(state.FSbalance))
 
    db := new(database.BoltDB)
    db.Init(bucketList,instances,"/tmp/fs_test.db")
    ftc.Prtln("Initialize Persistent Database")
    return db
}
