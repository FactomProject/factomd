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

// The wallet interface uses bytes.  This is because we want to 
// handle fixed length values in our maps and the database.  If
// we try to use strings, then the lengths vary based on encoding
// and that complicates the implementation without really making
// the interface more usable by developers.
type ISCWallet interface {
    database.ISCDatabase
    generateKey() (public []byte,private []byte, err error)
    GenerateAddress(name []byte) simplecoin.IHash
    
    GetAddressBalance(addr simplecoin.IHash) uint64
    GetAddressDetailsAddr(addr []byte) IWalletEntry
    GetAddressList() (names [][]byte, addresses []simplecoin.Address)
    GetAddressListByName(name []byte) (names[][]byte)
    
    /** Transaction calls **/
    CreateTransaction() simplecoin.ITransaction
    AddInput(simplecoin.ITransaction, IWalletEntry, uint64)
    AddOutput(simplecoin.ITransaction, IWalletEntry, uint64)
    AddECOutput(simplecoin.ITransaction, IWalletEntry, uint64)
    
    SubmitTransaction(simplecoin.ITransaction) error
}

var oneSCW SCWallet

type SCWallet struct {
    db database.MapDB
    r *rand.Rand
}



func (w *SCWallet) GetAddressDetailsAddr(name []byte) IWalletEntry {
    return w.db.GetRaw([]byte("wallet.address.addr"),name).(IWalletEntry)
}



func (w *SCWallet) GenerateAddress(name []byte,m int, n int) (hash simplecoin.IHash) {
    
    we := new(WalletEntry)
    
    if m == 1 && n == 1 {
        pub,pri,err := w.generateKey()
        if err != nil { panic("Failed to generate a key.  Should never happen")  }
        we.AddKey(pub,pri)
        we.SetName(name)
        simplecoin.NewRCD_1(pub)
        we.SetRCD(simplecoin.NewRCD_1(pub))

        // If the name exists already, then we store this as the hash of the name.
        // If that exists, then we store it as the hash of the hash and so forth.
        // This way, we can get a list of addresses with the same name.
        //
        nm  := w.db.GetRaw([]byte("wallet.address.name"),name)
        switch {
            case nm == nil :       // New Name
                b,err := we.MarshalBinary()
                if err != nil { panic("Wallet entry failed to Marshal.  Should never happen")  }
                hash = simplecoin.Sha(b)
                w.db.PutRaw([]byte("wallet.address.hash"),hash.Bytes(),we)
                w.db.PutRaw([]byte("wallet.address.name"),name,we)
                w.db.PutRaw([]byte("wallet.address.addr"),pub,we)
            case nm != nil :       // Duplicate name.  We generate a new name, and recurse.
                nh := simplecoin.Sha(name)
                return w.GenerateAddress(nh.Bytes(),m, n)
        }
        
    } else {
        panic("Not this far yet!")
    }
    return
}

func (w *SCWallet) Init () {
    if w.r != nil { return }
    w.r = rand.New(rand.NewSource(13436523)) 
    w.db.Init()
}
    
func (w *SCWallet) Read(buf []byte) (int, error) {
    //if r==nil { r = rand.New(rand.NewSource(time.Now().Unix())) }
    w.Init()
    for i := range buf {
        buf[i] = byte(w.r.Int())
    }
    return len(buf), nil
}

func (w *SCWallet) generateKey() (public []byte,private []byte, err error){
    pub,pri,err := ed25519.GenerateKey(w)
    return pub[:], pri[:], err
}

