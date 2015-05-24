// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wallet

import (
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/simplecoin"
	"github.com/agl/ed25519"
	"math/rand"
	"testing"
)

var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Write

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
}

func Test_GenerateAddress_scwallet(test *testing.T) {
    w := new(SCWallet)          // make me a wallet
    h1 := w.GenerateAddress([]byte("test 1"),1,1)
    h2 := w.GenerateAddress([]byte("test 1"),1,1)
    
    if h1.IsEqual(h2) { test.Fail() }   
    
    //    GetAddressDetailsAddr(addr []byte)
}

func Test_AddInput_scallet(test *testing.T) { }