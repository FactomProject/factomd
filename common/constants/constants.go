// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package constants

import (
	"time"
)

// Messages
const (
	EOM_MSG = iota
	ACK_MSG
	AUDIT_SERVER_FAULT_MSG
	COMMIT_CHAIN_MSG
	COMMIT_ENTRY_MSG
	DIRECTORY_BLOCK_SIGNATURE_MSG
	EOM_TIMEOUT_MSG
	FACTOID_TRANSACTION_MSG
	HEARTBEAT_MSG
	INVALID_ACK_MSG
	INVALID_DIRECTORY_BLOCK_MSG
	MISSING_ACK_MSG
	PROMOTION_DEMOTION_MSG
	REVEAL_ENTRY_MSG
	REQUEST_BLOCK_MSG
	SIGNATURE_TIMEOUT_MSG
)

const (
	ADDRESS_LENGTH       = 32    // Length of an Address or a Hash or Public Key
	PRIVATE_LENGTH       = 64    // length of a Private Key
	SIGNATURE_LENGTH     = 64    // Length of a signature
	MAX_TRANSACTION_SIZE = 10240 // 10K like everything else?
	MINIMUM_AMOUNT       = 1     // Not sure if we need a minimum amount.  Set at 1 Factoshi

	// Database
	//==================
	// Limit on size of keys, since Maps in Go can't handle variable length keys.
	KEY_LIMIT               = ADDRESS_LENGTH * 2
	DB_DIRECTORY_BLOCKS     = "Factom_Directory_Blocks"
	DB_FACTOID_BLOCKS       = "Factoid_Transaction_Blocks"
	DB_ADMIN_BLOCKS         = "Factom_Admin_Blocks"
	DB_ENTRY_CREDIT_BLOCKS  = "Factom_Entry_Credit_Blocks"
	DB_ENTRY_CHAIN_BLOCKS   = "Factom_Entry_Chain_Blocks"
	DB_ENTRIES              = "Factom_Entries"
	DB_DB_FORWARD           = "Directory_Block_Forward_Hashes"
	DB_FACTOID_FORWARD      = "Factoid_Block_Forward_Hashes"
	DB_ENTRY_CREDIT_FORWARD = "Entry_Credit_Forward_Hashes"
	DB_ENTRY_CHAIN_FORWARD  = "Entry_Chain_Forward_Hashes"

	// Wallet
	//==================
	W_SEEDS            = "wallet.address.seeds"      // Holds the root seeds for address generation
	W_SEED_HEADS       = "wallet.address.seed.heads" // Holds the latest generated seed for each root seed.
	W_RCD_ADDRESS_HASH = "wallet.address.addr"
	W_ADDRESS_PUB_KEY  = "wallet.public.key"
	W_NAME             = "wallet.address.name"
	DB_BUILD_TRANS     = "Transactions_Under_Construction"
	DB_TRANSACTIONS    = "Transactions_For_Addresses"

	// Block
	//==================
	MARKER                  = 0x00                       // Byte used to mark minute boundries in Factoid blocks
	TRANSACTION_PRIOR_LIMIT = int64(12 * 60 * 60 * 1000) // Transactions prior to 12hrs before a block are invalid
	TRANSACTION_POST_LIMIT  = int64(12 * 60 * 60 * 1000) // Transactions after 12hrs following a block are invalid

	//Entry Credit Blocks (For now, everyone gets the same cap)
	EC_CAP = 5      //Number of ECBlocks we start with.
	AB_CAP = EC_CAP //Administrative Block Cap for AB messages

	//Limits and Sizes
	//==================
	MAX_ENTRY_SIZE    = uint16(10240) //Maximum size for Entry External IDs and the Data
	HASH_LENGTH       = int(32)       //Length of a Hash
	SIG_LENGTH        = int(64)       //Length of a signature
	MAX_ORPHAN_SIZE   = int(5000)     //Prphan mem pool size
	MAX_TX_POOL_SIZE  = int(50000)    //Transaction mem pool size
	MAX_BLK_POOL_SIZE = int(500000)   //Block mem bool size
	MAX_PLIST_SIZE    = int(150000)   //MY Process List size

	MAX_ENTRY_CREDITS = uint8(10) //Max number of entry credits per entry
	MAX_CHAIN_CREDITS = uint8(20) //Max number of entry credits per chain

	COMMIT_TIME_WINDOW = time.Duration(12) //Time windows for commit chain and commit entry +/- 12 hours

	//NETWORK constants
	//==================
	VERSION_0     = byte(0)
	NETWORK_ID_DB = uint32(4203931041) //0xFA92E5A1
	NETWORK_ID_EB = uint32(4203931042) //0xFA92E5A2
	NETWORK_ID_CB = uint32(4203931043) //0xFA92E5A3

	// NETWORKS:

	NETWORK_MAIN   int = iota // 0
	NETWORK_TEST              // 1
	NETWORK_LOCAL             // 2
	NETWORK_CUSTOM            // 3

	//For Factom TestNet
	//==================
	NETWORK_ID_TEST = uint32(0) //0x0

	// Server Info
	//==================
	//Server running mode
	FULL_NODE   = "FULL"
	SERVER_NODE = "SERVER"
	LIGHT_NODE  = "LIGHT"
	
	CLIENT_MODE int = iota // 0
	SERVER_MODE 		   // 1 	
	AUDIT_SERVER_MODE	   // 2
	
	//Server public key for milestone 1
	SERVER_PUB_KEY = "0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a"
	//Genesis directory block timestamp in RFC3339 format
	GENESIS_BLK_TIMESTAMP = "2015-09-01T20:00:00+00:00"
	//Genesis directory block hash
	GENESIS_DIR_BLOCK_HASH = "cbd3d09db6defdc25dfc7d57f3479b339a077183cd67022e6d1ef6c041522b40"
)

// Slices and arrays that should not ever be modified:
//===================================================
// Used as a key in the wallet to find the current seed value.
var CURRENT_SEED = [32]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}

// Entry Credit Chain
var EC_CHAINID = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x0c}

// Directory Chain
var D_CHAINID = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x0d}

// Directory Chain
var ADMIN_CHAINID = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x0a}

// Factoid chain
var FACTOID_CHAINID = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x0f}

// Zero Hash
var ZERO_HASH = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var ZERO = []byte{0}

//---------------------------------------------------------------
// Types of entries (transactions) for Admin Block
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#adminid-bytes
//---------------------------------------------------------------
const (
	TYPE_MINUTE_NUM         uint8 = iota // 0
	TYPE_DB_SIGNATURE                    // 1
	TYPE_REVEAL_MATRYOSHKA               // 2
	TYPE_ADD_MATRYOSHKA                  // 3
	TYPE_ADD_SERVER_COUNT                // 4
	TYPE_ADD_FED_SERVER                  // 5
	TYPE_REMOVE_FED_SERVER               // 6
	TYPE_ADD_FED_SERVER_KEY              // 7
	TYPE_ADD_BTC_ANCHOR_KEY              //8
)
