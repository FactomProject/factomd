// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"encoding/gob"
	"math/rand"
	"net"
	"time"
)

type P2PConnection struct {
	conn           net.Conn
	SendChannel    chan Parcel // Send means "towards the network"
	ReceiveChannel chan Parcel // Recieve means "from the network"
	ConnectionID   uint64      // Random number used for loopback protection
	// and as "address" for sending messages to specific nodes.
	Online  bool         // Indicates if the connection is connected to a peer or not.
	encoder *gob.Encoder // Wire format is gobs in this version, may switch to binary
	decoder *gob.Decoder // Wire format is gobs in this version, may switch to binary
}

func (c *P2PConnection) Init() *P2PConnection {
	c.SendChannel = make(chan Parcel, 1000)
	c.ReceiveChannel = make(chan Parcel, 1000)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	c.ConnectionID = uint64(r.Int63())
	// f.ConnectionID = rand.Int(rand.Reader, math.MaxInt64)
	c.Online = false
	c.conn = nil
	return c
}

// Called when we are online.
func (c *P2PConnection) Configure(netConn net.Conn) {
	c.conn = netConn
	c.Online = true
	c.encoder = gob.NewEncoder(c.conn)
	c.decoder = gob.NewDecoder(c.conn)
	go c.processReceives()
	go c.processSends()
}

// processSendChannel gets all the messages from the network and sends them to the application
func (c *P2PConnection) processReceives() {
	for c.Online {
		var message Parcel
		err := c.decoder.Decode(&message)
		if nil != err {
			error(true, "P2PConnection.processReceives() got decoding error: %+v", err)
		} else {
			if c.isValidParcel(message) {
				// Store our connection ID so the controller can direct response to us.
				message.Header.ConnectionID = c.ConnectionID
				c.ReceiveChannel <- message
				message.PrintMessageType()
			} else {
				debug(true, "P2PConnection.processReceives() got invalid message")
				message.Print()
			}
		}
	}
}

// processReceiveChannel gets all the messages from the application and sends them out over the network
func (c *P2PConnection) processSends() {
	for c.Online {
		for parcel := range c.SendChannel {
			err := c.encoder.Encode(parcel)
			if nil != err {
				error(true, "P2PConnection.processSends() got encoding error: %+v", err)
			}
		}
	}
}

// TODO - make it easy to switch between encoding/binary and encoding/gob here.
// func (c *P2PConnection) encodeAndSend(parcel Parcel)l error {
// }

// func (c *P2PConnection) receiveAndDecode(parcel Parcel) bool {
// }

func (c *P2PConnection) dial(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		note(true, "P2PConnection.dial(%s) got error: %+v", address, err)
	} else {
		c.Configure(conn)
	}
}
func (c *P2PConnection) isValidParcel(parcel Parcel) bool {
	// TODO:  Check the network type (do we know it at this level?)
	// yeah NetworkID is a global
	// TODO:  Check the hash
	// TODO: Check all the header info, basically!
	return true
}

func (c *P2PConnection) gotBadMessage() {

	// TODO Track bad messages to ban bad peers at network level
	// Array of in P2PConnection of bad messages
	// Add this one to the array with timestamp
	// Filter all messages with timestamps over an hour (put value in protocol.go maybe an hour is too logn)
	// If count of bad messages in last hour exceeds threshold from protocol.go then we drop connection
	// Add this IP address to our banned peers (for an hour or day, also define in protocol.go)
}

func (c *P2PConnection) shutdown() {
	defer c.conn.Close()
}
