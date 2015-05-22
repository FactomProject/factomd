// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// This is a minimum wallet to be used to test the coin
// There isn't much in the way of interest in security 
// here, but rather provides a mechanism to create keys
// and sign transactions, etc.

package wallet

import (
    "github.com/FactomProject/simplecoin"
)

type IWalletEntry interface {
    simplecoin.IBlock
    SetRCD(simplecoin.IRCD)
    AddKey(public, private []byte)
}

type WalletEntry struct {
    rcd     simplecoin.IRCD // Verification block for this IWalletEntry
    name    string
    public  [][]byte        // Set of public keys necessary to sign the rcd
    private [][]byte        // Set of private keys necessary to sign the rcd
}

func (w WalletEntry) SetRCD(rcd simplecoin.IRCD) {
    w.rcd = rcd
}

func (w WalletEntry) AddKey(public, private []byte) {
    if len(public) != simplecoin.ADDRESS_LENGTH || 
       len(private) != simplecoin.PRIVATE_LENGTH {
        panic("Bad Keys presented to AddKey.  Should not happen.")
    }
    pu := make([]byte,simplecoin.ADDRESS_LENGTH,simplecoin.ADDRESS_LENGTH)
    pr := make([]byte,simplecoin.PRIVATE_LENGTH,simplecoin.PRIVATE_LENGTH)
    copy(pu,public)
    copy(pr,private)
    w.public = append(w.public,pu)
    w.private = append(w.private, pr)
}

func (w WalletEntry) SetName(name string) {
    w.name = name
}