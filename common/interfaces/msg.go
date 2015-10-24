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

	// Returns a byte indicating the type of message.
	Type() int

	// Return the int value of the message, if defined
	Int() int

	// Return the []byte value of the message, if defined
	Bytes() []byte

	// Validate the message, given the state.  Three possible results:
	//  < 0 -- Message is invalid.  Discard
	//  0   -- Cannot tell if message is Valid
	//  1   -- Message is valid
	Validate(IState) int

	// Returns true if this is a message for this server to execute as
	// a leader.
	Leader(IState) bool

	// Execute the leader functions of the given message
	LeaderExecute(IState) error

	// Returns true if this is a message for this server to execute as a follower
	Follower(IState) bool

	// Exeucte the follower functions of the given message
	FollowerExecute(IState) error
}
