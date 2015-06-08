// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
    "fmt"
    "time"
    "math/rand"
    sc "github.com/FactomProject/simplecoin"    
    "github.com/FactomProject/simplecoin/state"    
    "github.com/FactomProject/simplecoin/wallet"    
)

var _ = fmt.Printf


type test_state struct {
    state.FactomState
    clock int64
    twallet wallet.ISCWallet
    inputAddresses []sc.IAddress        // Genesis Address funds 10 addresses
    outputAddresses []sc.IAddress       // We consider our inputs and ten more addresses
    // as valid outputs.
    
}

func(fs *test_state) GetTime64() int64 {
    return time.Now().UnixNano()
}

func(fs *test_state) GetTime32() int64 {
    return time.Now().Unix()
}

func(fs *test_state) newTransaction() sc.ITransaction {
    
    fs.inputAddresses = make([]sc.IAddress,0,20)
    for _,output := range fs.outputAddresses {
        bal := fs.GetBalance(output)
        if bal > 100000 {
            fs.inputAddresses = append(fs.inputAddresses, output)
        }
    }
    // The following code is a function that creates an array
    // of addresses pulled from some source array of addresses
    // selected randomly.
    var makeList = func(source []sc.IAddress, cnt int) []sc.IAddress{
        adrs := make([]sc.IAddress,0,cnt)
        MainLoop: for len(adrs)<cnt {
            i := rand.Int()%len(source)
            adr := source[i]
            for _,adr2 := range adrs {
                if adr.IsEqual(adr2) == nil {
                    continue MainLoop
                }
            }
            adrs = append(adrs,adr)
        }
        return adrs
    }

    // Get one to five inputs, and one to five outputs
    numInputs := rand.Int()%5+1
    numOutputs := rand.Int()%5+1
    
    // Get my input and output addresses
    inputs := makeList(fs.inputAddresses,numInputs)
    outputs := makeList(fs.outputAddresses,numOutputs)

    var paid uint64
    t := fs.twallet.CreateTransaction()
    for _, adr := range inputs {
        balance := fs.GetBalance(adr)
        toPay := balance >> 16 
        paid = toPay+paid
        fs.twallet.AddInput(t,adr, toPay)
        
        //fmt.Printf("%s %x\n",adr.String(),balance)
    
    }
    for _, adr := range outputs {
        fs.twallet.AddOutput(t,adr,paid/uint64(len(outputs)))
    }
    fee,err := t.CalculateFee(fs.GetFactoshisPerEC())
    fs.twallet.UpdateInput(t,0,inputs[0], (paid/uint64(len(inputs)))+fee)
    
    valid, err := fs.twallet.SignInputs(t)
    if err != nil {
        sc.Prtln("Failed to sign transaction")
        panic(err)
    }
    if !valid {
        sc.Prtln("Transaction is not valid")
    }
    if !fs.Validate(t) {return fs.newTransaction() }
    return t
}