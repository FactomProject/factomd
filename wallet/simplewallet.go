// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// This is a minimum wallet to be used to test the coin
// There isn't much in the way of interest in security 
// here, but rather provides a mechanism to create keys
// and sign transactions, etc.

package wallet

import (
    "fmt"
    "github.com/agl/ed25519"
    "math/rand"
    "testing"
)

type ISCWallet interface {
    GenerateAddress(name string)
    GetAddressBalance(name string) uint64
    GetAddressDetails(name string)
    GetAddressList() names[]string
    SubmitTransaction(ITransaction) error
    GenerateMultisigAddress(name string, m int, n int, []string)
}

var oneSCW SCWallet



type SCWallet struct {
    var scWallet map[[simplecoin.ADDRESS_LENGTH]byte] IAuthorization
    var r *rand.Rand
}

func (w *SCWallet) WalletInit () {
    if r != nul { return }
    r = rand.New(rand.NewSource(13436523)) 
    
    for i:=0 ; i < 10 ; i++ {
        public, private, _ := ed25519.GenerateKey(w)
            
            
    }
    

func (w *SCWallet) Read(buf []byte) (int, error) {
    //if r==nil { r = rand.New(rand.NewSource(time.Now().Unix())) }
    w.WalletInit()
    for i := range buf {
        buf[i] = byte(r.Int())
    }
    return len(buf), nil
}

    
