// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wallet

import fct "github.com/FactomProject/factoid"
import (
	"encoding/hex"
    "encoding/binary"
	"fmt"
    "github.com/FactomProject/factoid"
    "github.com/FactomProject/ed25519"
	"math/rand"
	"testing"
    
)

var _ = hex.EncodeToString
var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Write
var _ = fct.Prtln   
 

func Test_create_scwallet(test *testing.T) {
    w := new(SCWallet)          // make me a wallet
    w.Init()
    we := new(WalletEntry)
    rcd := new(factoid.RCD_1)
    name := "John Smith"
    pub, pri, err := w.generateKey()
    
    if err != nil {
        factoid.Prtln("Generate Failed")
        test.Fail()
    }
    
    we.SetRCD(rcd)
    we.AddKey(pub,pri)
    we.SetName([]byte(name))
    
    txt,err := we.MarshalText()
    var _ = txt
   // factoid.Prtln(string(txt))
   
   
}

func Test_GenerateAddress_scwallet(test *testing.T) {
    w := new(SCWallet)          // make me a wallet
    w.Init()
    h1,err := w.GenerateFctAddress([]byte("test 1"),1,1)
    if err != nil { test.Fail(); return}
    h2,err := w.GenerateFctAddress([]byte("test 2"),1,1)
    if err != nil { test.Fail(); return }
    
    if h1.IsEqual(h2) == nil { 
        test.Fail() 
    }   
    
    if h1.IsEqual(h1) != nil { test.Fail() }
}

func Test_CreateTransaction_swcallet(test *testing.T) { 
    w := new(SCWallet)          // make me a wallet
    w.Init()
    h1,err := w.GenerateFctAddress([]byte("test 1"),1,1)
    if err != nil { test.Fail() }
    h2,err := w.GenerateFctAddress([]byte("test 2"),1,1)
    if err != nil { test.Fail() }
    
    t := w.CreateTransaction(0)
    
    w.AddInput(t,h1,1000000)
    w.AddOutput(t,h2,1000000-12000)
    
    signed,err := w.SignInputs(t)
    if !signed || err != nil {
        factoid.Prtln("Signed Fail: ",signed, err)
        test.Fail()
    }
    
    fee, err := t.CalculateFee(1000)
    if fee != 12000 || err != nil {
        factoid.Prtln("Fee Calculation Failed",fee,err)
        test.Fail() 
    }
    
    err2 := w.Validate(1,t)
    if( err2 != nil) {
        factoid.Prtln(err2)
        test.Fail()
    }
    
}

func Test_SignTransaction_swcallet(test *testing.T) { 
    w := new(SCWallet)          // make me a wallet
    w.Init()
    h0,err := w.GenerateFctAddress([]byte("test 0"),1,1)
    if err != nil { test.Fail() }
    h1,err := w.GenerateFctAddress([]byte("test 1"),1,1)
    if err != nil { test.Fail() }
    h2,err := w.GenerateFctAddress([]byte("test 2"),1,1)
    if err != nil { test.Fail() }
    h3,err := w.GenerateFctAddress([]byte("test 3"),1,1)
    if err != nil { test.Fail() }
    h4,err := w.GenerateFctAddress([]byte("test 4"),1,1)
    if err != nil { test.Fail() }
    
    t := w.CreateTransaction(0)
    
    w.AddInput(t,h1,1000000)
    w.AddInput(t,h2,1000000)
    w.AddOutput(t,h3,1000000)
    w.AddOutput(t,h4,1000000)
    w.AddInput(t,h0,0)
    fee, err := t.CalculateFee(1000)
    w.UpdateInput(t,2,h0,fee)
    signed,err := w.SignInputs(t)
    
    if !signed || err != nil {
        factoid.Prtln("Signed Fail: ",signed, err)
        test.Fail()
    }
    
    txt,err := t.MarshalText()
    fct.Prtln(string(txt), "\n ", fee )
    
    err = w.ValidateSignatures(t)
    if err != nil {
        factoid.Prtln(err)
        test.Fail()
    }
    
}
