// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"encoding/gob"
	"fmt"
	"hash/crc32"
	"io"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/FactomProject/factomd/common/primitives"
)

// Connection represents a single connection to another peer over the network. It communicates with the application
// via two channels, send and recieve.  These channels take structs of type ConnectionCommand or ConnectionParcel
// (defined below).
type Connection struct {
	conn           net.Conn
	Errors         chan error             // handle errors from connections.
	Commands       chan ConnectionCommand // handle connection commands
	SendChannel    chan interface{}       // Send means "towards the network" Channel sends Parcels and ConnectionCommands
	ReceiveChannel chan interface{}       // Recieve means "from the network" Channel recieves Parcels and ConnectionCommands
	// and as "address" for sending messages to specific nodes.
	encoder         *gob.Encoder      // Wire format is gobs in this version, may switch to binary
	decoder         *gob.Decoder      // Wire format is gobs in this version, may switch to binary
	peer            Peer              // the datastructure representing the peer we are talking to. defined in peer.go
	attempts        int               // reconnection attempts
	timeLastAttempt time.Time         // time of last attempt to connect via dial
	timeLastPing    time.Time         // time of last ping sent
	timeLastUpdate  time.Time         // time of last peer update sent
	timeLastStatus  time.Time         // last time we printed our status for debugging.
	timeLastMetrics time.Time         // last time we updated metrics
	state           uint8             // Current state of the connection. Private. Only communication
	isOutGoing      bool              // We keep track of outgoing dial() vs incomming accept() connections
	isPersistent    bool              // Persistent connections we always redail.
	notes           string            // Notes about the connection, for debugging (eg: error)
	metrics         ConnectionMetrics // Metrics about this connection
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
	Parcel Parcel
}

func (e *ConnectionParcel) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *ConnectionParcel) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *ConnectionParcel) String() string {
	str, _ := e.JSONString()
	return str
}

// ConnectionMetrics is used to encapsulate various metrics about the connection.
type ConnectionMetrics struct {
	MomentConnected  time.Time // when the connection started.
	BytesSent        uint32    // Keeping track of the data sent/recieved for console
	BytesReceived    uint32    // Keeping track of the data sent/recieved for console
	MessagesSent     uint32    // Keeping track of the data sent/recieved for console
	MessagesReceived uint32    // Keeping track of the data sent/recieved for console
	PeerAddress      string    // Peer IP Address
	PeerQuality      int32     // Quality of the connection.
	// Red: Below -50
	// Yellow: -50 - 100
	// Green: > 100
	ConnectionState string // Basic state of the connection
	ConnectionNotes string // Connectivity notes for the connection
}

// ConnectionCommand is used to instruct the Connection to carry out some functionality.
type ConnectionCommand struct {
	Command uint8
	Peer    Peer
	Delta   int32
	Metrics ConnectionMetrics
}

func (e *ConnectionCommand) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *ConnectionCommand) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *ConnectionCommand) String() string {
	str, _ := e.JSONString()
	return str
}

// These are the commands that connections can send/recieve
const (
	ConnectionIsClosed uint8 = iota // Notifies the controller that we are shut down and can be released
	ConnectionShutdownNow
	ConnectionUpdatingPeer
	ConnectionAdjustPeerQuality
	ConnectionUpdateMetrics
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
	c.setNotes("Incomming connection from accept()")
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

func (c *Connection) StatusString() string {
	return connectionStateStrings[c.state]
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
	c.setNotes("commonInit()")
	c.Errors = make(chan error, StandardChannelSize)
	c.Commands = make(chan ConnectionCommand, StandardChannelSize)
	c.SendChannel = make(chan interface{}, StandardChannelSize)
	c.ReceiveChannel = make(chan interface{}, StandardChannelSize)
	c.metrics = ConnectionMetrics{MomentConnected: time.Now()}
	c.timeLastMetrics = time.Now()
	c.timeLastAttempt = time.Now()
	c.timeLastStatus = time.Now()
}

func (c *Connection) Start() {
	go c.runLoop()
}

// runloop OWNs the connection.  It is the only goroutine that can change values in the connection struct
// runLoop operates the state machine and routes messages out to network (messages from network are routed in processReceives)
func (c *Connection) runLoop() {
	go c.processSends()
	go c.processReceives()

	for ConnectionClosed != c.state { // loop exits when we hit shutdown state
		time.Sleep(100 * time.Millisecond)
		// time.Sleep(time.Second * 1) // This can be a tight loop, don't want to starve the application
		c.updateStats() // Update controller with metrics
		c.connectionStatusReport()
		// if 2 == rand.Intn(100) {
		debug(c.peer.PeerFixedIdent(), "Connection.runloop() STATE IS: %s", connectionStateStrings[c.state])
		// }
		c.handleNetErrors()
		c.handleCommand()
		switch c.state {
		case ConnectionInitialized:
			if MinumumQualityScore > c.peer.QualityScore && !c.isPersistent {
				c.setNotes("Connection.runloop(%s) ConnectionInitialized quality score too low: %d", c.peer.PeerIdent(), c.peer.QualityScore)
				c.updatePeer() // every PeerSaveInterval * 0.90 we send an update peer to the controller.
				c.goShutdown()
			} else {
				c.setNotes("Connection.runLoop() ConnectionInitialized, going dialLoop(). %+v", c.peer.PeerIdent())
				c.dialLoop() // dialLoop dials until it connects or shuts down.
			}
		case ConnectionOnline:

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
			switch {
			case c.isOutGoing:
				note(c.peer.PeerIdent(), "Connection.runLoop() ConnectionOffline, going dialLoop().")
				c.dialLoop() // dialLoop dials until it connects or shuts down.
			default: // the connection dialed us, so we shutdown
				c.goShutdown()
			}
		case ConnectionShuttingDown:
			note(c.peer.PeerIdent(), "runLoop() in ConnectionShuttingDown state. The runloop() is sending ConnectionCommand{command: ConnectionIsClosed} Notes: %s", c.notes)
			c.state = ConnectionClosed
			BlockFreeChannelSend(c.ReceiveChannel, ConnectionCommand{Command: ConnectionIsClosed})
			fmt.Println(fmt.Sprintf("Connection(%s) has shut down. \nNotes: %s\n\n", c.peer.Address, c.notes))
			return // ending runloop() goroutine
		default:
			logfatal(c.peer.PeerIdent(), "runLoop() unknown state?: %s ", connectionStateStrings[c.state])
		}
	}
	significant(c.peer.PeerIdent(), "runLoop() Connection runloop() exiting %+v", c)
}

func (c *Connection) setNotes(format string, v ...interface{}) {
	c.notes = fmt.Sprintf(format, v...)
	significant(c.peer.PeerIdent(), c.notes)
}

// dialLoop:  dials the connection until giving up. Called in offline or initializing states.
// All exits from dialLoop change the state of the connection allowing the outside run_loop to proceed.
func (c *Connection) dialLoop() {
	c.setNotes(fmt.Sprintf("dialLoop() dialing: %+v", c.peer.PeerIdent()))
	if c.peer.QualityScore < MinumumQualityScore {
		c.setNotes("Connection.dialLoop() Quality Score too low, not dialing out again.")
		c.goShutdown()
		return
	}
	for {
		elapsed := time.Since(c.timeLastAttempt)
		debug(c.peer.PeerIdent(), "Connection.dialLoop() elapsed: %s Attempts: %d", elapsed.String(), c.attempts)
		if TimeBetweenRedials < elapsed {
			c.timeLastAttempt = time.Now()
			switch c.dial() {
			case true:
				c.setNotes("Connection.dialLoop() Connected, going online.")
				c.goOnline()
				return
			case false:
				switch {
				case c.isPersistent:
					c.setNotes("Connection.dialLoop() Persistent connection - Sleeping until next redial.")
					time.Sleep(TimeBetweenRedials)
				case !c.isOutGoing: // incomming connection we redial once, then give up.
					c.setNotes("Connection.dialLoop() Incomming Connection - One Shot re-dial, so we're shutting down. Last note was: %s", c.notes)
					c.goShutdown()
					return
				case ConnectionInitialized == c.state:
					c.setNotes("Connection.dialLoop() ConnectionInitialized - One Shot dial, so we're shutting down. Last note was: %s", c.notes)
					c.goShutdown() // We're dialing possibly many peers who are no longer there.
					return
				case ConnectionOffline == c.state: // We were online with the peer at one point.
					c.setNotes(fmt.Sprintf("Connection.dialLoop() ConnectionOffline - Attempts: %d - since redial: %s TimeBetweenRedials: %s", c.attempts, elapsed.String(), TimeBetweenRedials.String()))
					c.attempts++
					switch {
					case MaxNumberOfRedialAttempts < c.attempts:
						c.setNotes(fmt.Sprintf("Connection.dialLoop() MaxNumberOfRedialAttempts < Attempts: %d - since redial: %s TimeBetweenRedials: %s", c.attempts, elapsed.String(), TimeBetweenRedials.String()))
						c.goShutdown()
						return
					default:
						c.setNotes(fmt.Sprintf("Connection.dialLoop() MaxNumberOfRedialAttempts > Attempts: %d - since redial: %s TimeBetweenRedials: %s", c.attempts, elapsed.String(), TimeBetweenRedials.String()))
						time.Sleep(TimeBetweenRedials)
					}
				}
			}
		} else {
			c.setNotes("Connection.dialLoop() TimeBetweenRedials > elapsed")
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
		c.setNotes(fmt.Sprintf("Connection.dial(%s) got error: %+v", address, err))
		return false
	}
	c.conn = conn
	c.setNotes(fmt.Sprintf("Connection.dial(%s) was successful.", address))
	fmt.Println(fmt.Sprintf("Connection.dial(%s) was successful.", address))
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
	c.timeLastAttempt = now
	c.timeLastUpdate = now
	c.peer.LastContact = now
	// Probably shouldn't reset metrics when we go online. (Eg: say after a temp network problem)
	// c.metrics = ConnectionMetrics{MomentConnected: now} // Reset metrics
	// Now ask the other side for the peers they know about.
	parcel := NewParcel(CurrentNetwork, []byte("Peer Request"))
	parcel.Header.Type = TypePeerRequest
	BlockFreeChannelSend(c.SendChannel, ConnectionParcel{Parcel: *parcel})
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
	for ConnectionClosed != c.state && c.state != ConnectionShuttingDown {
		if nil == c.decoder || nil == c.conn {
			time.Sleep(100*time.Millisecond)
			continue
		}
		// note(c.peer.PeerIdent(), "Connection.processSends() called. Items in send channel: %d State: %s", len(c.SendChannel), c.ConnectionState())
		for ConnectionOnline == c.state {
			message := <-c.SendChannel
			switch message.(type) {
			case ConnectionParcel:
				parameters := message.(ConnectionParcel)
				c.sendParcel(parameters.Parcel)
			case ConnectionCommand:
				parameters := message.(ConnectionCommand)
				c.Commands <- parameters
			default:
			}
		}
		time.Sleep(300 * time.Millisecond)
	}
}

func (c *Connection) handleCommand() {
	select {
	case command := <-c.Commands:

		switch command.Command {
		case ConnectionShutdownNow:
			c.setNotes(fmt.Sprintf("Connection(%s) shutting down due to ConnectionShutdownNow message.", c.peer.AddressPort()))
			c.goShutdown()
		case ConnectionUpdatingPeer: // at this level we're only updating the quality score, to pass on application level demerits
			debug(c.peer.PeerIdent(), "handleCommand() ConnectionUpdatingPeer")
			peer := command.Peer
			if peer.QualityScore < c.peer.QualityScore {
				c.peer.QualityScore = peer.QualityScore
			}
		case ConnectionAdjustPeerQuality:
			delta := command.Delta
			note(c.peer.PeerIdent(), "handleCommand() ConnectionAdjustPeerQuality: Current Score: %d Delta: %d", c.peer.QualityScore, delta)
			c.peer.QualityScore = c.peer.QualityScore + delta
			if MinumumQualityScore > c.peer.QualityScore {
				debug(c.peer.PeerIdent(), "handleCommand() disconnecting peer: %s for quality score: %d", c.peer.PeerIdent(), c.peer.QualityScore)
				c.updatePeer()
				c.setNotes(fmt.Sprintf("Connection(%s) shutting down due to QualityScore %d being below MinumumQualityScore: %d.", c.peer.AddressPort(), c.peer.QualityScore, MinumumQualityScore))
				c.goShutdown()
			}
		case ConnectionGoOffline:
			debug(c.peer.PeerIdent(), "handleCommand() disconnecting peer: %s goOffline command recieved", c.peer.PeerIdent())
			c.goOffline()
		default:
			logfatal(c.peer.PeerIdent(), "handleCommand() unknown command?: %+v ", command)
		}
	default:
	}
}

func (c *Connection) sendParcel(parcel Parcel) {
	parcel.Header.NodeID = NodeID // Send it out with our ID for loopback.
	c.conn.SetWriteDeadline(time.Now().Add(NetworkDeadline * 500))

	//deadline := time.Now().Add(NetworkDeadline)
	//if len(parcel.Payload) > 1000*10 {
	//	ms := (len(parcel.Payload) * NetworkDeadline.Seconds())/1000
	//	deadline = time.Now().Add(time.Duration(ms)*time.Millisecond)
	//}
	//c.conn.SetWriteDeadline(deadline)
	encode := c.encoder
	err := encode.Encode(parcel)
	switch {
	case nil == err:
		c.metrics.BytesSent += parcel.Header.Length
		c.metrics.MessagesSent += 1
	default:
		c.Errors <- err
	}
}

// processReceives is a go routine This is essentially an infinite loop that exits
// when:
// -- a network error happens
// -- something causes our state to be offline
func (c *Connection) processReceives() {
	for ConnectionClosed != c.state && c.state != ConnectionShuttingDown {
		var message Parcel
		
		if nil == c.conn || nil == c.decoder {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		c.conn.SetReadDeadline(time.Now().Add(NetworkDeadline))
		err := c.decoder.Decode(&message)
		switch {
		case nil == err:
			c.metrics.BytesReceived += message.Header.Length
			c.metrics.MessagesReceived += 1
			message.Header.PeerAddress = c.peer.Address
			c.handleParcel(message)
		default:
			c.Errors <- err
		}
		time.Sleep(100 * time.Millisecond)
	}
}

//handleNetErrors Reacts to errors we get from encoder or decoder
func (c *Connection) handleNetErrors() {
	select {
	case err := <-c.Errors:
		nerr, isNetError := err.(net.Error)
		verbose(c.peer.PeerIdent(), "Connection.handleNetErrors() State: %s We got error: %+v", c.ConnectionState(), err)
		switch {
		case isNetError && nerr.Timeout(): /// buffer empty
			return
		case isNetError && nerr.Temporary(): /// Temporary error, try to reconnect.
			c.setNotes(fmt.Sprintf("handleNetErrors() Temporary error: %+v", nerr))
			c.goOffline()
		case io.EOF == err, io.ErrClosedPipe == err: // Remote hung up
			c.setNotes(fmt.Sprintf("handleNetErrors() Remote hung up - error: %+v", err))
			c.goOffline()
		case err == syscall.EPIPE: // "write: broken pipe"
			c.setNotes(fmt.Sprintf("handleNetErrors() Broken Pipe: %+v", err))
			c.goOffline()
		default:
			significant(c.peer.PeerIdent(), "Connection.handleNetErrors() State: %s We got unhandled coding error: %+v", c.ConnectionState(), err)
			c.setNotes(fmt.Sprintf("handleNetErrors() Unhandled error: %+v", err))
			c.goOffline()
		}
	default:
	}
}

// handleParcel checks the parcel command type, and either generates a response, or passes it along.
// return value:  Indicate whether we got a good message or not and thus whether we should keep reading from network
func (c *Connection) handleParcel(parcel Parcel) {
	defer func() {
		if r := recover(); r != nil {
			c.peer.demerit() /// so someone DDoS or just incompatible will eventually be cut off after 200+ panics
			fmt.Fprintf(os.Stdout, "Caught Exception in connection %s: %v\n", c.peer.PeerFixedIdent(), r)
			return
		}
	}()

	c.peer.Port = parcel.Header.PeerPort // Peers communicate their port in the header. Could be moved to a handshake
	validity := c.parcelValidity(parcel)
	switch validity {
	case InvalidDisconnectPeer:
		parcel.Trace("Connection.handleParcel()-InvalidDisconnectPeer", "I")
		debug(c.peer.PeerIdent(), "Connection.handleParcel() Disconnecting peer: %s", c.peer.PeerIdent())
		c.attempts = MaxNumberOfRedialAttempts + 50 // so we don't redial invalid Peer
		c.setNotes(fmt.Sprintf("Connection(%s) shutting down due to InvalidDisconnectPeer result from parcel. Previous notes: %s.", c.peer.AddressPort(), c.notes))
		c.goShutdown()
		return
	case InvalidPeerDemerit:
		parcel.Trace("Connection.handleParcel()-InvalidPeerDemerit", "I")
		debug(c.peer.PeerIdent(), "Connection.handleParcel() got invalid message")
		parcel.Print()
		c.peer.demerit()
		return
	case ParcelValid:
		parcel.Trace("Connection.handleParcel()-ParcelValid", "I")
		c.peer.LastContact = time.Now() // We only update for valid messages (incluidng pings and heartbeats)
		c.attempts = 0                  // reset since we are clearly in touch now.
		c.peer.merit()                  // Increase peer quality score.
		debug(c.peer.PeerIdent(), "Connection.handleParcel() got ParcelValid %s", parcel.MessageType())
		if Notes <= CurrentLoggingLevel {
			parcel.PrintMessageType()
		}
		c.handleParcelTypes(parcel) // handles both network commands and application messages
		return
	default:
		parcel.Trace("Connection.handleParcel()-fatal", "I")
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
		parcel.Trace("Connection.isValidParcel()-loopback", "H")
		c.setNotes(fmt.Sprintf("Connection.isValidParcel(), failed due to loopback!: %+v", parcel.Header))
		c.peer.QualityScore = MinumumQualityScore - 50 // Ban ourselves for a week
		return InvalidDisconnectPeer
	case parcel.Header.Network != CurrentNetwork:
		parcel.Trace("Connection.isValidParcel()-network", "H")
		c.setNotes(fmt.Sprintf("Connection.isValidParcel(), failed due to wrong network. Remote: %0x Us: %0x", parcel.Header.Network, CurrentNetwork))
		return InvalidDisconnectPeer
	case parcel.Header.Version < ProtocolVersionMinimum:
		parcel.Trace("Connection.isValidParcel()-version", "H")
		c.setNotes(fmt.Sprintf("Connection.isValidParcel(), failed due to wrong version: %+v", parcel.Header))
		return InvalidDisconnectPeer
	case parcel.Header.Length != uint32(len(parcel.Payload)):
		parcel.Trace("Connection.isValidParcel()-length", "H")
		significant(c.peer.PeerIdent(), "Connection.isValidParcel(), failed due to wrong length: %+v", parcel.Header)
		return InvalidPeerDemerit
	case parcel.Header.Crc32 != crc:
		parcel.Trace("Connection.isValidParcel()-checksum", "H")
		significant(c.peer.PeerIdent(), "Connection.isValidParcel(), failed due to bad checksum: %+v", parcel.Header)
		return InvalidPeerDemerit
	default:
		parcel.Trace("Connection.isValidParcel()-ParcelValid", "H")
		return ParcelValid
	}
}
func (c *Connection) handleParcelTypes(parcel Parcel) {
	switch parcel.Header.Type {
	case TypeAlert:
		parcel.Trace("Connection.handleParcelTypes()-TypeAlert", "J")
		significant(c.peer.PeerIdent(), "!!!!!!!!!!!!!!!!!! Alert: Alert feature not implemented.")
	case TypePing:
		// Send Pong
		parcel.Trace("Connection.handleParcelTypes()-TypePing", "J")
		pong := NewParcel(CurrentNetwork, []byte("Pong"))
		pong.Header.Type = TypePong
		debug(c.peer.PeerIdent(), "handleParcelTypes() GOT PING, Sending Pong: %s", pong.String())
		parcel.Print()
		BlockFreeChannelSend(c.SendChannel, ConnectionParcel{Parcel: *pong})
	case TypePong: // all we need is the timestamp which is set already
		parcel.Trace("Connection.handleParcelTypes()-TypePong", "J")

		debug(c.peer.PeerIdent(), "handleParcelTypes() GOT Pong.")
		return
	case TypePeerRequest:
		debug(c.peer.PeerIdent(), "handleParcelTypes() TypePeerRequest")
		parcel.Trace("Connection.handleParcelTypes()-TypePeerRequest", "J")

		BlockFreeChannelSend(c.ReceiveChannel, ConnectionParcel{Parcel: parcel}) // Controller handles these.
	case TypePeerResponse:
		parcel.Trace("Connection.handleParcelTypes()-TypePeerResponse", "J")

		debug(c.peer.PeerIdent(), "handleParcelTypes() TypePeerResponse")
		BlockFreeChannelSend(c.ReceiveChannel, ConnectionParcel{Parcel: parcel}) // Controller handles these.
	case TypeMessage:
		parcel.Trace("Connection.handleParcelTypes()-TypeMessage", "J")
		c.peer.QualityScore = c.peer.QualityScore + 1
		debug(c.peer.PeerIdent(), "handleParcelTypes() TypeMessage. Message is a: %s", parcel.MessageType())
		// Store our connection ID so the controller can direct response to us.
		parcel.Header.TargetPeer = c.peer.Hash
		parcel.Header.NodeID = NodeID
		BlockFreeChannelSend(c.ReceiveChannel, ConnectionParcel{Parcel: parcel}) // Controller handles these.
	case TypeMessagePart:
		parcel.Trace("Connection.handleParcelTypes()-TypeMessagePart", "J")
		debug(c.peer.PeerIdent(), "handleParcelTypes() TypeMessagePart. Message is a: %s", parcel.MessageType())
		c.peer.QualityScore = c.peer.QualityScore + 1
		debug(c.peer.PeerIdent(), "handleParcelTypes() TypeMessagePart. Message is a: %s", parcel.MessageType())
		// Store our connection ID so the controller can direct response to us.
		parcel.Header.TargetPeer = c.peer.Hash
		parcel.Header.NodeID = NodeID
		BlockFreeChannelSend(c.ReceiveChannel, ConnectionParcel{Parcel: parcel}) // Controller handles these.
	default:
		parcel.Trace("Connection.handleParcelTypes()-unknown", "J")
		significant(c.peer.PeerIdent(), "!!!!!!!!!!!!!!!!!! Got message of unknown type?")
		parcel.Print()
	}
}

func (c *Connection) pingPeer() {
	durationLastContact := time.Since(c.peer.LastContact)
	durationLastPing := time.Since(c.timeLastPing)
	if PingInterval < durationLastContact && PingInterval < durationLastPing {
		if MaxNumberOfRedialAttempts < c.attempts {
			note(c.peer.PeerIdent(), "pingPeer() GOING OFFLINE - No response to pings. Attempts: %d Ti  since last contact: %s and time since last ping: %s", PingInterval.String(), durationLastContact.String(), durationLastPing.String())
			c.goOffline()
			return
		} else {
			verbose(c.peer.PeerIdent(), "pingPeer() Connection State: %s", c.ConnectionState())
			note(c.peer.PeerIdent(), "pingPeer() Ping interval %s is less than duration since last contact: %s and time since last ping: %s", PingInterval.String(), durationLastContact.String(), durationLastPing.String())
			parcel := NewParcel(CurrentNetwork, []byte("Ping"))
			parcel.Header.Type = TypePing
			c.timeLastPing = time.Now()
			c.attempts++
			BlockFreeChannelSend(c.SendChannel, ConnectionParcel{Parcel: *parcel})
		}
	}
}

func (c *Connection) updatePeer() {
	verbose(c.peer.PeerIdent(), "updatePeer() SENDING ConnectionUpdatingPeer - Connection State: %s", c.ConnectionState())
	c.timeLastUpdate = time.Now()
	BlockFreeChannelSend(c.ReceiveChannel, ConnectionCommand{Command: ConnectionUpdatingPeer, Peer: c.peer})
}

func (c *Connection) updateStats() {
	if time.Second < time.Since(c.timeLastMetrics) {
		c.timeLastMetrics = time.Now()
		c.metrics.PeerAddress = c.peer.Address
		c.metrics.PeerQuality = c.peer.QualityScore
		c.metrics.ConnectionState = connectionStateStrings[c.state]
		c.metrics.ConnectionNotes = c.notes
		verbose(c.peer.PeerIdent(), "updatePeer() SENDING ConnectionUpdateMetrics - Bytes Sent: %d Bytes Received: %d", c.metrics.BytesSent, c.metrics.BytesReceived)
		BlockFreeChannelSend(c.ReceiveChannel, ConnectionCommand{Command: ConnectionUpdateMetrics, Metrics: c.metrics})
	}
}

func (c *Connection) ConnectionState() string {
	return connectionStateStrings[c.state]
}

func (c *Connection) connectionStatusReport() {
	reportDuration := time.Since(c.timeLastStatus)
	if reportDuration > ConnectionStatusInterval {
		c.timeLastStatus = time.Now()
		significant("connection-report", "\n\n===============================================================================\n     Connection: %s\n          State: %s\n          Notes: %s\n           Hash: %s\n     Persistent: %t\n       Outgoing: %t\n ReceiveChannel: %d\n    SendChannel: %d\n\tConnStatusInterval:\t%s\n\treportDuration:\t\t%s\n\tTime Online:\t\t%s \nMsgs/Bytes: %d / %d \n==============================================================================\n\n", c.peer.AddressPort(), c.ConnectionState(), c.Notes(), c.peer.Hash[0:12], c.IsPersistent(), c.IsOutGoing(), len(c.ReceiveChannel), len(c.SendChannel), ConnectionStatusInterval.String(), reportDuration.String(), time.Since(c.timeLastAttempt), c.metrics.MessagesReceived+c.metrics.MessagesSent, c.metrics.BytesSent+c.metrics.BytesReceived)
	}
}
