// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wallet

import sc "github.com/FactomProject/simplecoin"
import (
	"encoding/hex"
    "encoding/binary"
	"fmt"
    "github.com/FactomProject/simplecoin"
    "github.com/agl/ed25519"
	"math/rand"
	"testing"
    
)

var _ = hex.EncodeToString
var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Write
var _ = sc.Prtln   
 

func Test_create_scwallet(test *testing.T) {
    w := new(SCWallet)          // make me a wallet
    we := new(WalletEntry)
    rcd := new(simplecoin.RCD_1)
    name := "John Smith"
    pub, pri, err := w.generateKey()
    
    if err != nil {
        simplecoin.Prtln("Generate Failed")
        test.Fail()
    }
    
    we.SetRCD(rcd)
    we.AddKey(pub,pri)
    we.SetName([]byte(name))
    
    txt,err := we.MarshalText()
    var _ = txt
   // simplecoin.Prtln(string(txt))
}

func Test_GenerateAddress_scwallet(test *testing.T) {
    w := new(SCWallet)          // make me a wallet
    h1,err := w.GenerateAddress([]byte("test 1"),1,1)
    if err != nil { test.Fail() }
    h2,err := w.GenerateAddress([]byte("test 2"),1,1)
    if err != nil { test.Fail() }
    
    if h1.IsEqual(h2) { test.Fail() }   
    
    if !h1.IsEqual(h1) { test.Fail() }
}

func Test_CreateTransaction_swcallet(test *testing.T) { 
    w := new(SCWallet)          // make me a wallet
    h1,err := w.GenerateAddress([]byte("test 1"),1,1)
    if err != nil { test.Fail() }
    h2,err := w.GenerateAddress([]byte("test 2"),1,1)
    if err != nil { test.Fail() }
    
    t := w.CreateTransaction()
    
    w.AddInput(t,h1,1000000)
    w.AddOutput(t,h2,1000000-12000)
    
    signed,err := w.SignInputs(t)
    if !signed || err != nil {
        simplecoin.Prtln("Signed Fail: ",signed, err)
        test.Fail()
    }
    
    fee, err := t.CalculateFee(1000)
    if fee != 12000 || err != nil {
        simplecoin.Prtln("Fee Calculation Failed",fee,err)
        test.Fail() 
    }
    
    valid, err2 := w.Validate(t)
    if(!valid || err2 != nil) {
        simplecoin.Prtln(err2,valid)
        test.Fail()
    }
    
}