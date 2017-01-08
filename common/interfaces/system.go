// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

/**************************
 * IRCD  Interface for Redeem Condition Datastructures (RCD)
 *
 * https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#factoid-transaction
 **************************/

type ISystem interface {
	Printable
	BinaryMarshallable

	// Returns the timestamp for a message
	GetTimestamp() Timestamp

	// Send this message out over the NetworkOutQueue.  This is done with a method
	// to allow easier debugging and simulation.
	SendOut(IState, IMsg)

	Process(dbheight uint32, state IState) bool
}
