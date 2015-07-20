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
    fs.ecoutputAddresses = make([]fct.IAddress,0,10)
    fs.twallet = new(wallet.SCWallet)              // Wallet for our tests
    fs.twallet.Init()
    
    for i:=0; i<10; i++ {
        addr, err := fs.twallet.GenerateFctAddress([]byte("testin_"+cv.Itoa(i)),1,1)
        if err != nil { fct.Prtln(err); test.Fail() }
        fs.inputAddresses = append(fs.inputAddresses,addr)
        fs.outputAddresses = append(fs.outputAddresses,addr)
    }
    for i:=0; i<500; i++ {
        addr, err := fs.twallet.GenerateFctAddress([]byte("testout_"+cv.Itoa(i)),1,1)
        if err != nil { fct.Prtln(err); test.Fail() }
        fs.outputAddresses = append(fs.outputAddresses,addr)
    }
    for i:=0; i<50; i++ {
        addr, err := fs.twallet.GenerateECAddress([]byte("testecout_"+cv.Itoa(i)))
        if err != nil { fct.Prtln(err); test.Fail() }
        fs.ecoutputAddresses = append(fs.outputAddresses,addr)
    }
}


func Test_create_genesis_FactoidState (test *testing.T) {
    fmt.Print("\033[2J")
    
    numBlocks       := 5000
    numTransactions := 500
    maxIn           := 5
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
    fs.SetFactoshisPerEC(100000)
    err := fs.LoadState()
    if err != nil {
        fct.Prtln("Failed to load:", err)
        test.Fail()
        return
    }
    
    // Make the coinbase very generous
    block.UpdateAmount(100000000000)
    
    var cnt,max,min,maxblk int
    min = 100000
    // Create a number of blocks (i)
    for i:=0; i<numBlocks; i++ {
        
        periodMark := 1
        // Create a new block
        for j:=cnt; cnt < j+numTransactions; {      // Execute for some number RECORDED transactions
            
            if periodMark <=10 && cnt%(numTransactions/10)==0 {
                fs.EndOfPeriod(periodMark)
                periodMark++
            }
            
            tx := fs.newTransaction(maxIn,maxOut)
            
            // Test Marshal/UnMarshal
            m,err := tx.MarshalBinary()
            if err != nil { fmt.Println("\n Failed to Marshal: ",err); test.Fail(); return } 
            if len(m) > max { 
                fmt.Print("\033[33;0H")
                max = len(m)
                fmt.Println("Max Transaction",cnt,"is",len(m),"Bytes long. ",
                            len(tx.GetInputs()), "inputs and",
                            len(tx.GetOutputs()),"outputs and",
                            len(tx.GetECOutputs()),"ecoutputs                       ",)
                fmt.Print("\033[40;0H")
            }
            if len(m) < min { 
                fmt.Print("\033[34;0H")
                min = len(m)
                fmt.Println("Min Transaction",cnt,"is",len(m),"Bytes long. ",
                            len(tx.GetInputs()), "inputs and",
                            len(tx.GetOutputs()),"outputs and",
                            len(tx.GetECOutputs()),"ecoutputs                       ",)
                fmt.Print("\033[40;0H")
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
                if good && err != nil  { 
                    fmt.Println("Added a transaction that should have failed to be added")
                    fmt.Println(err)
                    test.Fail(); 
                }
                if !good {
                    fmt.Println("Failed to add a transaction that should have added")
                    test.Fail(); 
                }
            }
            
            if good {
                err = fs.AddTransaction(t)
            }
            
            if good && err != nil {   
                fmt.Println(err)
                fmt.Println("\nUnmarshal Failed. trans is good",
                            "\nand the error detected: ",err,
                            "\nand k:",k, "and flip:",flip)
                test.Fail() 
                return 
            } 
            
            if good {
                fmt.Print("\033[32;0H")
                fmt.Println("Bad Transactions: ",fs.stats.badAddresses,"   Total transactions: ",cnt,"\r")
                fmt.Print("\033[40;0H")
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
            fmt.Printf("\033[%d;%dH",(blk.GetDBHeight())%30+1, (((blk.GetDBHeight())/30)%5)*25+1)
            fmt.Printf("Blk:%6d %8d B ",blk.GetDBHeight(),len(blkdata))
            fmt.Printf("\033[%d;%dH",(blk.GetDBHeight())%30+2, (((blk.GetDBHeight())/30)%5)*25+1)
            fmt.Printf("%24s","=====================    ")
        }
//         blk:=fs.GetCurrentBlock()       // Get Current block, but hashes are set by processing.
        fs.ProcessEndOfBlock()             // Process the block.
//         fmt.Println(blk)                // Now print it.
        
    }
    fmt.Println("\nDone")
}

func Test_build_blocks_FactoidState (test *testing.T) {
    
    
}


