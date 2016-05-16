// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"encoding/gob"
	"hash/crc32"
	"io"
	"net"
	"time"
)

type Connection struct {
	conn           net.Conn
	SendChannel    chan Parcel // Send means "towards the network"
	ReceiveChannel chan Parcel // Recieve means "from the network"
	// and as "address" for sending messages to specific nodes.
	Online          bool         // Indicates if the connection is connected to a peer or not.
	Shutdown        bool         // Indicates that this connection is broken and should be shut down.
	encoder         *gob.Encoder // Wire format is gobs in this version, may switch to binary
	decoder         *gob.Decoder // Wire format is gobs in this version, may switch to binary
	timeLastContact time.Time    // We track how recently we have heard from a peer to determin if it is still active.
	peer            Peer         // the datastructure representing the peer we are talking to. defined in peer.go
	attempts        int          // reconnection attempts
	timeLastAttempt time.Time    // time of last attempt to connect
}

func (c *Connection) Init(peer Peer) *Connection {
	c.peer = peer
	note(c.peer.Hash, "Connection.Init() called.")
	c.SendChannel = make(chan Parcel, 1000)
	c.ReceiveChannel = make(chan Parcel, 1000)
	c.Online = false
	c.conn = nil
	c.Shutdown = false
	return c
}

// Called when we are online and connected to the peer.
func (c *Connection) Configure(netConn net.Conn) {
	note(c.peer.Hash, "Connection.Configure() called. %s", c.peer.Hash)
	c.conn = netConn
	c.Online = true
	c.encoder = gob.NewEncoder(c.conn)
	c.decoder = gob.NewDecoder(c.conn)
	c.timeLastContact = time.Now()
	c.timeLastAttempt = time.Now()
	c.attempts = 0
	// Start goroutines
	go c.processSends()
	go c.processReceives()
	// Now ask the other side for the peers they know about.
	parcel := NewParcel(CurrentNetwork, []byte("Peer Request"))
	parcel.Header.Type = TypePeerRequest
	c.SendChannel <- *parcel
}

func (c *Connection) dial() {
	note(c.peer.Hash, "Connection.dial() dialing: %+v", c.peer.Address)
	conn, err := net.Dial("tcp", c.peer.Address)
	if err != nil {
		note(c.peer.Hash, "Connection.dial(%s) got error: %+v", c.peer.Address, err)
	} else {
		debug(c.peer.Hash, "Connection.dial(%s) was successful.", c.peer.Address)
		c.Configure(conn)
	}
}

// processSends gets all the messages from the application and sends them out over the network
func (c *Connection) processSends() {
	note(c.peer.Hash, "Connection.processSends() called. Online? %+v", c.Online)
	for c.Online {
		note(c.peer.Hash, "Connection.processSends() called. Items in send channel: %d Online? %b", len(c.SendChannel), c.Online)
		for parcel := range c.SendChannel {
			debug(c.peer.Hash, "processSends() sending message to network of type: %s", parcel.PrintMessageType)
			parcel.Header.NodeID = NodeID // Send it out with our ID for loopback.
			err := c.encoder.Encode(parcel)
			if nil != err {
				logerror(c.peer.Hash, c.peer.Hash, "Connection.processSends() got encoding error: %+v", err)
				c.peer.demerit()
				if io.EOF == err {
					c.connectionDropped()
				}
			}
		}
	}
	note(c.peer.Hash, "Connection.processSends() exited. %d", c.peer.Hash)
}

// processReceives gets all the messages from the network and sends them to the application
func (c *Connection) processReceives() {
	note(c.peer.Hash, "Connection.processReceives() called. %d Online? %b", c.peer.Hash, c.Online)
	for c.Online {
		var message Parcel
		err := c.decoder.Decode(&message)
		if nil != err {
			// Golang apparently doesn't provide a good way to detect various error types.
			// So, errors from "Decode" are presumed to be network type- eg closed connection.
			// So we drop our end.
			logerror(c.peer.Hash, "Connection.processReceives() got decoding error: %+v", err)
			c.peer.demerit()
			// if io.EOF == err {
			c.connectionDropped()
			// }
		} else {
			c.handleParcel(message)
		}
	}
	note(c.peer.Hash, "Connection.processReceives() exited. %s", c.peer.Address)
}

// TODO - make it easy to switch between encoding/binary and encoding/gob here.
// func (c *Connection) encodeAndSend(parcel Parcel)l error {
// }

// func (c *Connection) receiveAndDecode(parcel Parcel) bool {
// }

// handleParcel checks the parcel command type, and either generates a response, or passes it along.
func (c *Connection) handleParcel(parcel Parcel) {
	parcel.Header.Timestamp = time.Now() // set the timestamp to the recieved time.
	validity := c.parcelValidity(parcel)
	switch validity {
	case InvalidDisconnectPeer:
		debug(c.peer.Hash, "Connection.handleParcel() Disconnecting peer for incompatibility: %s", c.peer.Address)
		c.attempts = MaxNumberOfRedialAttempts + 50 // so we don't redial invalid Peer
		c.shutdown()
	case InvalidPeerDemerit:
		debug(c.peer.Hash, "Connection.handleParcel() got invalid message")
		parcel.Print()
		c.peer.demerit()
	case ParcelValid:
		c.timeLastContact = time.Now() // We only update for valid messages (incluidng pings and heartbeats)
		c.attempts = 0                 // reset since we are clearly in touch now.
		c.peer.merit()                 // Increase peer quality score.
		debug(c.peer.Hash, "Connection.handleParcel() got ParcelValid %s", parcel.MessageType())
		if Notes <= CurrentLoggingLevel {
			parcel.PrintMessageType()
		}
		c.handleParcelTypes(parcel) // handles both network commands and application messages
	}
}

// These constants support the multiple penalties and responses for Parcel validation
const (
	ParcelValid           uint8 = iota
	InvalidPeerDemerit          // The peer sent an invalid message
	InvalidDisconnectPeer       // Eg they are on the wrong network or wrong version of the software
)

func (c *Connection) parcelValidity(parcel Parcel) uint8 {
	debug(c.peer.Hash, "Connection.isValidParcel(%s)", parcel.MessageType())
	crc := crc32.Checksum(parcel.Payload, CRCKoopmanTable)
	switch {
	case parcel.Header.NodeID == NodeID: // We are talking to ourselves!
		logerror(c.peer.Hash, "Connection.isValidParcel(), failed due to loopback!: %+v", parcel.Header)
		return InvalidDisconnectPeer
	case parcel.Header.Network != CurrentNetwork:
		logerror(c.peer.Hash, "Connection.isValidParcel(), failed due to wrong network: %+v", parcel.Header)
		return InvalidDisconnectPeer
	case parcel.Header.Version < ProtocolVersionMinimum:
		logerror(c.peer.Hash, "Connection.isValidParcel(), failed due to wrong version: %+v", parcel.Header)
		return InvalidDisconnectPeer
	case parcel.Header.Length != uint32(len(parcel.Payload)):
		logerror(c.peer.Hash, "Connection.isValidParcel(), failed due to wrong length: %+v", parcel.Header)
		return InvalidPeerDemerit
	case parcel.Header.Crc32 != crc:
		logerror(c.peer.Hash, "Connection.isValidParcel(), failed due to bad checksum: %+v", parcel.Header)
		return InvalidPeerDemerit
	default:
		return ParcelValid
	}
	return ParcelValid
}
func (c *Connection) handleParcelTypes(parcel Parcel) {
	switch parcel.Header.Type {
	case TypeAlert:
		silence(c.peer.Hash, "!!!!!!!!!!!!!!!!!! Alert: TODO Alert signature checking not supported yet! BUGBUG")
	case TypePing:
		// Send Pong
		pong := NewParcel(CurrentNetwork, []byte("Pong"))
		pong.Header.Type = TypePong
		debug(c.peer.Hash, "Sending Pong.")
		c.SendChannel <- parcel
	case TypePong: // all we need is the timestamp which is set already
		return
	case TypePeerRequest:
		c.ReceiveChannel <- parcel // Controller handles these.
	case TypePeerResponse:
		c.ReceiveChannel <- parcel // Controller handles these.
	case TypeMessage:
		// Store our connection ID so the controller can direct response to us.
		parcel.Header.TargetPeer = c.peer.Hash
		parcel.Header.NodeID = NodeID
		c.ReceiveChannel <- parcel
	default:
		silence(c.peer.Hash, "!!!!!!!!!!!!!!!!!! Got message of unknown type?")
		parcel.Print()
	}
}

// We're unable to talk to the other side, but might be able to reconnect.
// So don't set c.Shutdown, which causes us to give up on the peer.
func (c *Connection) connectionDropped() {
	debug(c.peer.Hash, "Connection.connectionDropped(%+v)", "")
	// Connection dropped.
	c.Online = false
	if nil != c.conn {
		defer c.conn.Close()
	}
	c.decoder = nil
	c.encoder = nil
	c.peer.demerit()
}

// We're hanging up, or giving up on this peer, the connection is going away.
func (c *Connection) shutdown() {
	debug(c.peer.Hash, "Connection.shutdown(%+v)", "")
	c.connectionDropped()
	c.Shutdown = true
}
