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

var twallet wallet.ISCWallet
var inputAddresses []sc.IAddress        // Genesis Address funds 10 addresses
var outputAddresses []sc.IAddress       // We consider our inputs and ten more addresses
                                        // as valid outputs.

type test_state struct {
    state.FactomState
    clock int64
}

func(fs *test_state) GetTime64() int64 {
    return time.Now().UnixNano()
}

func(fs *test_state) GetTime32() int64 {
    return time.Now().Unix()
}

func newTransaction(fs state.IFactomState) sc.ITransaction {
    
    // The following code is a function that creates an array
    // of addresses pulled from some source array of addresses
    // selected randomly.
    var makeList = func(source []sc.IAddress, cnt int) []sc.IAddress{
        adrs := make([]sc.IAddress,0,cnt)
        MainLoop: for len(adrs)<cnt {
            i := rand.Int()%len(source)
            adr := source[i]
            for _,adr2 := range adrs {
                if adr.IsEqual(adr2) {
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
    inputs := makeList(inputAddresses,numInputs)
    outputs := makeList(outputAddresses,numOutputs)

    transAmount := uint64(int64(1000000000) - rand.Int63() % int64(900000000))
    
    t := twallet.CreateTransaction()
    for _, adr := range inputs {
        twallet.AddInput(t,adr,transAmount/uint64(len(inputs)))
    }
    for _, adr := range outputs {
        twallet.AddOutput(t,adr,transAmount/uint64(len(outputs)))
    }
    fee,err := t.CalculateFee(fs.GetFactoshisPerEC())
    twallet.UpdateInput(t,0,inputs[0], (transAmount/uint64(len(inputs)))+fee)
    
    valid, err := twallet.  SignInputs(t)
    if err != nil {
        panic(err)
    }
    if !valid {
        panic("Transaction is not valid")
    }
  
    return t
}