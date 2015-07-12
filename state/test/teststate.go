// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package test

import (
    "fmt"
    "time"
    "math/rand"
    fct "github.com/FactomProject/factoid"    
    "github.com/FactomProject/factoid/state"    
    "github.com/FactomProject/factoid/wallet"    
)

var _ = fmt.Printf

type Stats struct {
    badAddresses int
    transactions int
    errors map[string]int
    start time.Time
    blocktimes []time.Time
}
func (s Stats) begin() {
    s.start = time.Now()
}
func (s Stats) endBlock() {
    s.blocktimes = append(s.blocktimes, time.Now())
}
func (s Stats) logError(err string) {
    if s.errors == nil {
        s.errors = make(map[string]int)
    }
    cnt := s.errors[err]
    s.errors[err] = cnt+1
}

type Test_state struct {
    state.FactoidState
    clock int64
    twallet wallet.ISCWallet
    inputAddresses []fct.IAddress        // Genesis Address funds 10 addresses
    outputAddresses []fct.IAddress       // We consider our inputs and ten more addresses
    stats Stats
}

func(fs *Test_state) GetWallet() wallet.ISCWallet {
    return fs.twallet
}

func(fs *Test_state) GetTime64() int64 {
    return time.Now().UnixNano()
}

func(fs *Test_state) GetTime32() int64 {
    return time.Now().Unix()
}

func(fs *Test_state) newTransaction() fct.ITransaction {
    
    fs.inputAddresses = make([]fct.IAddress,0,20)
    for _,output := range fs.outputAddresses {
        bal := fs.GetBalance(output)
        if bal > 1000000 {
            fs.inputAddresses = append(fs.inputAddresses, output)
        }
    }
    // The following code is a function that creates an array
    // of addresses pulled from some source array of addresses
    // selected randomly.
    var makeList = func(source []fct.IAddress, cnt int) []fct.IAddress{
        adrs := make([]fct.IAddress,0,cnt)
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
    t := fs.twallet.CreateTransaction(fs.GetTimeMilli())
    for _, adr := range inputs {
        balance := fs.GetBalance(adr)
        toPay := balance
        if balance > 1000000 {
            toPay = balance >> 8
        }
        paid = toPay+paid
        fs.twallet.AddInput(t,adr, toPay)
        
        //fmt.Printf("%s %x\n",adr.String(),balance)
    
    }
    for _, adr := range outputs {
        fs.twallet.AddOutput(t,adr,paid/uint64(len(outputs)))
    }
    fee,_ := t.CalculateFee(fs.GetFactoshisPerEC())
    fs.twallet.UpdateInput(t,0,inputs[0], (paid/uint64(len(inputs)))+fee)
    
    valid, err := fs.twallet.SignInputs(t)
    if err != nil {
        fct.Prtln("Failed to sign transaction")
        panic(err)
    }
    if !valid {
        fct.Prtln("Transaction is not valid")
    }
    if !fs.Validate(t) {
        fs.stats.badAddresses += 1
        return fs.newTransaction() 
    }
    return t
}