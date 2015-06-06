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
    
    // Wallet
    W_ADDRESS_HASH       = "wallet.address.addr"
    W_ADDRESS_PUB_KEY    = "wallet.public.key"
    W_NAME_HASH          = "wallet.address.name"
    
    // Validation
    WELL_FORMED          = "Pass"
    BALANCE_FAIL         = "The sum of the inputs do not cover the sum of the outputs"
    MIN_INPUT_FAIL       = "Must have at least one input"
    MIN_SPEND_FAIL       = "Must have a minimum spend in the transaction"
    RCD_INPUT_FAIL       = "Must have one and only one RCD per input"
    RCD_REPORT_FAIL      = "RCD failed to provide an address"
    RCD_MATCH_FAIL       = "The inputs and the RCDs must match, and be in the same order"
    DUP_INPUT_FAIL       = "No duplicate inputs allowed"

)
// Factoid chain
var FACTOID_CHAINID	= []byte {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0x1f}
var FACTOID_CHAINID_HASH = NewHash(FACTOID_CHAINID)
// Zero Hash
var ZERO_HASH = []byte {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}
