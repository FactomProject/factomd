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
    sc "github.com/FactomProject/simplecoin"
    "github.com/FactomProject/simplecoin/block"
    "github.com/FactomProject/simplecoin/database"
    "github.com/FactomProject/simplecoin/state"
)

var _ = fmt.Printf

func GetDatabase() database.ISCDatabase {
    
    var bucketList [][]byte
    var instances  map[[sc.ADDRESS_LENGTH]byte]sc.IBlock
    
    bucketList = make([][]byte,0,5)
    
    bucketList = append(bucketList,[]byte(sc.DB_FACTOID_BLOCKS))
    bucketList = append(bucketList,[]byte(sc.DB_F_BALANCES))
    bucketList = append(bucketList,[]byte(sc.DB_EC_BALANCES))    
    
    instances = make(map[[sc.ADDRESS_LENGTH]byte]sc.IBlock)
    
    var addinstance = func  (b sc.IBlock){
        key := new ([32]byte)
        copy(key[:],b.GetDBHash().Bytes())
        instances[*key] = b 
    }
    addinstance (new(sc.Address))
    addinstance (new(sc.Hash))
    addinstance (new(sc.InAddress))
    addinstance (new(sc.OutAddress))
    addinstance (new(sc.OutECAddress))
    addinstance (new(sc.RCD_1))
    addinstance (new(sc.RCD_2))
    addinstance (new(sc.Signature))
    addinstance (new(sc.Transaction))
    addinstance (new(block.SCBlock))
    addinstance (new(state.FSbalance))
 
    db := new(database.BoltDB)
    db.Init(bucketList,instances,"/tmp/fs_test.db")
    sc.Prtln("Initialize Persistent Database")
    return db
}
