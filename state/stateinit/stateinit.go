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
    fct "github.com/FactomProject/factoid"
    "github.com/FactomProject/factoid/block"
    "github.com/FactomProject/factoid/database"
    "github.com/FactomProject/factoid/state"
    "github.com/FactomProject/factoid/wallet"
)

var _ = fmt.Printf

func NewFactoidState(filename string) state.IFactoidState{
    fs := new(state.FactoidState)
    wall := new(wallet.SCWallet)
    wall.Init()

    fs.SetWallet(wall)
    fs.GetWallet().GetDB().Init()
    
    // Use Bolt DB
    if true {
        fs.SetDB(new(database.MapDB))
        fs.GetDB().Init()
        db := GetDatabase(filename)
        fs.GetDB().SetPersist(db)
        fs.GetDB().SetBacker(db)
        fs.GetWallet().GetDB().SetPersist(db)
        fs.GetWallet().GetDB().SetBacker(db)
        
        fs.GetDB().DoNotPersist(fct.DB_F_BALANCES)
        fs.GetDB().DoNotPersist(fct.DB_EC_BALANCES)
        fs.GetDB().DoNotPersist(fct.DB_BUILD_TRANS)
        fs.GetDB().DoNotCache(fct.DB_FACTOID_BLOCKS)
        fs.GetDB().DoNotCache(fct.DB_TRANSACTIONS)
        
    }else{
        fs.SetDB(GetDatabase(filename))
    }
    
    return fs
}

func GetDatabase(filename string) database.IFDatabase {
    
    var bucketList [][]byte
    var instances  map[[fct.ADDRESS_LENGTH]byte]fct.IBlock
    
    bucketList = make([][]byte,0,5)
    
    bucketList = append(bucketList,[]byte(fct.DB_FACTOID_BLOCKS))
    bucketList = append(bucketList,[]byte(fct.DB_F_BALANCES))
    bucketList = append(bucketList,[]byte(fct.DB_EC_BALANCES))
    
    bucketList = append(bucketList,[]byte(fct.DB_BUILD_TRANS))    
    bucketList = append(bucketList,[]byte(fct.DB_TRANSACTIONS))    
    bucketList = append(bucketList,[]byte(fct.W_RCD_ADDRESS_HASH))    
    bucketList = append(bucketList,[]byte(fct.W_ADDRESS_PUB_KEY))    
    bucketList = append(bucketList,[]byte(fct.W_NAME))    
        
    instances = make(map[[fct.ADDRESS_LENGTH]byte]fct.IBlock)
    
    var addinstance = func  (b fct.IBlock){
        key := new ([32]byte)
        copy(key[:],b.GetDBHash().Bytes())
        instances[*key] = b 
    }
    addinstance (new(fct.Address))
    addinstance (new(fct.Hash))
    addinstance (new(fct.InAddress))
    addinstance (new(fct.OutAddress))
    addinstance (new(fct.OutECAddress))
    addinstance (new(fct.RCD_1))
    addinstance (new(fct.RCD_2))
    addinstance (new(fct.Signature))
    addinstance (new(fct.Transaction))
    addinstance (new(block.FBlock))
    addinstance (new(state.FSbalance))
    addinstance (new(wallet.WalletEntry))
 
    db := new(database.BoltDB)
    db.Init(bucketList,instances,filename)
    return db
}
