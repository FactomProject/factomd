// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
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
}

var oneSCW SCWallet

type scwEntry struct {
    authBlk IAuthorization

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

    
