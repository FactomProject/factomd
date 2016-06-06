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

// Connection represents a single connection to another peer over the network. It communicates with the application
// via two channels, send and recieve.  These channels take structs of type ConnectionCommand or ConnectionParcel
// (defined below).
type Connection struct {
	conn           net.Conn
	SendChannel    chan interface{} // Send means "towards the network" Channel takes Parcels and ConnectionCommands
	ReceiveChannel chan interface{} // Recieve means "from the network" Channel sends Parcels and ConnectionCommands
	// and as "address" for sending messages to specific nodes.
	encoder         *gob.Encoder // Wire format is gobs in this version, may switch to binary
	decoder         *gob.Decoder // Wire format is gobs in this version, may switch to binary
	timeLastContact time.Time    // We track how recently we have heard from a peer to determin if it is still active.
	peer            Peer         // the datastructure representing the peer we are talking to. defined in peer.go
	attempts        int          // reconnection attempts
	timeLastAttempt time.Time    // time of last attempt to connect via dial
	timeLastPing    time.Time    // time of last ping sent
	timeLastUpdate  time.Time    // time of last peer update sent
	state           uint8        // Current state of the connection. Private. Only communication
	isOutGoing      bool         // We keep track of outgoing dial() vs incomming accept() connections
}

// Each connection is a simple state machine.  The state is managed by a single goroutine which also does netowrking.
// The flow is this:  Connection gets initialized, and either has a peer or a net connection (From an accept())
// If no network connection, the Connection dials.  If the dial is successful, it moves to the Online state
// If not, it moves to the Shutdown state-- we only dial out once when initialized with a peer.
// If we are online and get a network error, we shift to offline mode.  In offline state we attempt to reconnect for
// a period defined in protocol.go.  IF successful, we go back Online.  If too many attempts are made, we go to
// The ConnectionShutdown state, and exit the runloop.  In the Shutdown state we notify the controller so that we can be
// cleaned up.
const (
	ConnectionInitialized  uint8 = iota //Structure created, have peer info. Dial command moves us to Online or Shutdown (depending)
	ConnectionOnline                    // We're connected to the other side.  Normal state
	ConnectionOffline                   // We've been disconnected for whatever reason.  Attempt to reconnect some number of times. Moves to Online if successful, Shutdown if not.
	ConnectionShuttingDown              // We're shutting down, the recieves loop exits.
	// ConnectionReceivesShutdown              // The receives loop has exited, now waiting on the runloop
	ConnectionClosed // We're shut down, the runloop sets this state right before exiting. Controller can clean us up.
)

// Map of network ids to strings for easy printing of network ID
var connectionStateStrings = map[uint8]string{
	ConnectionInitialized:  "Initialized",
	ConnectionOnline:       "Online",
	ConnectionOffline:      "Offline",
	ConnectionShuttingDown: "Shutting Down",
	// ConnectionReceivesShutdown: "Receive Loop Shutting Down",
	ConnectionClosed: "Closed",
}

// ConnectionParcel is sent to convey an appication message destined for the network.
type ConnectionParcel struct {
	parcel Parcel
}

// ConnectionCommand is used to instruct the Connection to carry out some functionality.
type ConnectionCommand struct {
	command uint8
	peer    Peer
	delta   int32
}

// These are the commands that connections can send/recieve
const (
	ConnectionIsClosed uint8 = iota // Notifies the controller that we are shut down and can be released
	ConnectionShutdownNow
	ConnectionUpdatingPeer
	ConnectionAdjustPeerQuality
	ConnectionGoOffline // Notifies the connection it should go offinline (eg from another goroutine)
)

//////////////////////////////
//
// Public API
//
//////////////////////////////

// InitWithConn is called from our accept loop when a peer dials into us and we already have a network conn
func (c *Connection) InitWithConn(conn net.Conn, peer Peer) *Connection {
	c.conn = conn
	c.isOutGoing = false // InitWithConn is called by controller's accept() loop
	c.commonInit(peer)
	debug(c.peer.Hash, "Connection.InitWithConn() called.")
	c.goOnline()
	return c
}

// Init is called when we have peer info and need to dial into the peer
func (c *Connection) Init(peer Peer) *Connection {
	c.conn = nil
	c.isOutGoing = true
	c.commonInit(peer)
	debug(c.peer.Hash, "Connection.Init() called.")
	return c
}

func (c *Connection) IsOutGoing() bool {
	return c.isOutGoing
}

//////////////////////////////
//
// Private API
//
//////////////////////////////

func (c *Connection) commonInit(peer Peer) {
	c.state = ConnectionInitialized
	c.peer = peer
	c.SendChannel = make(chan interface{}, 10000)
	c.ReceiveChannel = make(chan interface{}, 10000)
	c.timeLastUpdate = time.Now()
	// go c.processReceives() // Need seperate goroutine for send and recieves since they are blocking, or long timeouts
	go c.runLoop() // handles sending messages, processing commands
}

// runloop OWNs the connection.  It is the only goroutine that can change values in the connection struct
// runLoop operates the state machine and routes messages out to network (messages from network are routed in processReceives)
func (c *Connection) runLoop() {
	for ConnectionClosed != c.state { // loop exits when we hit shutdown state
		// time.Sleep(time.Second * 1) // This can be a tight loop, don't want to starve the application
		time.Sleep(time.Millisecond * 1) // This can be a tight loop, don't want to starve the application
		switch c.state {
		case ConnectionInitialized:
			if c.dial() {
				c.goOnline()
			} else { //  we did not connect successfully
				// BUGBUG Go offline for testing now. Ask brian how tough we should be
				// c.goShutdown()
				c.goOffline()
			}
		case ConnectionOnline:
			c.processReceives()
			c.processSends()
			if ConnectionOnline == c.state {
				c.pingPeer() // sends a ping periodically if things have been quiet
				if PeerSaveInterval < time.Since(c.timeLastUpdate) {
					debug(c.peer.Hash, "runLoop() PeerSaveInterval interval %s is less than duration since last update: %s ", PeerSaveInterval.String(), time.Since(c.timeLastUpdate).String())
					c.updatePeer() // every PeerSaveInterval * 0.90 we send an update peer to the controller.
				}
			}
		case ConnectionOffline:
			duration := time.Since(c.timeLastAttempt)
			if TimeBetweenRedials < duration && MaxNumberOfRedialAttempts > c.attempts {
				if c.dial() {
					c.goOnline()
				} else { //  we did not connect successfully
					c.attempts++
					c.timeLastAttempt = time.Now()
				}
				if MaxNumberOfRedialAttempts <= c.attempts {
					c.goShutdown()
				} else {
					time.Sleep(TimeBetweenRedials)
				}
			}
		case ConnectionShuttingDown:
			debug(c.peer.Hash, "runLoop() ConnectionShuttingDown STATE runloop() cleaning up. ")
			c.state = ConnectionClosed
			c.ReceiveChannel <- ConnectionCommand{command: ConnectionIsClosed}
			return // ending runloop() goroutine
		default:
			logfatal(c.peer.Hash, "runLoop() unknown state?: %s ", connectionStateStrings[c.state])
		}
	}
}

func (c *Connection) dial() bool {
	note(c.peer.Hash, "Connection.dial() dialing: %+v", c.peer.Address)
	// conn, err := net.Dial("tcp", c.peer.Address)
	conn, err := net.DialTimeout("tcp", c.peer.Address, time.Second*10)
	if err != nil {
		silence(c.peer.Hash, "Connection.dial(%s) got error: %+v", c.peer.Address, err)
		return false
	}
	c.conn = conn
	debug(c.peer.Hash, "Connection.dial(%s) was successful.", c.peer.Address)
	return true
}

// Called when we are online and connected to the peer.
func (c *Connection) goOnline() {
	debug(c.peer.Hash, "Connection.goOnline() called. %s", c.peer.Hash)
	c.encoder = gob.NewEncoder(c.conn)
	c.decoder = gob.NewDecoder(c.conn)
	c.timeLastPing = time.Now()
	c.timeLastContact = time.Now()
	c.timeLastAttempt = time.Now()
	c.timeLastUpdate = time.Now()
	c.attempts = 0
	c.state = ConnectionOnline
	// Now ask the other side for the peers they know about.
	parcel := NewParcel(CurrentNetwork, []byte("Peer Request"))
	parcel.Header.Type = TypePeerRequest
	c.SendChannel <- ConnectionParcel{parcel: *parcel}
}

func (c *Connection) goOffline() {
	debug(c.peer.Hash, "Connection.goOffline()")
	c.state = ConnectionOffline
	c.attempts = 0
	c.peer.demerit()
}

func (c *Connection) goShutdown() {
	c.goOffline()
	if nil != c.conn {
		defer c.conn.Close()
	}
	c.decoder = nil
	c.encoder = nil
	c.state = ConnectionShuttingDown
}

// processSends gets all the messages from the application and sends them out over the network
func (c *Connection) processSends() {
	// note(c.peer.Hash, "Connection.processSends() called. Items in send channel: %d State: %s", len(c.SendChannel), c.ConnectionState())
	for 0 < len(c.SendChannel) && ConnectionOnline == c.state {
		message := <-c.SendChannel
		switch message.(type) {
		case ConnectionParcel:
			verbose(c.peer.Hash, "processSends() ConnectionParcel")
			parameters := message.(ConnectionParcel)
			c.sendParcel(parameters.parcel)
		case ConnectionCommand:
			verbose(c.peer.Hash, "processSends() ConnectionCommand")
			parameters := message.(ConnectionCommand)
			c.handleCommand(parameters)
		default:
			logfatal(c.peer.Hash, "processSends() unknown message?: %+v ", message)
		}
	}
}

func (c *Connection) handleCommand(command ConnectionCommand) {
	switch command.command {
	case ConnectionShutdownNow:
		c.goShutdown()
	case ConnectionUpdatingPeer: // at this level we're only updating the quality score, to pass on application level demerits
		debug(c.peer.Hash, "handleCommand() ConnectionUpdatingPeer")
		peer := command.peer
		if peer.QualityScore < c.peer.QualityScore {
			c.peer.QualityScore = peer.QualityScore
		}
	case ConnectionAdjustPeerQuality:
		debug(c.peer.Hash, "handleCommand() ConnectionAdjustPeerQuality")
		delta := command.delta
		c.peer.QualityScore = c.peer.QualityScore + delta
		if MinumumQualityScore > c.peer.QualityScore {
			debug(c.peer.Hash, "handleCommand() disconnecting peer: %s for quality score: %d", c.peer.Hash, c.peer.QualityScore)
			c.updatePeer()
			c.goShutdown()
		}
	case ConnectionGoOffline:
		debug(c.peer.Hash, "handleCommand() disconnecting peer: %s goOffline command recieved", c.peer.Hash)
		c.goOffline()
	default:
		logfatal(c.peer.Hash, "handleCommand() unknown command?: %+v ", command)
	}
}

func (c *Connection) sendParcel(parcel Parcel) {
	debug(c.peer.Hash, "sendParcel() sending message to network of type: %s", parcel.MessageType())
	parcel.Header.NodeID = NodeID // Send it out with our ID for loopback.
	verbose(c.peer.Hash, "sendParcel() Sanity check. State: %s Encoder: %+v, Parcel: %s", c.ConnectionState(), c.encoder, parcel.MessageType())
	c.conn.SetWriteDeadline(time.Now().Add(20 * time.Millisecond))
	err := c.encoder.Encode(parcel)
	switch {
	case nil == err:
		verbose(c.peer.Hash, "Connection.sendParcel() Timeout()  State: %s", c.ConnectionState())
	default:
		c.handleNetErrors(err)
		return
	}
}

// New version: Recieves is called as part of runloop
func (c *Connection) processReceives() {
	for ConnectionOnline == c.state {
		var message Parcel
		verbose(c.peer.Hash, "Connection.processReceives() called. State: %s", c.ConnectionState())
		c.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		err := c.decoder.Decode(&message)
		switch {
		case nil == err:
			note(c.peer.Hash, "Connection.processReceives() RECIEVED FROM NETWORK!  State: %s MessageType: %s", c.ConnectionState(), message.MessageType())
			c.handleParcel(message)
		default:
			c.handleNetErrors(err)
			return
		}
	}
}

//handleNetErrors Reacts to errors we get from encoder or decoder
func (c *Connection) handleNetErrors(err error) {
	nerr, isNetError := err.(net.Error)
	verbose(c.peer.Hash, "Connection.handleNetErrors() got error: %+v", err)
	switch {
	case isNetError && nerr.Timeout(): /// buffer empty
		return
	case isNetError && nerr.Temporary(): /// Temporary error, try to reconnect.
		c.goOffline()
	case io.EOF == err, io.ErrClosedPipe == err: // Remote hung up
		c.goOffline()
	default:
		silence(c.peer.Hash, "Connection.handleNetErrors() got unhandled coding error: %+v", err)
		c.goOffline()
	}

}

// handleParcel checks the parcel command type, and either generates a response, or passes it along.
func (c *Connection) handleParcel(parcel Parcel) {
	parcel.Header.Timestamp = time.Now() // set the timestamp to the recieved time.
	validity := c.parcelValidity(parcel)
	switch validity {
	case InvalidDisconnectPeer:
		debug(c.peer.Hash, "Connection.handleParcel() Disconnecting peer: %s", c.peer.Address)
		c.attempts = MaxNumberOfRedialAttempts + 50 // so we don't redial invalid Peer
		c.goShutdown()
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
	default:
		logfatal(c.peer.Hash, "handleParcel() unknown parcelValidity?: %+v ", validity)

	}
}

// These constants support the multiple penalties and responses for Parcel validation
const (
	ParcelValid           uint8 = iota
	InvalidPeerDemerit          // The peer sent an invalid message
	InvalidDisconnectPeer       // Eg they are on the wrong network or wrong version of the software
)

func (c *Connection) parcelValidity(parcel Parcel) uint8 {
	verbose(c.peer.Hash, "Connection.isValidParcel(%s)", parcel.MessageType())
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
		debug(c.peer.Hash, "handleParcelTypes() GOT PING, Sending Pong: %s", pong.String())
		c.SendChannel <- ConnectionParcel{parcel: parcel}
	case TypePong: // all we need is the timestamp which is set already
		debug(c.peer.Hash, "handleParcelTypes() GOT Pong.")
		return
	case TypePeerRequest:
		debug(c.peer.Hash, "handleParcelTypes() TypePeerRequest")
		c.ReceiveChannel <- ConnectionParcel{parcel: parcel} // Controller handles these.
	case TypePeerResponse:
		debug(c.peer.Hash, "handleParcelTypes() TypePeerResponse")
		c.ReceiveChannel <- ConnectionParcel{parcel: parcel} // Controller handles these.
	case TypeMessage:
		debug(c.peer.Hash, "handleParcelTypes() TypeMessage. Message is a: %s", parcel.MessageType())
		// Store our connection ID so the controller can direct response to us.
		parcel.Header.TargetPeer = c.peer.Hash
		parcel.Header.NodeID = NodeID
		c.ReceiveChannel <- ConnectionParcel{parcel: parcel}
	default:

		silence(c.peer.Hash, "!!!!!!!!!!!!!!!!!! Got message of unknown type?")
		parcel.Print()
	}
}

func (c *Connection) pingPeer() {
	durationLastContact := time.Since(c.timeLastContact)
	durationLastPing := time.Since(c.timeLastPing)
	if PingInterval < durationLastContact && PingInterval < durationLastPing {
		if MaxNumberOfRedialAttempts < c.attempts {
			debug(c.peer.Hash, "pingPeer() GOING OFFLINE - No response to pings. Attempts: %d Ti  since last contact: %s and time since last ping: %s", PingInterval.String(), durationLastContact.String(), durationLastPing.String())

			c.goOffline()
			return
		} else {
			verbose(c.peer.Hash, "pingPeer() Connection State: %s", c.ConnectionState())
			debug(c.peer.Hash, "pingPeer() Ping interval %s is less than duration since last contact: %s and time since last ping: %s", PingInterval.String(), durationLastContact.String(), durationLastPing.String())
			parcel := NewParcel(CurrentNetwork, []byte("Ping"))
			parcel.Header.Type = TypePing
			c.timeLastPing = time.Now()
			c.attempts++
			c.SendChannel <- ConnectionParcel{parcel: *parcel}
		}
	}
}

func (c *Connection) updatePeer() {
	verbose(c.peer.Hash, "updatePeer() SENDING ConnectionUpdatingPeer - Connection State: %s", c.ConnectionState())
	c.timeLastUpdate = time.Now()
	c.ReceiveChannel <- ConnectionCommand{command: ConnectionUpdatingPeer, peer: c.peer}
}

func (c *Connection) ConnectionState() string {
	return connectionStateStrings[c.state]
}
