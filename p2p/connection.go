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

type Connection struct {
	conn           net.Conn
	SendChannel    chan Parcel // Send means "towards the network"
	ReceiveChannel chan Parcel // Recieve means "from the network"
	ConnectionID   uint64      // Random number used for loopback protection
	// and as "address" for sending messages to specific nodes.
	Online  bool         // Indicates if the connection is connected to a peer or not.
	encoder *gob.Encoder // Wire format is gobs in this version, may switch to binary
	decoder *gob.Decoder // Wire format is gobs in this version, may switch to binary
}

func (c *Connection) Init() *Connection {
	note(true, "Connection.Init() called.")
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
func (c *Connection) Configure(netConn net.Conn) {
	note(true, "Connection.Configure() called. %d", c.ConnectionID)
	c.conn = netConn
	c.Online = true
	c.encoder = gob.NewEncoder(c.conn)
	c.decoder = gob.NewDecoder(c.conn)
	go c.processReceives()
	go c.processSends()
	// go c.heartbeat()
}

// func (c *Connection) heartbeat() {
// 	for {
// 		note(true, "Connection.heartbeat() called. %d Online? %+v", c.ConnectionID, c.Online)
// 		time.Sleep(time.Second * 1)
// 	}

// }

// processSendChannel gets all the messages from the network and sends them to the application
func (c *Connection) processReceives() {
	note(true, "Connection.processReceives() called. %d Online? %b", c.ConnectionID, c.Online)
	for c.Online {
		var message Parcel
		debug(true, "Connection.processReceives() Calling Decoder")
		err := c.decoder.Decode(&message)
		debug(true, "Connection.processReceives() Out from Decoder")
		if nil != err {
			logerror(true, "Connection.processReceives() got decoding error: %+v", err)
		} else {
			if c.isValidParcel(message) {
				// Store our connection ID so the controller can direct response to us.
				message.Header.ConnectionID = c.ConnectionID
				c.ReceiveChannel <- message
				message.PrintMessageType()
			} else {
				debug(true, "Connection.processReceives() got invalid message")
				message.Print()
			}
		}
	}
	note(true, "Connection.processReceives() exited. %d", c.ConnectionID)

}

// processReceiveChannel gets all the messages from the application and sends them out over the network
func (c *Connection) processSends() {
	note(true, "Connection.processSends() called. %d Online? %b", c.ConnectionID, c.Online)
	for c.Online {
		note(true, "Connection.processSends() called. Items in send channel: %d Online? %b", len(c.SendChannel), c.Online)

		for parcel := range c.SendChannel {
			debug(true, "Connection.processSends() Calling Encoder")
			err := c.encoder.Encode(parcel)
			debug(true, "Connection.processSends() BACK from Calling Encoder")
			if nil != err {
				logerror(true, "Connection.processSends() got encoding error: %+v", err)
			}
		}
	}
	note(true, "Connection.processSends() exited. %d", c.ConnectionID)

}

// TODO - make it easy to switch between encoding/binary and encoding/gob here.
// func (c *Connection) encodeAndSend(parcel Parcel)l error {
// }

// func (c *Connection) receiveAndDecode(parcel Parcel) bool {
// }

func (c *Connection) dial(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		note(true, "Connection.dial(%s) got error: %+v", address, err)
	} else {
		debug(true, "Connection.dial(%s) was successful.", address)
		c.Configure(conn)
	}
}
func (c *Connection) isValidParcel(parcel Parcel) bool {
	debug(true, "Connection.isValidParcel(%+v)", parcel)
	// TODO:  Check the network type (do we know it at this level?)
	// yeah NetworkID is a global
	// TODO:  Check the hash
	// TODO: Check all the header info, basically!
	return true
}

func (c *Connection) gotBadMessage() {
	debug(true, "Connection.gotBadMessage()")
	// TODO Track bad messages to ban bad peers at network level
	// Array of in Connection of bad messages
	// Add this one to the array with timestamp
	// Filter all messages with timestamps over an hour (put value in protocol.go maybe an hour is too logn)
	// If count of bad messages in last hour exceeds threshold from protocol.go then we drop connection
	// Add this IP address to our banned peers (for an hour or day, also define in protocol.go)
}

func (c *Connection) shutdown() {
	debug(true, "Connection.shutdown(%+v)", "")

	defer c.conn.Close()
}
