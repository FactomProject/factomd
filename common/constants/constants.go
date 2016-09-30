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
	FULL_SERVER_FAULT_MSG                     // 4
	COMMIT_CHAIN_MSG                          // 5
	COMMIT_ENTRY_MSG                          // 6
	DIRECTORY_BLOCK_SIGNATURE_MSG             // 7
	EOM_TIMEOUT_MSG                           // 8
	FACTOID_TRANSACTION_MSG                   // 9
	HEARTBEAT_MSG                             // 10
	INVALID_ACK_MSG                           // 11
	INVALID_DIRECTORY_BLOCK_MSG               // 12

	REVEAL_ENTRY_MSG      // 13
	REQUEST_BLOCK_MSG     // 14
	SIGNATURE_TIMEOUT_MSG // 15
	MISSING_MSG           // 16
	MISSING_DATA          // 17
	DATA_RESPONSE         // 18
	MISSING_MSG_RESPONSE  //19

	DBSTATE_MSG          // 20
	DBSTATE_MISSING_MSG  // 21
	ADDSERVER_MSG        // 22
	CHANGESERVER_KEY_MSG // 23
	REMOVESERVER_MSG     // 24
	NEGOTIATION_MSG      // 25

	BOUNCE_MSG      // 26	test message
	BOUNCEREPLY_MSG // 26	test message
)

const (
	// Replay
	INTERNAL_REPLAY = 1
	NETWORK_REPLAY  = 2
	TIME_TEST       = 4 // Checks the time_stamp;  Don't put actual hashes into the map with this.
	REVEAL_REPLAY   = 8 // Checks for Reveal Entry Replays ... No duplicate Entries within our 4 hours!

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
	VERSION_0                = byte(0)
	FACTOMD_VERSION          = 4000000
	MAIN_NETWORK_ID   uint32 = 0xFA92E5A1
	TEST_NETWORK_ID   uint32 = 0xFA92E5A2
	LOCAL_NETWORK_ID  uint32 = 0xFA92E5A3
	CUSTOM_NETWORK_ID uint32 = 0xFA92E5A4
	MaxBlocksPerMsg          = 500

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
	TYPE_SERVER_FAULT
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

//---------------------------------------------------------------------
// Checkpoints Directory Block KeyMR
//---------------------------------------------------------------------
var CheckPoints = map[uint32]string{
	2:   "5328d4bbe7ea6efc31cf7bfc45192378454cf4e1908c56a35e6a64456a691751",
	48:  "471ef865fecf2b1a98b2e1f87434cc65a6672cc9c8fe19ab2471112431f54b36",
	100: "cde346e7ed87957edfd68c432c984f35596f29c7d23de6f279351cddecd5dc66",
	150: "d029c37c0679bfc1cdf1096237f320ae6535def5f64aeffc5105554013aa9e23",
	198: "534a4188b92155c55f9626bbf4b02721468f6237e467f9c550c03eaafc913003",
	248: "37d3801c4079012c989bef9d35cf608773855197acd63f5a253bcfebf74b87d8",
	300: "4757d31b255e789435c807cb76f5de6cd6590a39a1e5bcfa576d0290eb52dd34",
	348: "86d2159871316f4868a81586f44b742d5fbe99e2053237ff6575df67657639af",
	400: "a3c4336ff44989717664233892edfc018c579d868699d9db29f188dbae3a1f3f",
	448: "abc0048b027735c87761b3b99d69e12e1b6434ff3f2c090b23a9738d440fa9b2",
	480: "ed0da6e9879495d15425f693c2fb120192298aeaa92772fd9f963b7eb69b1bb6",
}
