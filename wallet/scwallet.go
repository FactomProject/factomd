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
    GenerateAddress(name []byte, m int, n int) (simplecoin.IHash, error)
    
    GetAddressBalance(addr simplecoin.IHash) uint64
    GetAddressDetailsAddr(addr []byte) IWalletEntry
    GetAddressList() (names [][]byte, addresses []simplecoin.Address)
    GetAddressListByName(name []byte) (names[][]byte)
    
    /** Transaction calls **/
    CreateTransaction() simplecoin.ITransaction
    AddInput(simplecoin.ITransaction, simplecoin.IHash, uint64) error
    AddOutput(simplecoin.ITransaction, simplecoin.IHash, uint64) error
    AddECOutput(simplecoin.ITransaction, simplecoin.IHash, uint64) error
    ValidateTransaction(simplecoin.ITransaction) (bool,error)
    
    SubmitTransaction(simplecoin.ITransaction) error
}

var factoshisPerEC uint64 = 0x100000

var oneSCW SCWallet

type SCWallet struct {
    ISCWallet
    db database.MapDB
    r *rand.Rand
}

var _ ISCWallet = (*SCWallet)(nil)

func (w *SCWallet) GetAddressDetailsAddr(name []byte) IWalletEntry {
    return w.db.GetRaw([]byte("wallet.address.addr"),name).(IWalletEntry)
}

func (w *SCWallet) GenerateAddress(name []byte,m int, n int) (hash simplecoin.IHash, err error) {
    
    we := new(WalletEntry)
    
    nm := w.db.GetRaw([]byte("wallet.address.name"),name)
    if nm != nil {
        return nil, fmt.Errorf("Duplicate Name")
    }
    
    if m == 1 && n == 1 {
        pub,pri,err := w.generateKey()
        if err != nil { return nil, err  }
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
                if err != nil { return nil, err }
                hash = simplecoin.Sha(b)
                w.db.PutRaw([]byte("wallet.address.hash"),hash.Bytes(),we)
                w.db.PutRaw([]byte("wallet.address.name"),name,we)
                w.db.PutRaw([]byte("wallet.address.addr"),pub,we)
            case nm != nil :       // Duplicate name.  We generate a new name, and recurse.
                return nil, fmt.Errorf("Should never get here!  This is disabled!")
                nh := simplecoin.Sha(name)
                return w.GenerateAddress(nh.Bytes(),m, n)
            default :
                return nil, fmt.Errorf("Should never get here!  This isn't possible!")
        }
        
    } else {
        return nil, fmt.Errorf("Not this far yet!")
    }
    return
}

func (w *SCWallet) Init (a ...interface{}) {
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

func (w *SCWallet)  CreateTransaction() simplecoin.ITransaction {
    return new(simplecoin.Transaction)
}

func (w *SCWallet) AddInput(trans simplecoin.ITransaction, hash simplecoin.IHash, amount uint64) error {
     we := w.db.GetRaw([]byte("wallet.address.hash"),hash.Bytes()).(*WalletEntry)
     if we == nil { 
         return fmt.Errorf("Unknown address")
     }
     adr, err := we.GetAddress()
     trans.AddInput(simplecoin.CreateAddress(adr),amount)
     if err != nil {
         return err
     }
     return nil
}     
     
func (w *SCWallet) AddOutput(trans simplecoin.ITransaction, hash simplecoin.IHash, amount uint64) error {
    we := w.db.GetRaw([]byte("wallet.address.hash"),hash.Bytes()).(*WalletEntry)
    if we == nil { 
        return fmt.Errorf("Unknown address")
    }
    adr, err := we.GetAddress()
    trans.AddOutput(simplecoin.CreateAddress(adr),amount)
    if err != nil {
        return err
    }
    return nil
}     
 
func (w *SCWallet) AddECOutput(trans simplecoin.ITransaction, hash simplecoin.IHash, amount uint64) error {
    we := w.db.GetRaw([]byte("wallet.address.hash"),hash.Bytes()).(*WalletEntry)
    if we == nil { 
        return fmt.Errorf("Unknown address")
    }
    adr, err := we.GetAddress()
    trans.AddECOutput(simplecoin.CreateAddress(adr),amount)
    if err != nil {
        return err
    }
    return nil
}

func (w *SCWallet) ValidateTransaction(simplecoin.ITransaction) (bool,error){
    
    return true,nil
}    
    
