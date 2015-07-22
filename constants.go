// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

const (
	ADDRESS_LENGTH       = 32    // Length of an Address or a Hash or Public Key
	PRIVATE_LENGTH       = 64    // length of a Private Key
	SIGNATURE_LENGTH     = 64    // Length of a signature
	MAX_TRANSACTION_SIZE = 10240 // 10K like everything else?
	MINIMUM_AMOUNT       = 1     // Not sure if we need a minimum amount.  Set at 1 Factoshi

	// Database
	KEY_LIMIT            = ADDRESS_LENGTH*2  // Limit on size of keys, since Maps in Go can't
                                             // handle variable length keys.
	DB_FACTOID_BLOCKS    = "Factoid_Transaction_Blocks"
    DB_BAD_TRANS         = "Bad_Transactions_Encountered"
    DB_F_BALANCES        = "Factoid_Address_balances"
	DB_EC_BALANCES       = "Entry_Credit_Address_balances"
    
    // Wallet
    W_RCD_ADDRESS_HASH   = "wallet.address.addr"
    W_ADDRESS_PUB_KEY    = "wallet.public.key"
    W_NAME               = "wallet.address.name"
    DB_BUILD_TRANS       = "Transactions_Under_Construction"
    DB_TRANSACTIONS      = "Transactions_For_Addresses"
    
    // Block 
    MARKER                   = 0x00                    // Byte used to mark minute boundries in Factoid blocks
    TRANSACTION_PRIOR_LIMIT  = int64(12*60*60*1000)    // Transactions prior to 12hrs before a block are invalid
    TRANSACTION_POST_LIMIT   = int64(12*60*60*1000)    // Transactions after 12hrs following a block are invalid
)
// Factoid chain
var FACTOID_CHAINID	= []byte {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0x0f}
var FACTOID_CHAINID_HASH = NewHash(FACTOID_CHAINID)
// Zero Hash
var ZERO_HASH = []byte {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}
