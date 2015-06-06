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
    "github.com/FactomProject/simplecoin/wallet"
    "github.com/FactomProject/simplecoin/block"
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

// sets up teststate.go                                         
func Test_setup_factomstate (test *testing.T) {
    inputAddresses = make([]sc.IAddress,0,10)
    outputAddresses = make([]sc.IAddress,0,10)
    twallet = new(wallet.SCWallet)              // Wallet for our tests
    twallet.Init()
    
    for i:=0; i<10; i++ {
        addr, err := twallet.GenerateAddress([]byte("tes,mbbm,btin_"+cv.Itoa(i)),1,1)
        if err != nil { sc.Prtln(err); test.Fail() }
        inputAddresses = append(inputAddresses,addr)
        outputAddresses = append(outputAddresses,addr)
    }
    for i:=0; i<10; i++ {
        addr, err := twallet.GenerateAddress([]byte("testout_"+cv.Itoa(i)),1,1)
        if err != nil { sc.Prtln(err); test.Fail() }
        outputAddresses = append(outputAddresses,addr)
    }
}

var TransBlkHead = sc.NewHash(sc.FACTOID_CHAINID)
func Test_create_genesis_factomstate (test *testing.T) {
    
    dbheight := uint32(1)
    
    // Create a Test State
    fs := new(test_state)
    // Use Bolt DB
    fs.SetDB(stateinit.GetDatabase())
    // Set the price for Factoids
    fs.SetFactoshisPerEC(10000)
    
    // The first block is the genesis block
    lb := block.GetGenesisBlock(1000000,10,200000000000)
    lb.SetDBHeight(dbheight)
    fs.PutTransactionBlock(lb.GetHash(),lb)
    fs.PutTransactionBlock(TransBlkHead,lb)
    // Create a number of blocks (i)
    for i:=0; i<3; i++ {
        // Create a new block
        nb := block.NewSCBlock(fs.GetFactoshisPerEC(),1)
        dbheight += 1
        nb.SetDBHeight(dbheight)
        // Link to the previous block
        nb.SetPrevBlock(lb.GetHash().Bytes())
        for i:=0; i<1; i++ {
            t := newTransaction(fs)
            data,err := t.MarshalBinary()
            t2 := new(sc.Transaction)
            err = t2.UnmarshalBinary(data)
            if err != nil { test.Fail() }
            if !t2.IsEqual(t) { test.Fail() }
            valid, err := nb.AddTransaction(t2)
            if err != nil { sc.Prtln(err); test.Fail() }
            if !valid { sc.Prtln("All transactions should be valid"); test.Fail() }
        }
        // Save it away
        fs.PutTransactionBlock(nb.GetHash(),nb)
        nb2 := fs.GetTransactionBlock(nb.GetHash())
        data, err := nb.MarshalBinary()
        if err != nil { sc.Prtln("Unmarshal test failed"); test.Fail(); continue }
        nb4 := new(block.SCBlock)
        err = nb4.UnmarshalBinary(data)
        if err != nil { sc.Prtln("Marshal test failed"); test.Fail(); continue }
        if !nb.IsEqual(nb4) { sc.Prtln("Just Marshaling is failing",nb,nb4); test.Fail(); continue }
        if !nb.IsEqual(nb2) { sc.Prtln("Failed to get block back"); test.Fail(); continue }
        fs.PutTransactionBlock(TransBlkHead,nb)
        nb3 := fs.GetTransactionBlock(TransBlkHead)
        if !nb.IsEqual(nb3) { sc.Prtln("Failed to get the block head back"); test.Fail() }
        lb = nb // The new block is now the last block
    }
    fmt.Println("Build complete")
    fs.LoadState()
}

func Test_build_blocks_factomstate (test *testing.T) {
    
    
}


