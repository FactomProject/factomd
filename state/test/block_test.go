// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
    "encoding/hex"
    "encoding/binary"
    "fmt"
    cv "strconv"
    sc "github.com/FactomProject/simplecoin"
    "github.com/FactomProject/simplecoin/state/stateinit"
    "github.com/FactomProject/simplecoin/state"
    "github.com/FactomProject/simplecoin/database"
    "github.com/FactomProject/simplecoin/wallet"
    "github.com/agl/ed25519"
    "math/rand"
    "testing"
)

var _ = state.FactomState{}
var _ = hex.EncodeToString
var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Write
var _ = sc.Prtln 
var _ = stateinit.GetDatabase
var _ = database.MapDB{}

var fs *test_state
// sets up teststate.go                                         
func Test_setup_factomstate (test *testing.T) {
    // Create a Test State
    fs = new(test_state)
    
    fs.inputAddresses = make([]sc.IAddress,0,10)
    fs.outputAddresses = make([]sc.IAddress,0,10)
    fs.twallet = new(wallet.SCWallet)              // Wallet for our tests
    fs.twallet.Init()
    
    for i:=0; i<10; i++ {
        addr, err := fs.twallet.GenerateAddress([]byte("tes,mbbm,btin_"+cv.Itoa(i)),1,1)
        if err != nil { sc.Prtln(err); test.Fail() }
        fs.inputAddresses = append(fs.inputAddresses,addr)
        fs.outputAddresses = append(fs.outputAddresses,addr)
    }
    for i:=0; i<10; i++ {
        addr, err := fs.twallet.GenerateAddress([]byte("testout_"+cv.Itoa(i)),1,1)
        if err != nil { sc.Prtln(err); test.Fail() }
        fs.outputAddresses = append(fs.outputAddresses,addr)
    }
}


func Test_create_genesis_factomstate (test *testing.T) {
        
    // Use Bolt DB
    if true {
        fs.SetDB(new(database.MapDB))
        fs.GetDB().Init()
        db := stateinit.GetDatabase()
        fs.GetDB().SetPersist(db)
        fs.GetDB().SetBacker(db)
        fs.GetDB().DoNotPersist(sc.DB_F_BALANCES)
        fs.GetDB().DoNotPersist(sc.DB_EC_BALANCES)
         
    }else{
        fs.SetDB(stateinit.GetDatabase())
    }
    // Set the price for Factoids
    fs.SetFactoshisPerEC(10000)
   
    err := fs.LoadState()
    if err != nil {
        sc.Prtln(err)
        test.Fail()
        return
    }
    // Create a number of blocks (i)
    for i:=0; i<30000; i++ {
        sc.Prt(i," ")
        // Create a new block
        for j:=0; j<10; j++ {
            t := fs.newTransaction()
            added := fs.AddTransaction(t)
            if !added { 
                sc.Prt("F:",i,"-",j," ",t) 
            }
        }
        fs.ProcessEndOfBlock()
    }
    fmt.Println("\nDone")
}

func Test_build_blocks_factomstate (test *testing.T) {
    
    
}


