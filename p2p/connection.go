// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"encoding/gob"
	"hash/crc32"
	"io"
	// "math/rand"
	"net"
	"time"
)

type Connection struct {
	conn           net.Conn
	SendChannel    chan Parcel // Send means "towards the network"
	ReceiveChannel chan Parcel // Recieve means "from the network"
	ConnectionID   string      // Random number used for loopback protection
	// and as "address" for sending messages to specific nodes.
	Online          bool         // Indicates if the connection is connected to a peer or not.
	encoder         *gob.Encoder // Wire format is gobs in this version, may switch to binary
	decoder         *gob.Decoder // Wire format is gobs in this version, may switch to binary
	timeLastContact time.Time    // We track how recently we have heard from a peer to determin if it is still active.
	peer            Peer         // the datastructure representing the peer we are talking to. defined in peer.go
	attempts        int          // reconnection attempts
	timeLastAttempt time.Time    // time of last attempt to connect
}

func (c *Connection) Init(peer Peer) *Connection {
	note(true, "Connection.Init() called.")
	c.peer = peer
	c.SendChannel = make(chan Parcel, 1000)
	c.ReceiveChannel = make(chan Parcel, 1000)
	// r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// c.ConnectionID = uint64(r.Int63())
	// c.ConnectionID = c.peer.Hash  // I think this is redundant, just use the peer hash.
	c.Online = false
	c.conn = nil
	return c
}

// Called when we are online and connected to the peer.
func (c *Connection) Configure(netConn net.Conn) {
	note(true, "Connection.Configure() called. %d", c.ConnectionID)
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
}

// processSends gets all the messages from the application and sends them out over the network
func (c *Connection) processSends() {
	note(true, "Connection.processSends() called. %d Online? %b", c.ConnectionID, c.Online)
	for c.Online {
		note(true, "Connection.processSends() called. Items in send channel: %d Online? %b", len(c.SendChannel), c.Online)
		for parcel := range c.SendChannel {
			parcel.Header.ConnectionID = c.ConnectionID // Send it out with our ID for loopback.
			debug(true, "Connection.processSends() Calling Encoder")
			err := c.encoder.Encode(parcel)
			debug(true, "Connection.processSends() BACK from Calling Encoder")
			if nil != err {
				logerror(true, "Connection.processSends() got encoding error: %+v", err)
				c.peer.demerit()
				if io.EOF == err {
					c.connectionDropped()
				}
			}
		}
	}
	note(true, "Connection.processSends() exited. %d", c.ConnectionID)
}

func (c *Connection) connectionDropped() {
	// Connection dropped.
	c.Online = false
	defer c.conn.Close()
	c.decoder = nil
	c.encoder = nil
	c.peer.demerit()
}


// processReceives gets all the messages from the network and sends them to the application
func (c *Connection) processReceives() {
	note(true, "Connection.processReceives() called. %d Online? %b", c.ConnectionID, c.Online)
	for c.Online {
		var message Parcel
		debug(true, "Connection.processReceives() Calling Decoder")
		err := c.decoder.Decode(&message)
		debug(true, "Connection.processReceives() Out from Decoder")
		if nil != err {
			logerror(true, "Connection.processReceives() got decoding error: %+v", err)
			c.peer.demerit()
			if io.EOF == err {
				c.connectionDropped()
			}
		} else {
			c.handleParcel(message)
		}
	}
	note(true, "Connection.processReceives() exited. %s", c.peer.address)
}


// TODO - make it easy to switch between encoding/binary and encoding/gob here.
// func (c *Connection) encodeAndSend(parcel Parcel)l error {
// }

// func (c *Connection) receiveAndDecode(parcel Parcel) bool {
// }

func (c *Connection) dial() {
	conn, err := net.Dial("tcp", c.peer.address)
	if err != nil {
		c.timeLastAttempt = time.Now()
		note(true, "Connection.dial(%s) got error: %+v", c.peer.address, err)
	} else {
		debug(true, "Connection.dial(%s) was successful.", c.peer.address)
		c.Configure(conn)
	}
}

// handleParcel checks the parcel command type, and either generates a response, or passes it along.
func (c *Connection) handleParcel(parcel Parcel) {
	parcel.Header.Timestamp = time.Now() // set the timestamp to the recieved time.

	switch validity := c.parcelValidity(parcel) {
	case InvalidDisconnectPeer:
				debug(true, "Connection.handleParcel() Disconnecting peer for incompatibility: %s", c.peer.address)
				c.attempts = MaxNumberOfRedialAttempts + 50 // so we don't redial invalid Peer
				c.shutdown()
	case InvalidPeerDemerit:
		debug(true, "Connection.processReceives() got invalid message")
		message.Print()
				c.peer.demerit()
	case ParcelValid:
		c.timeLastContact = time.Now() // We only update for valid messages (incluidng pings and heartbeats)
		c.peer.merit() // Increase peer quality score.
		message.PrintMessageType()
		c.handleParcelTypes(parcel) // handles both network commands and application messages
	}
}

// These constants support the multiple penalties and responses for Parcel validation
const (
	ParcelValid    uint8 = iota
	InvalidPeerDemerit          // The peer sent an invalid message
	InvalidDisconnectPeer       // Eg they are on the wrong network or wrong version of the software
)

func (c *Connection) parcelValidity(parcel Parcel) uint8 {
	debug(true, "Connection.isValidParcel(%+v)", parcel)
	crc := crc32.Checksum(parcel.Payload, CRCKoopmanTable)
	switch {
case parcel.Header.ConnectionID == c.ConnectionID: // We are talking to ourselves!
		logerror(true, "Connection.isValidParcel(), failed due to loopback!: %+v", parcel)
		return InvalidDisconnectPeer
	case parcel.Header.Network != CurrentNetwork:
		logerror(true, "Connection.isValidParcel(), failed due to wrong network: %+v", parcel)
		return InvalidDisconnectPeer
	case parcel.Header.Version < ProtocolVersionMinimum:
		logerror(true, "Connection.isValidParcel(), failed due to wrong version: %+v", parcel)
		return InvalidDisconnectPeer
	case parcel.Header.Length != uint32(len(parcel.Payload)):
		logerror(true, "Connection.isValidParcel(), failed due to wrong length: %+v", parcel)
		return InvalidPeerDemerit
	case parcel.Header.Crc32 != crc:
		logerror(true, "Connection.isValidParcel(), failed due to bad checksum: %+v", parcel)
		return InvalidPeerDemerit
	}
	return ParcelValid
}
func (c *Connection) handleParcelTypes(parcel Parcel) {
	switch parcel.Header.Type {
	case TypeAlert:
		silence("!!!!!!!!!!!!!!!!!! Alert: TODO Alert signature checking not supported yet! BUGBUG")
	case TypeHeartbeat:
		return // NO-OP, we have already updated lastContact.
	case TypePing:
		// Send Pong
	pong := NewParcel(CurrentNetwork, []byte("Pong"))
		pong.Header.Type = TypePong
		c.SendChannel <- parcel
	case TypePong:
		return 
	case TypePeerRequest:
		c.ReceiveChannel <- parcel  // Controller handles these.
	case TypePeerResponse:
		c.ReceiveChannel <- parcel  // Controller handles these.
	case TypeMessage:
		// Store our connection ID so the controller can direct response to us.
		parcel.Header.ConnectionID = c.ConnectionID
		c.ReceiveChannel <- parcel
	default:
		silence("!!!!!!!!!!!!!!!!!! Got message of unknown type?")
		parcel.Print()
	}
}

// func (c *Connection) gotBadMessage() {
// 	debug(true, "Connection.gotBadMessage()")
// 	// TODO Track bad messages to ban bad peers at network level
// 	// Array of in Connection of bad messages
// 	// Add this one to the array with timestamp
// 	// Filter all messages with timestamps over an hour (put value in protocol.go maybe an hour is too logn)
// 	// If count of bad messages in last hour exceeds threshold from protocol.go then we drop connection
// 	// Add this IP address to our banned peers (for an hour or day, also define in protocol.go)
// }

func (c *Connection) shutdown() {
	debug(true, "Connection.shutdown(%+v)", "")
	c.Online = false
	if nil != c.conn {
		defer c.conn.Close()

	}
}
