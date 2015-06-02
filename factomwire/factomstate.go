// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// This package doesn't pretend to be Factom.  What it does
// is provide data so I can validate Transaction Processing.

package factomwire

import (
    "bytes"
    "encoding/binary"
    sc "github.com/FactomProject/simplecoin"
    db "github.com/FactomProject/simplecoin/database"
)

type IFactomState interface {
    // Set the database for the Coin State.  This is where
    // we manage the balances for transactions.  We also look
    // for previous blocks here.
    SetDB(db.ISCDatabase)       
    // Update balance updates the balance for an address in
    // the database.  Note that we take an int64 to allow debits
    // as well as credits
    UpdateBalance(address sc.IAddress, amount int64)  error
    // Return the balance for an address
    GetBalance(address sc.IAddress) uint64
    // Return the transaction with this hash
}

type FactomState struct {
    database db.ISCDatabase
}

func(fs *FactomState) SetDB(database db.ISCDatabase){
    fs.database = database
}

// Any address that is not defined has a zero balance.
func(fs *FactomState) GetBalance(address sc.IAddress) uint64 {
    balance := uint64(0)
    b  := fs.database.GetRaw([]byte("factomAddress.balances"),address.Bytes())
    if b != nil  {
        balance, data := binary.BigEndian.Uint64(b)
    }
    return balance
}

// Update balance throws an error if your update will drive the balance negative.
func(fs *FactomState) UpdateBalance(address sc.IAddress, amount int64) error {
    nbalance += int64(fs.GetBlance(address))+amount
    if nbalance < 0 {return fmt.Errorf("New balance cannot be negative")}
    balance = uint64(nbalance)
    var out bytes.Buffer
    binary.Write(&out, binary.BigEndian, uint64(balance))
    fs.database.PutRaw([]byte("factomAddress.balances"),address.Bytes(),out.Bytes())
}    
    
    
