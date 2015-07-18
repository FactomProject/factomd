// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package test

import (
    "encoding/hex"
    "encoding/binary"
    "fmt"
    "time"
    cv "strconv"
    fct "github.com/FactomProject/factoid"
    "github.com/FactomProject/factoid/state/stateinit"
    "github.com/FactomProject/factoid/state"
    "github.com/FactomProject/factoid/block"
    "github.com/FactomProject/factoid/database"
    "github.com/FactomProject/factoid/wallet"
    "github.com/FactomProject/ed25519"
    "math/rand"
    "testing"
)

var _ = time.Second
var _ = state.FactoidState{}
var _ = hex.EncodeToString
var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Write
var _ = fct.Prtln 
var _ = stateinit.GetDatabase
var _ = database.MapDB{}

var fs *Test_state
// sets up teststate.go                                         
func Test_setup_FactoidState (test *testing.T) {
    // Create a Test State
    fs = new(Test_state)
    
    fs.inputAddresses = make([]fct.IAddress,0,10)
    fs.outputAddresses = make([]fct.IAddress,0,10)
    fs.twallet = new(wallet.SCWallet)              // Wallet for our tests
    fs.twallet.Init()
    
    for i:=0; i<10; i++ {
        addr, err := fs.twallet.GenerateFctAddress([]byte("testin_"+cv.Itoa(i)),1,1)
        if err != nil { fct.Prtln(err); test.Fail() }
        fs.inputAddresses = append(fs.inputAddresses,addr)
        fs.outputAddresses = append(fs.outputAddresses,addr)
    }
    for i:=0; i<1000; i++ {
        addr, err := fs.twallet.GenerateFctAddress([]byte("testout_"+cv.Itoa(i)),1,1)
        if err != nil { fct.Prtln(err); test.Fail() }
        fs.outputAddresses = append(fs.outputAddresses,addr)
    }
}


func Test_create_genesis_FactoidState (test *testing.T) {
     
    numBlocks       := 5000
    numTransactions := 200
    maxIn           := 10
    maxOut          := 20
    if testing.Short() {
        fmt.Print("\nDoing Short Tests\n")
        numBlocks       = 5
        numTransactions = 20
        maxIn           = 5
        maxOut          = 5
    }
    
    // Use Bolt DB
    if false {
        fs.SetDB(new(database.MapDB))
        fs.GetDB().Init()
        db := stateinit.GetDatabase("/tmp/fct_test.db")
        fs.GetDB().SetPersist(db)
        fs.GetDB().SetBacker(db)
        
        fs.GetDB().DoNotPersist(fct.DB_F_BALANCES)
        fs.GetDB().DoNotPersist(fct.DB_EC_BALANCES)
        fs.GetDB().DoNotPersist(fct.DB_BUILD_TRANS)
        fs.GetDB().DoNotCache(fct.DB_FACTOID_BLOCKS)
        fs.GetDB().DoNotCache(fct.DB_TRANSACTIONS)
    }else{
        fs.SetDB(new(database.MapDB))
        fs.GetDB().Init()
    }
    // Set the price for Factoids
    fs.SetFactoshisPerEC(10000)
    fct.Prt("Loading....")
    err := fs.LoadState()
    if err != nil {
        fct.Prtln("Failed to load:", err)
        test.Fail()
        return
    }
    var cnt,max,min,maxblk int
    min = 100000
    // Create a number of blocks (i)
    for i:=0; i<numBlocks; i++ {
        
        periodMark := 1
        fct.Prt(" ",fs.GetDBHeight(),":",cnt,"--",fs.stats.badAddresses)
        // Create a new block
        for j:=cnt; cnt < j+numTransactions; {      // Execute for some number RECORDED transactions
            
            if periodMark <=10 && cnt%2==0 {
                fs.EndOfPeriod(periodMark)
                periodMark++
            }
            
            tx := fs.newTransaction(maxIn,maxOut)
            
            // Test Marshal/UnMarshal
            m,err := tx.MarshalBinary()
            if err != nil { fmt.Println("\n Failed to Marshal: ",err); test.Fail(); return } 
            if len(m) > max { 
                max = len(m)
                fmt.Println("Transaction",cnt,"is",len(m),"Bytes long. ",
                            len(tx.GetInputs()), "inputs and",
                            len(tx.GetOutputs()),"outputs and",
                            len(tx.GetECOutputs()),"ecoutputs",)
            }
            if len(m) < min { 
                min = len(m)
                fmt.Println("Transaction",cnt,"is",len(m),"Bytes long. ",
                            len(tx.GetInputs()), "inputs and",
                            len(tx.GetOutputs()),"outputs and",
                            len(tx.GetECOutputs()),"ecoutputs",)
            }
            good := true
            k := rand.Int()%(len(m)-2)
            k++
            flip := rand.Int()%100
            // To simulate bad data, I mess up some of the data here.
            if rand.Int()%100 < 5 { // Mess up 5 percent of the transactions
                good = false
                if flip < 49 {    // Flip a coin
                    m = m[k:]
                }else{
                    m = m[:k]
                }
            }
            
            t := new(fct.Transaction)
            err = t.UnmarshalBinary(m)
            
            if good && tx.IsEqual(t) != nil { 
                fmt.Println("\nFail valid Unmarshal")
                test.Fail()
                return
            }
            if err == nil {
                added := fs.AddTransaction(t)
                if good != added  { 
                    if good {
                        fmt.Println("Failed to add a transaction that should have added")
                    }else{
                        fmt.Println("Added a transaction that should have failed to be added")
                    }
                    test.Fail(); 
                    return 
                }
                
            }
            
            if good && err != nil {         
                fmt.Println("\nUnmarshal Failed. trans is good",
                            "\nand the error detected: ",err,
                            "\nand k:",k, "and flip:",flip)
                test.Fail() 
                return 
            } 
            
            if good {
                fmt.Print("\rBad Transactions: ",fs.stats.badAddresses,"   Total transactions: ",cnt,"\r")
                time.Sleep(9000)
                cnt += 1
            }else{
                fs.stats.badAddresses += 1
            }
            
        }
        //
        // Serialization deserialization tests for blocks
        //
        blkdata,err := fs.GetCurrentBlock().MarshalBinary()
        if err != nil { test.Fail(); return }
        blk := fs.GetCurrentBlock().GetNewInstance().(block.IFBlock)
        err = blk.UnmarshalBinary(blkdata)
        if err != nil { test.Fail(); return }
        if len(blkdata)>maxblk {
            fmt.Println("Block",blk.GetDBHeight(),"is",len(blkdata),"bytes")
        }
//         blk:=fs.GetCurrentBlock()       // Get Current block, but hashes are set by processing.
        fs.ProcessEndOfBlock()             // Process the block.
//         fmt.Println(blk)                // Now print it.
        
    }
    fmt.Println("\nDone")
}

func Test_build_blocks_FactoidState (test *testing.T) {
    
    
}


