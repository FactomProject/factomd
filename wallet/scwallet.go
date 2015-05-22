// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// This is a minimum wallet to be used to test the coin
// There isn't much in the way of interest in security 
// here, but rather provides a mechanism to create keys
// and sign transactions, etc.

package wallet

import (
    "github.com/agl/ed25519"
    "math/rand"
    "github.com/FactomProject/simplecoin"
    "github.com/FactomProject/simplecoin/database"
)

type ISCWallet interface {
    database.ISCDatabase
    generateKey() (public []byte,private []byte, err error)
    GenerateAddress(name string)
    GetAddressBalance(name string) uint64
    GetAddressDetails(name string)
    GetAddressList() (names[]string)
    SubmitTransaction(simplecoin.ITransaction) error
}

var oneSCW SCWallet

type SCWallet struct {
    database.SCDatabase
    r *rand.Rand
}

func (w *SCWallet) WalletInit () {
    if w.r != nil { return }
    w.r = rand.New(rand.NewSource(13436523)) 
}
    
func (w *SCWallet) Read(buf []byte) (int, error) {
    //if r==nil { r = rand.New(rand.NewSource(time.Now().Unix())) }
    w.WalletInit()
    for i := range buf {
        buf[i] = byte(w.r.Int())
    }
    return len(buf), nil
}

func (w *SCWallet) generateKey() (public []byte,private []byte, err error){
    pub,pri,err := ed25519.GenerateKey(w)
    return pub[:], pri[:], err
}

