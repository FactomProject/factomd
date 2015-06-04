// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package simplecoin

const (
	ADDRESS_LENGTH       = 32    // Length of an Address or a Hash or Public Key
    PRIVATE_LENGTH       = 64    // length of a Private Key
	SIGNATURE_LENGTH     = 64    // Length of a signature
	MAX_TRANSACTION_SIZE = 10240 // 10K like everything else?
	MINIMUM_AMOUNT       = 1     // Not sure if we need a minimum amount.  Set at 1 Factoshi
	
	// Database
	DB_FACTOID_BLOCKS    = "Factoid_Transaction_Blocks"
	DB_F_BALANCES        = "Factoid_Address_balances"
    DB_EC_BALANCES       = "Entry_Credit_Address_balances"  
)

// Factoid chain
var FACTOID_CHAINID	= []byte {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0x1f}
    