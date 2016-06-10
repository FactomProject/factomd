// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

/**************************
 * IRCD  Interface for Redeem Condition Datastructures (RCD)
 *
 * https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#factoid-transaction
 **************************/

type IMsg interface {
	Printable
	BinaryMarshallable

	// Returns a byte indicating the type of message.
	Type() byte

	// A local message is never broadcast to the greater network.
	IsLocal() bool
	SetLocal(bool)

	// Returns the origin of this message; used to track
	// where a message came from. If int == -1, then this
	// FactomNode generated the message.
	GetOrigin() int
	SetOrigin(int)

	GetNetworkOrigin() string
	SetNetworkOrigin(string)

	// Returns the timestamp for a message
	GetTimestamp() Timestamp

	// Hash for this message as used by Consensus (i.e. what we match)
	GetHash() IHash

	// Hash of this message.  Each message must be unique (excludes signatures)
	GetMsgHash() IHash

	// Returns the full message hash of a message (includes signatures)
	GetFullMsgHash() IHash

	// If this message should only reply to a peer, this is true.  If to
	// be broadcast, this should be false.  If the Origin is 0, then the
	// network can pick a peer to try.  If Origin is > 0, then the message
	// must go back to that peer (this message is a reply).
	IsPeer2Peer() bool
	SetPeer2Peer(bool)

	// Validate the message, given the state.  Three possible results:
	//  < 0 -- Message is invalid.  Discard
	//  0   -- Cannot tell if message is Valid
	//  1   -- Message is valid
	Validate(IState) int

	//Set the VMIndex for a message
	ComputeVMIndex(IState)

	// Call here if the server is a leader
	LeaderExecute(IState)

	// Debugging thing to track the leader responsible for a message ack.
	GetLeaderChainID() IHash
	SetLeaderChainID(IHash)

	// Call here if the server is a follower
	FollowerExecute(IState)

	// Process.  When we get a sequence of acknowledgements that we trust, we process.
	// A message will only be processed once, and in order, guaranteed.
	// Returns true if able to process, false if process is waiting on something.
	Process(dbheight uint32, state IState) bool

	// Some Messages need to be processed on certain VMs.  We set this and querry
	// the indexes of these machines here.
	GetVMIndex() int
	SetVMIndex(int)
	GetVMHash() []byte
	SetVMHash([]byte)
	GetMinute() byte
	SetMinute(byte)

	// Stall handling
	IsStalled() bool
	SetStall(bool)
}
