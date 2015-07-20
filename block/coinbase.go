// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package block

import (
    fct "github.com/FactomProject/factoid"
    "github.com/FactomProject/factoid/wallet"
    "fmt"
)

var _ = fct.Prt
var _ = fmt.Println

var adrs []fct.IAddress 
var amount     uint64 = 25000000           // One Factoid (remember, fixed point math!
var addressCnt int    = 10                  // 10 addresses for now.


// Allows the amount paid in the coinbase to be modified.   This is
// NOT allowed in production!  That's why it is here in Test!
func UpdateAmount(amt uint64) {
    amount = amt
}

// This routine generates the Coinbase.  This is a fixed amount to be
// paid to the federated servers.  
//
// Currently we are paying just a few fixed addresses.
//
func GetCoinbase(ftime uint64) fct.ITransaction {
    
    if adrs == nil {
        var w wallet.ISCWallet 
        w = new (wallet.SCWallet)
        w.Init()
        
        adrs = make([]fct.IAddress,addressCnt)
        for i:=0;i<addressCnt;i++ {
            adr,_ := w.GenerateFctAddress([]byte("adr"+string(i)),1,1)
            adrs[i] = adr
        }
    }
    
    coinbase := new(fct.Transaction)
    coinbase.SetMilliTimestamp(ftime)
        
    for _,adr := range adrs {
        coinbase.AddOutput(adr,amount)   // add specified amount
    }
        
    return coinbase
}