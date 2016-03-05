// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

// The Peer interface allows Factom to connect to any implementation of a p2p network.
// The simulator uses an implementation of IPeer to simulate various networks
type IPeer interface {
	Init(name string) IPeer	// Name of peer
	GetName() string		// Return the name of the peer
	Send(IMsg) (error)		// Send a message to this peer
	Recieve() (IMsg, error)	// Recieve a message from this peer; nil if no message is ready.
	Len() int				// Returns the number of messages waiting to be read
}

