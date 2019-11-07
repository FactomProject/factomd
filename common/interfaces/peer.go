// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

// The Peer interface allows Factom to connect to any implementation of a p2p network.
// The simulator uses an implementation of IPeer to simulate various networks
type IPeer interface {
	Initialize(nameTo, nameFrom string) IPeer // Name of peer
	GetNameTo() string                        // Return the name of the peer
	GetNameFrom() string                      // Return the name of the peer
	Send(IMsg) error                          // Send a message to this peer
	Receive() (IMsg, error)                   // Read a message from this peer; nil if no message is ready.
	Len() int                                 // Returns the number of messages waiting to be read
	Equals(IPeer) bool                        // Is this connection equal to parm connection
	Weight() int                              // How many nodes does this peer represent?
	BytesOut() int                            // Bytes sent out per second from this peer
	BytesIn() int                             // Bytes received per second from this peer
}
