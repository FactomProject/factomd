// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package coinbase

import (
	"fmt"

	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
)

var _ = fmt.Println

var adrs []interfaces.IAddress
var amount uint64 = 5000000000 // One Factoid (remember, fixed point math!
var addressCnt int = 0         // No coinbase payments until Milestone 3

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
func GetCoinbase(ftime uint64) interfaces.ITransaction {
	coinbase := new(factoid.Transaction)
	coinbase.SetMilliTimestamp(ftime)

	for _, adr := range adrs {
		coinbase.AddOutput(adr, amount) // add specified amount
	}

	return coinbase
}
