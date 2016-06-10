// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"encoding/gob"
	"fmt"
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
	isPersistent    bool         // Persistent connections we always redail. BUGBUG - should this be handled by peer type logic?
	notes           string       // Notes about the connection, for debugging (eg: error)
	bytesSent       uint32       // Keeping track of the data sent/recieved for console
	bytesReceived   uint32       // Keeping track of the data sent/recieved for console
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
	ConnectionClosed                    // We're shut down, the runloop sets this state right before exiting. Controller can clean us up.
)

// Map of network ids to strings for easy printing of network ID
var connectionStateStrings = map[uint8]string{
	ConnectionInitialized:  "Initialized",
	ConnectionOnline:       "Online",
	ConnectionOffline:      "Offline",
	ConnectionShuttingDown: "Shutting Down",
	ConnectionClosed:       "Closed",
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
	c.isPersistent = false
	debug(c.peer.PeerIdent(), "Connection.InitWithConn() called.")
	c.goOnline()
	c.notes = "Incomming connection from accept()"
	return c
}

// Init is called when we have peer info and need to dial into the peer
func (c *Connection) Init(peer Peer, persistent bool) *Connection {
	c.conn = nil
	c.isOutGoing = true
	c.commonInit(peer)
	c.isPersistent = persistent
	debug(c.peer.PeerIdent(), "Connection.Init() called.")
	return c
}

func (c *Connection) IsOutGoing() bool {
	return c.isOutGoing
}

func (c *Connection) IsOnline() bool {
	return ConnectionOnline == c.state
}

func (c *Connection) IsPersistent() bool {
	return c.isPersistent
}
func (c *Connection) Notes() string {
	return c.notes
}

//////////////////////////////
//
// Private API
//
//////////////////////////////

func (c *Connection) commonInit(peer Peer) {
	c.state = ConnectionInitialized
	c.peer = peer
	c.notes = "initialized"
	c.SendChannel = make(chan interface{}, 10000)
	c.ReceiveChannel = make(chan interface{}, 10000)
	c.timeLastUpdate = time.Now()
	c.timeLastAttempt = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	// go c.processReceives() // Need seperate goroutine for send and recieves since they are blocking, or long timeouts
	go c.runLoop() // handles sending messages, processing commands
}

// runloop OWNs the connection.  It is the only goroutine that can change values in the connection struct
// runLoop operates the state machine and routes messages out to network (messages from network are routed in processReceives)
func (c *Connection) runLoop() {
	for ConnectionClosed != c.state { // loop exits when we hit shutdown state
		// time.Sleep(time.Second * 1) // This can be a tight loop, don't want to starve the application
		time.Sleep(time.Millisecond * 10) // This can be a tight loop, don't want to starve the application
		switch c.state {
		case ConnectionInitialized:
			// BUGBUG Note this means that we will redial ourselves if we are set as a persistent connection with ourselves as peer.
			if MinumumQualityScore > c.peer.QualityScore && !c.isPersistent {
				note(c.peer.PeerIdent(), "Connection.runloop(%s) ConnectionInitialized quality score too low: %d", c.peer.PeerIdent(), c.peer.QualityScore)
				c.updatePeer() // every PeerSaveInterval * 0.90 we send an update peer to the controller.
				c.goShutdown()
			} else {
				note(c.peer.PeerIdent(), "Connection.runLoop() ConnectionInitialized, going dialLoop().")
				c.dialLoop() // dialLoop dials until it connects or shuts down.
			}
		case ConnectionOnline:
			c.processSends()
			c.processReceives() // We may get messages that change state (Eg: loopback error)
			if ConnectionOnline == c.state {
				c.pingPeer() // sends a ping periodically if things have been quiet
				if PeerSaveInterval < time.Since(c.timeLastUpdate) {
					debug(c.peer.PeerIdent(), "runLoop() PeerSaveInterval interval %s is less than duration since last update: %s ", PeerSaveInterval.String(), time.Since(c.timeLastUpdate).String())
					c.updatePeer() // every PeerSaveInterval * 0.90 we send an update peer to the controller.
				}
			}
			if MinumumQualityScore > c.peer.QualityScore && !c.isPersistent {
				note(c.peer.PeerIdent(), "Connection.runloop(%s) ConnectionOnline quality score too low: %d", c.peer.PeerIdent(), c.peer.QualityScore)
				c.updatePeer() // every PeerSaveInterval * 0.90 we send an update peer to the controller.
				c.goShutdown()
			}
		case ConnectionOffline:
			note(c.peer.PeerIdent(), "Connection.runLoop() ConnectionOffline, going dialLoop().")
			c.dialLoop() // dialLoop dials until it connects or shuts down.
		case ConnectionShuttingDown:
			debug(c.peer.PeerIdent(), "runLoop() ConnectionShuttingDown STATE runloop() cleaning up. ")
			c.state = ConnectionClosed
			c.ReceiveChannel <- ConnectionCommand{command: ConnectionIsClosed}
			return // ending runloop() goroutine
		default:
			logfatal(c.peer.PeerIdent(), "runLoop() unknown state?: %s ", connectionStateStrings[c.state])
		}
	}
}

// dialLoop:  dials the connection until giving up. Called in offline or initializing states.
// All exits from dialLoop change the state of the connection allowing the outside run_loop to proceed.
func (c *Connection) dialLoop() {
	c.notes = fmt.Sprintf("dialLoop() dialing: %+v", c.peer.PeerIdent())
	for {
		elapsed := time.Since(c.timeLastAttempt)
		debug(c.peer.PeerIdent(), "Connection.dialLoop() elapsed: %s Attempts: %d", elapsed.String(), c.attempts)
		if TimeBetweenRedials < elapsed {
			c.timeLastAttempt = time.Now()
			switch c.dial() {
			case true:
				note(c.peer.PeerIdent(), "Connection.dialLoop() Connected, going online.")
				c.goOnline()
				return
			case false:
				switch {
				case c.isPersistent:
					note(c.peer.PeerIdent(), "Connection.dialLoop() Persistent connection - Sleeping until next redial.")
					time.Sleep(TimeBetweenRedials)
				case !c.isOutGoing: // incomming connection we redial once, then give up.
					note(c.peer.PeerIdent(), "Connection.dialLoop() Incomming Connections - One Shot dial, so we're shutting down.")
					c.goShutdown()
					return
				case ConnectionInitialized == c.state:
					note(c.peer.PeerIdent(), "Connection.dialLoop() ConnectionInitialized - One Shot dial, so we're shutting down.")
					c.goShutdown() // We're dialing possibly many peers who are no longer there.
					return
				case ConnectionOffline == c.state: // We were online with the peer at one point.
					note(c.peer.PeerIdent(), "Connection.dialLoop() ConnectionOffline - Attempts: %d - since redial: %s TimeBetweenRedials: %s", c.attempts, elapsed.String(), TimeBetweenRedials.String())
					c.attempts++
					switch {
					case MaxNumberOfRedialAttempts < c.attempts:
						note(c.peer.PeerIdent(), "Connection.dialLoop() MaxNumberOfRedialAttempts < Attempts: %d - since redial: %s TimeBetweenRedials: %s", c.attempts, elapsed.String(), TimeBetweenRedials.String())
						c.goShutdown()
						return
					default:
						note(c.peer.PeerIdent(), "Connection.dialLoop() MaxNumberOfRedialAttempts > Attempts: %d - since redial: %s TimeBetweenRedials: %s", c.attempts, elapsed.String(), TimeBetweenRedials.String())
						time.Sleep(TimeBetweenRedials)
					}
				}
			}
		} else {
			note(c.peer.PeerIdent(), "Connection.dialLoop() TimeBetweenRedials > elapsed")
			time.Sleep(TimeBetweenRedials)
		}
	}
}

// dial() handles connection logic and shifts states based on results.
func (c *Connection) dial() bool {
	address := c.peer.AddressPort()
	note(c.peer.PeerIdent(), "Connection.dial() dialing: %+v", address)
	// conn, err := net.Dial("tcp", c.peer.Address)
	conn, err := net.DialTimeout("tcp", address, time.Second*10)
	if nil != err {
		note(c.peer.PeerIdent(), "Connection.dial(%s) got error: %+v", address, err)
		c.notes = fmt.Sprintf("dial() error: %+v", err)
		return false
	}
	c.conn = conn
	note(c.peer.PeerIdent(), "Connection.dial(%s) was successful.", address)
	c.notes = "Successful Dial"
	return true
}

// Called when we are online and connected to the peer.
func (c *Connection) goOnline() {
	debug(c.peer.PeerIdent(), "Connection.goOnline() called.")
	c.state = ConnectionOnline
	now := time.Now()
	c.encoder = gob.NewEncoder(c.conn)
	c.decoder = gob.NewDecoder(c.conn)
	c.attempts = 0
	c.timeLastPing = now
	c.timeLastContact = now
	c.timeLastAttempt = now
	c.timeLastUpdate = now
	c.peer.LastContact = now
	// Now ask the other side for the peers they know about.
	parcel := NewParcel(CurrentNetwork, []byte("Peer Request"))
	parcel.Header.Type = TypePeerRequest
	c.SendChannel <- ConnectionParcel{parcel: *parcel}
}

func (c *Connection) goOffline() {
	debug(c.peer.PeerIdent(), "Connection.goOffline()")
	c.state = ConnectionOffline
	c.attempts = 0
	c.peer.demerit()
}

func (c *Connection) goShutdown() {
	c.goOffline()
	c.updatePeer()
	if nil != c.conn {
		defer c.conn.Close()
	}
	c.decoder = nil
	c.encoder = nil
	c.state = ConnectionShuttingDown
}

// processSends gets all the messages from the application and sends them out over the network
func (c *Connection) processSends() {
	// note(c.peer.PeerIdent(), "Connection.processSends() called. Items in send channel: %d State: %s", len(c.SendChannel), c.ConnectionState())
	for 0 < len(c.SendChannel) && ConnectionOnline == c.state {
		message := <-c.SendChannel
		switch message.(type) {
		case ConnectionParcel:
			verbose(c.peer.PeerIdent(), "processSends() ConnectionParcel")
			parameters := message.(ConnectionParcel)
			c.sendParcel(parameters.parcel)
		case ConnectionCommand:
			verbose(c.peer.PeerIdent(), "processSends() ConnectionCommand")
			parameters := message.(ConnectionCommand)
			c.handleCommand(parameters)
		default:
			logfatal(c.peer.PeerIdent(), "processSends() unknown message?: %+v ", message)
		}
	}
}

func (c *Connection) handleCommand(command ConnectionCommand) {
	switch command.command {
	case ConnectionShutdownNow:
		c.goShutdown()
	case ConnectionUpdatingPeer: // at this level we're only updating the quality score, to pass on application level demerits
		debug(c.peer.PeerIdent(), "handleCommand() ConnectionUpdatingPeer")
		peer := command.peer
		if peer.QualityScore < c.peer.QualityScore {
			c.peer.QualityScore = peer.QualityScore
		}
	case ConnectionAdjustPeerQuality:
		debug(c.peer.PeerIdent(), "handleCommand() ConnectionAdjustPeerQuality")
		delta := command.delta
		c.peer.QualityScore = c.peer.QualityScore + delta
		if MinumumQualityScore > c.peer.QualityScore {
			debug(c.peer.PeerIdent(), "handleCommand() disconnecting peer: %s for quality score: %d", c.peer.PeerIdent(), c.peer.QualityScore)
			c.updatePeer()
			c.goShutdown()
		}
	case ConnectionGoOffline:
		debug(c.peer.PeerIdent(), "handleCommand() disconnecting peer: %s goOffline command recieved", c.peer.PeerIdent())
		c.goOffline()
	default:
		logfatal(c.peer.PeerIdent(), "handleCommand() unknown command?: %+v ", command)
	}
}

func (c *Connection) sendParcel(parcel Parcel) {
	debug(c.peer.PeerIdent(), "sendParcel() sending message to network of type: %s", parcel.MessageType())
	parcel.Header.NodeID = NodeID // Send it out with our ID for loopback.
	verbose(c.peer.PeerIdent(), "sendParcel() Sanity check. State: %s Encoder: %+v, Parcel: %s", c.ConnectionState(), c.encoder, parcel.MessageType())
	c.conn.SetWriteDeadline(time.Now().Add(20 * time.Millisecond))
	err := c.encoder.Encode(parcel)
	switch {
	case nil == err:
		c.bytesSent += parcel.Header.Length
	default:
		c.handleNetErrors(err)
	}
}

// New version: Recieves is called as part of runloop
func (c *Connection) processReceives() {
	for ConnectionOnline == c.state {
		var message Parcel
		verbose(c.peer.PeerIdent(), "Connection.processReceives() called. State: %s", c.ConnectionState())
		c.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		err := c.decoder.Decode(&message)
		switch {
		case nil == err:
			note(c.peer.PeerIdent(), "Connection.processReceives() RECIEVED FROM NETWORK!  State: %s MessageType: %s", c.ConnectionState(), message.MessageType())
			c.bytesReceived += message.Header.Length
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
	verbose(c.peer.PeerIdent(), "Connection.handleNetErrors() got error: %+v", err)
	switch {
	case isNetError && nerr.Timeout(): /// buffer empty
		return
	case isNetError && nerr.Temporary(): /// Temporary error, try to reconnect.
		c.notes = fmt.Sprintf("handleNetErrors() Temporary error: %+v", nerr)
		c.goOffline()
	case io.EOF == err, io.ErrClosedPipe == err: // Remote hung up
		c.notes = fmt.Sprintf("handleNetErrors() Remote hung up - error: %+v", err)
		c.goOffline()
	default:
		significant(c.peer.PeerIdent(), "Connection.handleNetErrors() got unhandled coding error: %+v", err)
		c.notes = fmt.Sprintf("handleNetErrors() Unhandled error: %+v", err)
		c.goOffline()
	}

}

// handleParcel checks the parcel command type, and either generates a response, or passes it along.
// return value:  Indicate whether we got a good message or not and thus whether we should keep reading from network
func (c *Connection) handleParcel(parcel Parcel) {
	c.peer.Port = parcel.Header.PeerPort // Peers communicate their port in the header. Could be moved to a handshake
	validity := c.parcelValidity(parcel)
	switch validity {
	case InvalidDisconnectPeer:
		debug(c.peer.PeerIdent(), "Connection.handleParcel() Disconnecting peer: %s", c.peer.PeerIdent())
		c.attempts = MaxNumberOfRedialAttempts + 50 // so we don't redial invalid Peer
		c.goShutdown()
		return
	case InvalidPeerDemerit:
		debug(c.peer.PeerIdent(), "Connection.handleParcel() got invalid message")
		parcel.Print()
		c.peer.demerit()
		return
	case ParcelValid:
		c.timeLastContact = time.Now() // We only update for valid messages (incluidng pings and heartbeats)
		c.attempts = 0                 // reset since we are clearly in touch now.
		c.peer.merit()                 // Increase peer quality score.
		debug(c.peer.PeerIdent(), "Connection.handleParcel() got ParcelValid %s", parcel.MessageType())
		if Notes <= CurrentLoggingLevel {
			parcel.PrintMessageType()
		}
		c.handleParcelTypes(parcel) // handles both network commands and application messages
		return
	default:
		logfatal(c.peer.PeerIdent(), "handleParcel() unknown parcelValidity?: %+v ", validity)
		return
	}
}

// These constants support the multiple penalties and responses for Parcel validation
const (
	ParcelValid           uint8 = iota
	InvalidPeerDemerit          // The peer sent an invalid message
	InvalidDisconnectPeer       // Eg they are on the wrong network or wrong version of the software
)

func (c *Connection) parcelValidity(parcel Parcel) uint8 {
	verbose(c.peer.PeerIdent(), "Connection.isValidParcel(%s)", parcel.MessageType())
	crc := crc32.Checksum(parcel.Payload, CRCKoopmanTable)
	switch {
	case parcel.Header.NodeID == NodeID: // We are talking to ourselves!
		significant(c.peer.PeerIdent(), "Connection.isValidParcel(), failed due to loopback!: %+v", parcel.Header)
		return InvalidDisconnectPeer
	case parcel.Header.Network != CurrentNetwork:
		significant(c.peer.PeerIdent(), "Connection.isValidParcel(), failed due to wrong network. Remote: %0x Us: %0x", parcel.Header.Network, CurrentNetwork)
		return InvalidDisconnectPeer
	case parcel.Header.Version < ProtocolVersionMinimum:
		significant(c.peer.PeerIdent(), "Connection.isValidParcel(), failed due to wrong version: %+v", parcel.Header)
		return InvalidDisconnectPeer
	case parcel.Header.Length != uint32(len(parcel.Payload)):
		significant(c.peer.PeerIdent(), "Connection.isValidParcel(), failed due to wrong length: %+v", parcel.Header)
		return InvalidPeerDemerit
	case parcel.Header.Crc32 != crc:
		significant(c.peer.PeerIdent(), "Connection.isValidParcel(), failed due to bad checksum: %+v", parcel.Header)
		return InvalidPeerDemerit
	default:
		return ParcelValid
	}
	return ParcelValid
}
func (c *Connection) handleParcelTypes(parcel Parcel) {
	switch parcel.Header.Type {
	case TypeAlert:
		significant(c.peer.PeerIdent(), "!!!!!!!!!!!!!!!!!! Alert: TODO Alert signature checking not supported yet! BUGBUG")
	case TypePing:
		// Send Pong
		pong := NewParcel(CurrentNetwork, []byte("Pong"))
		pong.Header.Type = TypePong
		debug(c.peer.PeerIdent(), "handleParcelTypes() GOT PING, Sending Pong: %s", pong.String())
		parcel.Print()
		c.SendChannel <- ConnectionParcel{parcel: *pong}
	case TypePong: // all we need is the timestamp which is set already
		debug(c.peer.PeerIdent(), "handleParcelTypes() GOT Pong.")
		return
	case TypePeerRequest:
		debug(c.peer.PeerIdent(), "handleParcelTypes() TypePeerRequest")
		c.ReceiveChannel <- ConnectionParcel{parcel: parcel} // Controller handles these.
	case TypePeerResponse:
		debug(c.peer.PeerIdent(), "handleParcelTypes() TypePeerResponse")
		c.ReceiveChannel <- ConnectionParcel{parcel: parcel} // Controller handles these.
	case TypeMessage:
		debug(c.peer.PeerIdent(), "handleParcelTypes() TypeMessage. Message is a: %s", parcel.MessageType())
		// Store our connection ID so the controller can direct response to us.
		parcel.Header.TargetPeer = c.peer.Hash
		parcel.Header.NodeID = NodeID
		c.ReceiveChannel <- ConnectionParcel{parcel: parcel}
	default:

		significant(c.peer.PeerIdent(), "!!!!!!!!!!!!!!!!!! Got message of unknown type?")
		parcel.Print()
	}
}

func (c *Connection) pingPeer() {
	durationLastContact := time.Since(c.timeLastContact)
	durationLastPing := time.Since(c.timeLastPing)
	if PingInterval < durationLastContact && PingInterval < durationLastPing {
		if MaxNumberOfRedialAttempts < c.attempts {
			debug(c.peer.PeerIdent(), "pingPeer() GOING OFFLINE - No response to pings. Attempts: %d Ti  since last contact: %s and time since last ping: %s", PingInterval.String(), durationLastContact.String(), durationLastPing.String())
			c.goOffline()
			return
		} else {
			verbose(c.peer.PeerIdent(), "pingPeer() Connection State: %s", c.ConnectionState())
			debug(c.peer.PeerIdent(), "pingPeer() Ping interval %s is less than duration since last contact: %s and time since last ping: %s", PingInterval.String(), durationLastContact.String(), durationLastPing.String())
			parcel := NewParcel(CurrentNetwork, []byte("Ping"))
			parcel.Header.Type = TypePing
			c.timeLastPing = time.Now()
			c.attempts++
			c.SendChannel <- ConnectionParcel{parcel: *parcel}
		}
	}
}

func (c *Connection) updatePeer() {
	verbose(c.peer.PeerIdent(), "updatePeer() SENDING ConnectionUpdatingPeer - Connection State: %s", c.ConnectionState())
	c.timeLastUpdate = time.Now()
	c.ReceiveChannel <- ConnectionCommand{command: ConnectionUpdatingPeer, peer: c.peer}
}

func (c *Connection) ConnectionState() string {
	return connectionStateStrings[c.state]
}
