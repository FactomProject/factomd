// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

/**************************
 * IRCD  Interface for Redeem Condition Datastructures (RCD)
 *
 * https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#factoid-transaction
 **************************/

type IMsg interface {
	Printable
	BinaryMarshallable

	// Returns a byte indicating the type of message.
	Type() int

	// Returns the timestamp for a message
	GetTimestamp() Timestamp

	// Hash (does not include the signature)
	GetHash() IHash

	// Return the []byte value of the message, if defined
	Bytes() []byte

	// Validate the message, given the state.  Three possible results:
	//  < 0 -- Message is invalid.  Discard
	//  0   -- Cannot tell if message is Valid
	//  1   -- Message is valid
	Validate(dbheight uint32, state IState) int

	// Returns true if this is a message for this server to execute as
	// a leader.
	Leader(IState) bool

	// Execute the leader functions of the given message
	LeaderExecute(IState) error

	// Returns true if this is a message for this server to execute as a follower
	Follower(IState) bool

	// Exeucte the follower functions of the given message
	FollowerExecute(IState) error

	// Process.  When we get a sequence of acknowledgements that we trust, we process.
	// A message will only be processed once, and in order, guaranteed.
	Process(dbheight uint32, state IState)
}
