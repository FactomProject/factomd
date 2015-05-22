// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// This is a minimum wallet to be used to test the coin
// There isn't much in the way of interest in security 
// here, but rather provides a mechanism to create keys
// and sign transactions, etc.

package wallet

import (
)

type IWalletEntry interface {
    GetAddressDetails(name string)
    GetAddressList() names[]string
    GetAddressBalance(name string) uint64
    SubmitTransaction(ITransaction) error
    GenerateAddress(name string)
    GenerateMultisigAddress(name string, m int, n int, []string)
}