// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package constants

import (
	"time"
)

// Messages
const (
	EOM_MSG                       byte = iota // 0
	ACK_MSG                                   // 1
	FED_SERVER_FAULT_MSG                      // 2
	AUDIT_SERVER_FAULT_MSG                    // 3
	COMMIT_CHAIN_MSG                          // 4
	COMMIT_ENTRY_MSG                          // 5
	DIRECTORY_BLOCK_SIGNATURE_MSG             // 6
	EOM_TIMEOUT_MSG                           // 7
	FACTOID_TRANSACTION_MSG                   // 8
	HEARTBEAT_MSG                             // 9
	INVALID_ACK_MSG                           // 10
	INVALID_DIRECTORY_BLOCK_MSG               // 11

	REVEAL_ENTRY_MSG      // 12
	REQUEST_BLOCK_MSG     // 13
	SIGNATURE_TIMEOUT_MSG // 14
	MISSING_MSG           // 15
	MISSING_DATA          // 16
	DATA_RESPONSE         // 17
	MISSING_MSG_RESPONSE  //18

	DBSTATE_MSG          // 19
	DBSTATE_MISSING_MSG  // 20
	ADDSERVER_MSG        // 21
	CHANGESERVER_KEY_MSG // 22
	REMOVESERVER_MSG     // 23
)

const (
	// Replay
	INTERNAL_REPLAY = 1
	NETWORK_REPLAY  = 2

	ADDRESS_LENGTH = 32 // Length of an Address or a Hash or Public Key
	// length of a Private Key
	SIGNATURE_LENGTH     = 64    // Length of a signature
	MAX_TRANSACTION_SIZE = 10240 // 10K like everything else?
	// Not sure if we need a minimum amount.  Set at 1 Factoshi

	// Database
	//==================
	// Limit on size of keys, since Maps in Go can't handle variable length keys.

	// Wallet
	//==================
	// Holds the root seeds for address generation
	// Holds the latest generated seed for each root seed.

	// Block
	//==================
	MARKER                  = 0x00                       // Byte used to mark minute boundries in Factoid blocks
	TRANSACTION_PRIOR_LIMIT = int64(12 * 60 * 60 * 1000) // Transactions prior to 12hrs before a block are invalid
	TRANSACTION_POST_LIMIT  = int64(12 * 60 * 60 * 1000) // Transactions after 12hrs following a block are invalid

	//Entry Credit Blocks (For now, everyone gets the same cap)
	EC_CAP = 5 //Number of ECBlocks we start with.
	//Administrative Block Cap for AB messages

	//Limits and Sizes
	//==================
	//Maximum size for Entry External IDs and the Data
	HASH_LENGTH = int(32) //Length of a Hash
	//Length of a signature
	//Prphan mem pool size
	//Transaction mem pool size
	//Block mem bool size
	//MY Process List size

	//Max number of entry credits per entry
	//Max number of entry credits per chain

	COMMIT_TIME_WINDOW = time.Duration(12) //Time windows for commit chain and commit entry +/- 12 hours

	//NETWORK constants
	//==================
	VERSION_0       = byte(0)
	FACTOMD_VERSION = 4000000
	//0xFA92E5A1
	//0xFA92E5A2
	//0xFA92E5A3
	MaxBlocksPerMsg = 500

	// NETWORKS:
)
const (
	NETWORK_MAIN   int = iota // 0
	NETWORK_TEST              // 1
	NETWORK_LOCAL             // 2
	NETWORK_CUSTOM            // 3

	//For Factom TestNet
	//==================
	//0x0

	// Server Info
	//==================
	//Server running mode

)

const (
	// 0
	// 1
	// 2

	//Server public key for milestone 1
	SERVER_PUB_KEY = "0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a"
	//Genesis directory block timestamp in RFC3339 format

	//Genesis directory block hash

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
	TYPE_ADD_AUDIT_SERVER                // 6
	TYPE_REMOVE_FED_SERVER               // 7
	TYPE_ADD_FED_SERVER_KEY              // 8
	TYPE_ADD_BTC_ANCHOR_KEY              // 9
)

//---------------------------------------------------------------------
// Identity Status Types
//---------------------------------------------------------------------
const (
	IDENTITY_UNASSIGNED               int = iota // 0
	IDENTITY_FEDERATED_SERVER                    // 1
	IDENTITY_AUDIT_SERVER                        // 2
	IDENTITY_FULL                                // 3
	IDENTITY_PENDING_FEDERATED_SERVER            // 4
	IDENTITY_PENDING_AUDIT_SERVER                // 5
	IDENTITY_PENDING_FULL                        // 6
	IDENTITY_PENDING                             // 7
)
