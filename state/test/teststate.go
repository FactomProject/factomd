// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package test

import (
    "fmt"
    "time"
    "bytes"
    "strings"
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
    full map[string]string
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
    ecoutputAddresses []fct.IAddress     // Entry Credit Addresses
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

func(fs *Test_state) newTransaction(maxIn, maxOut int) fct.ITransaction {
    var max, max2 uint64
    fs.inputAddresses = make([]fct.IAddress,0,20)
    for _,output := range fs.outputAddresses {
        bal := fs.GetBalance(output)
        if bal > 1000000000 {
            fs.inputAddresses = append(fs.inputAddresses, output)
        }
        if max < bal {
            max2 = max
            max = bal
        }else{
            if max2 < bal {
                max2 = bal 
            }
        }        
    }
    
    fmt.Printf("\033[35;0H Inputs %4d, Max %20s Max2 %20s            ",
               len(fs.inputAddresses),
               strings.TrimSpace(fct.ConvertDecimal(max)),
               strings.TrimSpace(fct.ConvertDecimal(max2)))
    fmt.Print("\033[40;0H")
    
    // The following code is a function that creates an array
    // of addresses pulled from some source array of addresses
    // selected randomly.
    var makeList = func(source []fct.IAddress, cnt int) []fct.IAddress{
        adrs := make([]fct.IAddress,0,cnt)
        for len(adrs)<cnt {
            i := rand.Int()%len(source)
            adr := source[i]
            adrs = append(adrs,adr)
        }
        return adrs
    }

    mIn := maxIn
    mOut := maxOut
    
    joker := rand.Int()%100
    if joker < 1 { mIn = maxIn*100 }
    joker = rand.Int()%100
    if joker < 1 { mOut = maxOut*200 }
    
    // Get one to five inputs, and one to five outputs
    numInputs := rand.Int()%mIn+1
    numOutputs := rand.Int()%mOut
    mumECOutputs := rand.Int()%mOut
 
 
 
    numInputs = (numInputs%(len(fs.inputAddresses)-2))+1

   // fmt.Println("inputs outputs",numInputs,numOutputs, "limits",len(fs.inputAddresses),len(fs.outputAddresses))
    
    
    // Get my input and output addresses
    inputs := makeList(fs.inputAddresses,numInputs)
    outputs := makeList(fs.outputAddresses,numOutputs)
    ecoutputs := makeList(fs.ecoutputAddresses,mumECOutputs)
    var paid uint64
    t := fs.twallet.CreateTransaction(fs.GetTimeMilli())
    for _, adr := range inputs {
        balance := fs.GetBalance(adr)
        toPay := uint64(rand.Int63())%(balance)
        paid = toPay+paid
        fs.twallet.AddInput(t,adr, toPay)
        //fmt.Print("\033[10;3H")
        //fmt.Printf("%s %s    \n",adr.String(),fct.ConvertDecimal(toPay))
        //fmt.Print("\033[40;3H")
    }
    
    paid = paid - fs.GetFactoshisPerEC()*uint64(len(ecoutputs))
    
    for _, adr := range outputs {
        fs.twallet.AddOutput(t,adr,paid/uint64(len(outputs)))
    }
    
    for _, adr := range ecoutputs {
        fs.twallet.AddECOutput(t,adr,fs.GetFactoshisPerEC())
    }
    
    fee,_ := t.CalculateFee(fs.GetFactoshisPerEC())
    toPay := t.GetInputs()[0].GetAmount()
    fs.twallet.UpdateInput(t,0,inputs[0], toPay+fee)
        
    valid, err1 := fs.twallet.SignInputs(t)
    if err1 != nil {
        fct.Prtln("Failed to sign transaction")
    }
    if !valid {
        fct.Prtln("Transaction is not valid")
    }
    if err := fs.Validate(len(fs.GetCurrentBlock().GetTransactions()),t); err != nil || err1 != nil {
        
        fs.GetDB().Put(fct.DB_BAD_TRANS, t.GetHash(), t)
        
        fs.stats.badAddresses += 1
        
        str := []byte(err.Error())[:10]
        if bytes.Compare(str,[]byte("The inputs"))!=0 {
            str = []byte(err.Error())[:30]
        }
        fs.stats.errors[string(str)] += 1
        fs.stats.full[string(str)] = err.Error()
        
        fmt.Print("\033[32;0H")
        fmt.Println("Bad Transactions: ",fs.stats.badAddresses,"\r")
        fmt.Print("\033[40;0H")
        return fs.newTransaction(maxIn,maxOut) 
    }
    return t
}