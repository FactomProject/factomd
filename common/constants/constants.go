// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package constants

import (
	"fmt"
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
	REVEAL_ENTRY_MSG                          // 13
	REQUEST_BLOCK_MSG                         // 14
	SIGNATURE_TIMEOUT_MSG                     // 15
	MISSING_MSG                               // 16
	MISSING_DATA                              // 17
	DATA_RESPONSE                             // 18
	MISSING_MSG_RESPONSE                      // 19
	DBSTATE_MSG                               // 20
	DBSTATE_MISSING_MSG                       // 21
	ADDSERVER_MSG                             // 22
	CHANGESERVER_KEY_MSG                      // 23
	REMOVESERVER_MSG                          // 24
	BOUNCE_MSG                                // 25	test message
	BOUNCEREPLY_MSG                           // 26	test message
	MISSING_ENTRY_BLOCKS                      // 27
	ENTRY_BLOCK_RESPONSE                      // 28
	INTERNALADDLEADER                         // 29
	INTERNALREMOVELEADER                      // 30
	INTERNALADDAUDIT                          // 31
	INTERNALREMOVEAUDIT                       // 32
	INTERNALTIMEOUT                           // 33
	INTERNALEOMSIG                            // 34
	INTERNALAUTHLIST                          // 35
	VOLUNTEERAUDIT                            // 36
	VOLUNTEERPROPOSAL                         // 37
	VOLUNTEERLEVELVOTE                        // 38
	FEDVOTE_MSG_BASE                          // 39
	SYNC_MSG                                  // 40

	NUM_MESSAGES // Not used, just a counter for the number of messages.
)

// Entry Credit Block entries
const (
	ECIDServerIndexNumber byte = iota // 0 Must be these values, per the specification
	ECIDMinuteNumber                  // 1
	ECIDChainCommit                   // 2
	ECIDEntryCommit                   // 3
	ECIDBalanceIncrease               // 4
)

func MessageName(Type byte) string {
	switch Type {
	case EOM_MSG:
		return "EOM"
	case ACK_MSG:
		return "Ack"
	case AUDIT_SERVER_FAULT_MSG:
		return "Audit Server Fault"
	case FED_SERVER_FAULT_MSG:
		return "Fed Server Fault"
	case FULL_SERVER_FAULT_MSG:
		return "Full Server Fault"
	case COMMIT_CHAIN_MSG:
		return "Commit Chain"
	case COMMIT_ENTRY_MSG:
		return "Commit Entry"
	case DIRECTORY_BLOCK_SIGNATURE_MSG:
		return "Directory Block Signature"
	case EOM_TIMEOUT_MSG:
		return "EOM Timeout"
	case FACTOID_TRANSACTION_MSG:
		return "Factoid Transaction"
	case HEARTBEAT_MSG:
		return "HeartBeat"
	case INVALID_ACK_MSG:
		return "Invalid Ack"
	case INVALID_DIRECTORY_BLOCK_MSG:
		return "Invalid Directory Block"
	case MISSING_MSG:
		return "Missing Msg"
	case MISSING_MSG_RESPONSE:
		return "Missing Msg Response"
	case MISSING_DATA:
		return "Missing Data"
	case DATA_RESPONSE:
		return "Data Response"
	case REVEAL_ENTRY_MSG:
		return "Reveal Entry"
	case REQUEST_BLOCK_MSG:
		return "Request Block"
	case SIGNATURE_TIMEOUT_MSG:
		return "Signature Timeout"
	case DBSTATE_MISSING_MSG:
		return "DBState Missing"
	case ADDSERVER_MSG:
		return "ADDSERVER"
	case CHANGESERVER_KEY_MSG:
		return "CHANGESERVER_KEY"
	case REMOVESERVER_MSG:
		return "REMOVESERVER"
	case DBSTATE_MSG:
		return "DBState"
	case BOUNCE_MSG:
		return "Bounce Message"
	case BOUNCEREPLY_MSG:
		return "Bounce Reply Message"
	case MISSING_ENTRY_BLOCKS: // 27
		return "MISSING_ENTRY_BLOCKS"
	case ENTRY_BLOCK_RESPONSE: // 28
		return "ENTRY_BLOCK_RESPONSE"
	case VOLUNTEERAUDIT:
		return "Volunteer Audit"
	case VOLUNTEERPROPOSAL:
		return "Volunteer Proposal"
	case VOLUNTEERLEVELVOTE:
		return "Volunteer Level Vote"
	case INTERNALADDLEADER:
		return "INTERNALADDLEADER"
	case INTERNALREMOVELEADER:
		return "INTERNALREMOVELEADER"
	case INTERNALADDAUDIT:
		return "INTERNALADDAUDIT"
	case INTERNALAUTHLIST: // 35
		return "INTERNALAUTHLIST"
	case INTERNALREMOVEAUDIT:
		return "INTERNALREMOVEAUDIT"
	case INTERNALTIMEOUT:
		return "INTERNALTIMEOUT"
	case INTERNALEOMSIG:
		return "INTERNALEOMSIG"
	case FEDVOTE_MSG_BASE:
		return "FEDVOTE_MSG_BASE"
	case SYNC_MSG:
		return "Sync Msg"

	default:
		return "Unknown:" + fmt.Sprintf(" %d", Type)
	}
}

const (
	// Limits for keeping inputs from flooding our execution
	INMSGQUEUE_HIGH = 1000
	INMSGQUEUE_MED  = 500
	INMSGQUEUE_LOW  = 100

	DBSTATE_REQUEST_LIM_HIGH = 200
	DBSTATE_REQUEST_LIM_MED  = 50

	// Replay -- Dynamic Replay filter based on messages as they are processed.
	INTERNAL_REPLAY = 1
	NETWORK_REPLAY  = 2
	TIME_TEST       = 4 // Checks the time_stamp;  Don't put actual hashes into the map with this.
	REVEAL_REPLAY   = 8 // Checks for Reveal Entry Replays ... No duplicate Entries within our 4 hours!

	// FReplay -- Block based Replay filter constructed by processing the blocks, from the database
	//            then from blocks either passed to a node, or constructed by messages.
	BLOCK_REPLAY = 16 // Ensures we don't add the same transaction to multiple blocks.
	//todo: Clay -- I changed this to not match in an experiment

	ADDRESS_LENGTH = 32 // Length of an Address or a Hash or Public Key
	// length of a Private Key
	SIGNATURE_LENGTH     = 64    // Length of a signature
	MAX_TRANSACTION_SIZE = 10240 // 10K like everything else?
	// Not sure if we need a minimum amount.  Set at 1 Factoshi

	// Database
	//==================
	// Limit on size of keys, since Maps in Go can't handle variable length keys.

	// Cross Boot Replay
	//==================
	// This is the salt filter on rebooting leaders. How long should the filter last for?
	CROSSBOOT_SALT_REPLAY_DURATION = time.Minute * 10

	// Wallet
	//==================
	// Holds the root seeds for address generation
	// Holds the latest generated seed for each root seed.

	// Block
	//==================
	MARKER                  = 0x00                       // Byte used to mark minute boundaries in Factoid blocks
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
	VERSION_0               = byte(0)
	MAIN_NETWORK_ID  uint32 = 0xFA92E5A2
	TEST_NETWORK_ID  uint32 = 0xFA92E5A3
	LOCAL_NETWORK_ID uint32 = 0xFA92E5A4
	MaxBlocksPerMsg         = 500
)

const (
	// NETWORKS:
	NETWORK_MAIN   int = iota // 0
	NETWORK_TEST              // 1
	NETWORK_LOCAL             // 2
	NETWORK_CUSTOM            // 3
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
	IDENTITY_UNASSIGNED               uint8 = iota // 0
	IDENTITY_FEDERATED_SERVER                      // 1
	IDENTITY_AUDIT_SERVER                          // 2
	IDENTITY_FULL                                  // 3
	IDENTITY_PENDING_FEDERATED_SERVER              // 4
	IDENTITY_PENDING_AUDIT_SERVER                  // 5
	IDENTITY_PENDING_FULL                          // 6
	IDENTITY_SKELETON                              // 7 - Skeleton Identity
)
