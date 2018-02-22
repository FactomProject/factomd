// Copyright 2017 Factom Foundation
// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"encoding/gob"
	"fmt"
	"hash/crc32"
	"net"
	"os"
	"time"

	"github.com/FactomProject/factomd/common/primitives"

	atomic2 "github.com/FactomProject/factomd/util/atomic"
	log "github.com/sirupsen/logrus"
)

// conLogger is the general logger for all connection related logs. You can add additional fields,
// or create more context loggers off of this
var conLogger = packageLogger.WithFields(log.Fields{"subpack": "connection"})

// Connection represents a single connection to another peer over the network. It communicates with the application
// via two channels, send and receive.  These channels take structs of type ConnectionCommand or ConnectionParcel
// (defined below).
type Connection struct {
	mu             atomic2.DebugMutex // lock this connection
	conn           net.Conn
	Errors         chan error              // handle errors from connections.
	Commands       chan *ConnectionCommand // handle connection commands
	SendChannel    chan interface{}        // Send means "towards the network" Channel sends Parcels and ConnectionCommands
	ReceiveChannel chan interface{}        // Receive means "from the network" Channel receives Parcels and ConnectionCommands
	ReceiveParcel  chan *Parcel            // Parcels to be handled.
	// and as "address" for sending messages to specific nodes.
	encoder         *gob.Encoder      // Wire format is gobs in this version, may switch to binary
	decoder         *gob.Decoder      // Wire format is gobs in this version, may switch to binary
	peer            Peer              // the data structure representing the peer we are talking to. defined in peer.go
	attempts        int               // reconnection attempts
	TimeLastPacket  time.Time         // Time we last successfully received a packet or command.
	timeLastAttempt time.Time         // time of last attempt to connect via dial
	timeLastPing    time.Time         // time of last ping sent
	timeLastUpdate  time.Time         // time of last peer update sent
	timeLastStatus  time.Time         // last time we printed our status for debugging.
	timeLastMetrics time.Time         // last time we updated metrics
	state           uint8             // Current state of the connection. Private. Only communication
	isOutGoing      bool              // We keep track of outgoing dial() vs incoming accept() connections
	isPersistent    bool              // Persistent connections we always redial.
	notes           string            // Notes about the connection, for debugging (eg: error)
	metrics         ConnectionMetrics // Metrics about this connection
	metricsMutex    atomic2.DebugMutex
	Logger          *log.Entry
}

// Each connection is a simple state machine.  The state is managed by a single goroutine which also does networking.
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
	ConnectionShuttingDown              // We're shutting down, the receives loop exits.
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

// ConnectionParcel is sent to convey an application message destined for the network.
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
	BytesSent        uint32    // Keeping track of the data sent/received for console
	BytesReceived    uint32    // Keeping track of the data sent/received for console
	MessagesSent     uint32    // Keeping track of the data sent/received for console
	MessagesReceived uint32    // Keeping track of the data sent/received for console
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

func (e *ConnectionCommand) JSONByte() ([]byte, error)   { return primitives.EncodeJSON(e) }
func (e *ConnectionCommand) JSONString() (string, error) { return primitives.EncodeJSONString(e) }
func (e *ConnectionCommand) String() string {
	str, _ := e.JSONString()
	return str
}

// These are the commands that connections can send/receive
const (
	ConnectionIsClosed uint8 = iota // Notifies the controller that we are shut down and can be released
	ConnectionShutdownNow
	ConnectionUpdatingPeer
	ConnectionAdjustPeerQuality
	ConnectionUpdateMetrics
	ConnectionGoOffline // Notifies the connection it should go offline (eg from another goroutine)
)

//////////////////////////////
//
// Public API
//
//////////////////////////////

// InitWithConn is called from our accept loop when a peer dials into us and we already have a network conn
func (c *Connection) InitWithConn(conn net.Conn, peer Peer) *Connection {
	c.mu.Lock()
	c.conn = conn
	c.isOutGoing = false
	c.isPersistent = false
	c.mu.Unlock()
	c.commonInit(peer)
	c.goOnline()
	return c
}

// Init is called when we have peer info and need to dial into the peer
func (c *Connection) Init(peer Peer, persistent bool) *Connection {
	c.mu.Lock()
	c.conn = nil
	c.isOutGoing = true
	c.isPersistent = persistent
	c.mu.Unlock()
	c.commonInit(peer)
	return c
}

func (c *Connection) QualityScore() int32 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.peer.QualityScore
}

func (c *Connection) State() uint8 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

func (c *Connection) IsOutGoing() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isOutGoing
}

func (c *Connection) IsOnline() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return ConnectionOnline == c.state
}

func (c *Connection) StatusString() string {
	c.mu.Lock()
	s := c.state
	c.mu.Unlock()

	switch s {
	case ConnectionInitialized:
		return "Initialized"
	case ConnectionOnline:
		return "Online"
	case ConnectionOffline:
		return "Offline"
	case ConnectionShuttingDown:
		return "Shutting Down"
	case ConnectionClosed:
		return "Closed"
	default:
		return fmt.Sprintf("UnknownConnState%x", s)
	}
}

func (c *Connection) IsPersistent() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isPersistent
}
func (c *Connection) Notes() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.notes
}

//////////////////////////////
//
// Private API
//
//////////////////////////////

func (c *Connection) commonInit(peer Peer) {
	p2pConnectionCommonInit.Inc() // Prometheus
	c.mu.Lock()
	c.state = ConnectionInitialized
	c.peer = peer
	c.Errors = make(chan error, StandardChannelSize)
	c.Commands = make(chan *ConnectionCommand, StandardChannelSize)
	c.SendChannel = make(chan interface{}, StandardChannelSize)
	c.ReceiveChannel = make(chan interface{}, StandardChannelSize)
	c.ReceiveParcel = make(chan *Parcel, StandardChannelSize)
	now := time.Now()
	c.timeLastMetrics = now
	c.timeLastAttempt = now
	c.timeLastStatus = now
	c.Logger = conLogger.WithField("peer", c.peer.PeerFixedIdent())
	c.mu.Unlock()
	c.metricsMutex.Lock()
	c.metrics = ConnectionMetrics{MomentConnected: now} // L
	c.metricsMutex.Unlock()
	c.setNotes("commonInit()") // needs to be unlocked before setNotes()
}

func (c *Connection) Start() {
	go c.runLoop()
}

// runloop OWNs the connection.  It is the only goroutine that can change values in the connection struct
// runLoop operates the state machine and routes messages out to network (messages from network are routed in processReceives)
func (c *Connection) runLoop() {
	go c.processSends()
	go c.processReceives()
	p2pConnectionsRunLoop.Inc()
	defer p2pConnectionsRunLoop.Dec()

	for ConnectionClosed != c.State() { // loop exits when we hit shutdown state
		time.Sleep(100 * time.Millisecond)
		// time.Sleep(time.Second * 1) // This can be a tight loop, don't want to starve the application
		c.updateStats() // Update controller with metrics
		c.connectionStatusReport()
		// if 2 == rand.Intn(100) {
		debug(c.peer.PeerFixedIdent(), "Connection.runloop() STATE IS: %s", c.StatusString())
		// }
		c.handleNetErrors(false)
		c.handleCommand()

	parcelloop:
		for {
			select {
			case m := <-c.ReceiveParcel:
				c.mu.Lock()
				c.TimeLastPacket = time.Now()
				c.mu.Unlock()
				c.handleParcel(*m)

			default:
				break parcelloop
			}
		}

		switch c.State() {
		case ConnectionInitialized:
			p2pConnectionRunLoopInitalized.Inc()
			if MinimumQualityScore > c.QualityScore() && !c.IsPersistent() {
				c.updatePeer() // every PeerSaveInterval * 0.90 we send an update peer to the controller.
				c.goShutdown()
			} else {
				c.dialLoop() // dialLoop dials until it connects or shuts down.
			}
		case ConnectionOnline:
			p2pConnectionRunLoopOnline.Inc()
			c.pingPeer() // sends a ping periodically if things have been quiet
			if PeerSaveInterval < time.Since(c.timeLastUpdate) {
				c.updatePeer() // every PeerSaveInterval * 0.90 we send an update peer to the controller.
			}

			if MinimumQualityScore > c.QualityScore() && !c.IsPersistent() {
				c.updatePeer() // every PeerSaveInterval * 0.90 we send an update peer to the controller.
				c.goShutdown()
			}

		case ConnectionOffline:
			p2pConnectionRunLoopOffline.Inc()
			switch {
			case c.isOutGoing:
				c.dialLoop() // dialLoop dials until it connects or shuts down.
			default: // the connection dialed us, so we shutdown
				c.goShutdown()
			}
		case ConnectionShuttingDown:
			p2pConnectionRunLoopShutdown.Inc()
			c.mu.Lock()
			c.state = ConnectionClosed
			c.mu.Unlock()
			BlockFreeChannelSend(c.ReceiveChannel, ConnectionCommand{Command: ConnectionIsClosed})
			return // ending runloop() goroutine
		default:
			logfatal(c.peer.PeerIdent(), "runLoop() unknown state: %s ", c.StatusString())
		}
	}
	significant(c.peer.PeerIdent(), "runLoop() Connection runloop() exiting %+v", c)
}

// assumes it is called with c.mu already locked
func (c *Connection) setNotes(format string, v ...interface{}) {
	note := fmt.Sprintf(format, v...)
	c.mu.Lock()
	c.notes = note
	c.mu.Unlock()
	significant(c.peer.PeerIdent(), c.notes)

}

// dialLoop:  dials the connection until giving up. Called in offline or initializing states.
// All exits from dialLoop change the state of the connection allowing the outside run_loop to proceed.
func (c *Connection) dialLoop() {
	p2pConnectionDialLoop.Inc()
	defer p2pConnectionDialLoop.Dec()

	for {
		c.timeLastAttempt = time.Now()
		if c.dial() {
			c.goOnline()
			return
		}
		switch {
		case c.isPersistent:
		case ConnectionOffline == c.State(): // We were online with the peer at one point.
			c.attempts++
			if MaxNumberOfRedialAttempts < c.attempts {
				c.goShutdown()
				return
			}
		default:
			c.goShutdown()
			return
		}
		time.Sleep(TimeBetweenRedials)
	}
}

// dial() handles connection logic and shifts states based on results.
func (c *Connection) dial() bool {
	address := c.peer.AddressPort()
	// conn, err := net.Dial("tcp", c.peer.Address)
	conn, err := net.DialTimeout("tcp", address, time.Second*10)
	if nil == err {
		c.mu.Lock()
		c.conn = conn
		c.mu.Unlock()
		return true
	}
	return false
}

// Called when we are online and connected to the peer.
func (c *Connection) goOnline() {
	p2pConnectionOnlineCall.Inc()
	now := time.Now()
	c.mu.Lock()
	c.encoder = gob.NewEncoder(c.conn)
	c.decoder = gob.NewDecoder(c.conn)
	c.attempts = 0
	c.timeLastPing = now
	c.timeLastAttempt = now
	c.timeLastUpdate = now
	c.peer.LastContact = now
	c.state = ConnectionOnline
	c.mu.Unlock()
	// Drain the handleNetErrors to avoid immediate disconnect
	c.handleNetErrors(true)
	// Probably shouldn't reset metrics when we go online. (Eg: say after a temp network problem)
	// c.metrics = ConnectionMetrics{MomentConnected: now} // Reset metrics
	// Now ask the other side for the peers they know about.
	parcel := NewParcel(CurrentNetwork, []byte("Peer Request"))
	parcel.Header.Type = TypePeerRequest
	BlockFreeChannelSend(c.SendChannel, ConnectionParcel{Parcel: *parcel})
}

func (c *Connection) goOffline() {
	p2pConnectionOfflineCall.Inc()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state = ConnectionOffline
	c.attempts = 0
	c.peer.demerit()
}

func (c *Connection) goShutdown() {
	c.goOffline()
	c.updatePeer()
	c.mu.Lock()
	defer c.mu.Unlock()
	if nil != c.conn {
		defer c.conn.Close()
	}
	c.decoder = nil
	c.encoder = nil
	c.state = ConnectionShuttingDown

}

// processSends gets all the messages from the application and sends them out over the network
func (c *Connection) processSends() {
	p2pProcessSendsGuage.Inc()
	defer p2pProcessSendsGuage.Dec()

	defer func() {
		if r := recover(); r != nil {
			// Just ignore the possible nil pointer error that can occur because
			// we have cleared the pointer to the encoder or decoder outside this
			// go routine.
		}
	}()

	c.mu.Lock()                  // lock so I can read the state
	sendChannel := c.SendChannel // L
	commands := c.Commands       // L
	errors := c.Errors
	c.mu.Unlock()
	for ConnectionClosed != c.State() && c.State() != ConnectionShuttingDown {
		// note(c.peer.PeerIdent(), "Connection.processSends() called. Items in send channel: %d State: %s", len(c.SendChannel), c.ConnectionState())
		c.mu.Lock()          // lock so I can read the state
		encoder := c.encoder // L
		conn := c.conn       // L
		c.mu.Unlock()
		if nil != encoder && nil != conn { // Skip this if the connection is not generally valid
			for ConnectionOnline == c.State() && len(sendChannel) > 0 {
				// This was blocking. By checking the length of the channel before entering, this does not block.
				// The problem was this routine was blocked on a closed connection. Ideally we do want to block
				// on a 0 length channel, and this is still possible if use a select and close the channel when we
				// close the connection.
				message := <-sendChannel
				switch message.(type) {
				case ConnectionParcel:
					parameters := message.(ConnectionParcel)
					c.sendParcel(parameters.Parcel, encoder, conn, errors)
				case ConnectionCommand:
					parameters := message.(ConnectionCommand)
					commands <- &parameters
				default:
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (c *Connection) sendParcel(parcel Parcel, encoder *gob.Encoder, conn net.Conn, errors chan error) {

	parcel.Header.NodeID = NodeID // Send it out with our ID for loopback detection.

	conn.SetWriteDeadline(time.Now().Add(NetworkDeadline * 500)) // So really no deadline ....
	err := encoder.Encode(parcel)
	switch {
	case nil == err:
		c.metricsMutex.Lock()
		c.metrics.BytesSent += parcel.Header.Length //L
		c.metrics.MessagesSent += 1                 //L
		c.metricsMutex.Unlock()
	default:
		errors <- err
	}
}

// processReceives is a go routine This is essentially an infinite loop that exits
// when:
// -- a network error happens
// -- something causes our state to be offline
func (c *Connection) processReceives() {
	p2pProcessReceivesGuage.Inc()
	defer p2pProcessReceivesGuage.Dec()

	defer func() {
		if r := recover(); r != nil {
			// Just ignore the possible nil pointer error that can occur because
			// we have cleared the pointer to the encoder or decoder outside this
			// go routine.
		}
	}()

	for ConnectionClosed != c.State() && c.State() != ConnectionShuttingDown {
		for c.State() == ConnectionOnline {
			var message Parcel

			c.mu.Lock()
			// c.conn.SetReadDeadline(time.Now().Add(NetworkDeadline))
			decoder := c.decoder
			receiveParcel := c.ReceiveParcel
			c.mu.Unlock()
			if decoder != nil {
				err := decoder.Decode(&message)
				switch {
				case nil == err:
					now := time.Now() // never do now() in a locked context. It can cause context switches and be gone for a long time (>3s)
					c.metricsMutex.Lock()
					c.metrics.BytesReceived += message.Header.Length //L metrics
					c.metrics.MessagesReceived += 1                  //L metrics
					c.metricsMutex.Unlock()
					c.mu.Lock()
					message.Header.PeerAddress = c.peer.Address // L
					c.TimeLastPacket = now                      // L
					c.mu.Unlock()
					receiveParcel <- &message // U
				default:
					c.mu.Lock()
					errorsChan := c.Errors
					c.mu.Unlock()
					errorsChan <- err

				}
			}
		}
		// If not online, give some time up to handle states that are not online, closed, or shuttingdown.
		time.Sleep(1 * time.Second) // clay -- seems a bit long
	}
}

func (c *Connection) handleCommand() {
	select {
	case command := <-c.Commands:

		switch command.Command {
		case ConnectionShutdownNow:
			c.setNotes("Connection(%s) shutting down due to ConnectionShutdownNow message.", c.peer.AddressPort())
			c.goShutdown()
		case ConnectionUpdatingPeer: // at this level we're only updating the quality score, to pass on application level demerits
			debug(c.peer.PeerIdent(), "handleCommand() ConnectionUpdatingPeer")
			peer := command.Peer
			if peer.QualityScore < c.QualityScore() {
				c.mu.Lock()
				c.peer.QualityScore = peer.QualityScore
				c.mu.Unlock()
			}
		case ConnectionAdjustPeerQuality:
			delta := command.Delta
			note(c.peer.PeerIdent(), "handleCommand() ConnectionAdjustPeerQuality: Current Score: %d Delta: %d", c.peer.QualityScore, delta)
			c.peer.QualityScore = c.peer.QualityScore + delta
			if MinimumQualityScore > c.QualityScore() {
				debug(c.peer.PeerIdent(), "handleCommand() disconnecting peer: %s for quality score: %d", c.peer.PeerIdent(), c.QualityScore())
				c.updatePeer()
				c.setNotes("Connection(%s) shutting down due to QualityScore %d being below MinimumQualityScore: %d.",
					c.peer.AddressPort(), c.QualityScore(), MinimumQualityScore)
				c.goShutdown()
			}
		case ConnectionGoOffline:
			debug(c.peer.PeerIdent(), "handleCommand() disconnecting peer: %s goOffline command received", c.peer.PeerIdent())
			c.goOffline()
		default:
			logfatal(c.peer.PeerIdent(), "handleCommand() unknown command?: %+v ", command)
		}
	default:
	}
}

//handleNetErrors Reacts to errors we get from encoder or decoder
func (c *Connection) handleNetErrors(toss bool) {
	done := false
	c.mu.Lock()
	for {
		select {
		case err := <-c.Errors:
			nerr, isNetError := err.(net.Error)
			switch {
			case isNetError && nerr.Timeout(): /// buffer empty
				c.mu.Unlock()
				return
			default:
				// Only go offline once per handleNetErrors call
				if !toss && !done {
					c.mu.Unlock()
					if err != nil {
						c.Logger.WithField("func", "HandleNetErrors").Errorf("Going offline due to -- %s", err.Error())
					}
					c.goOffline()
					c.mu.Lock()
				}

				done = true
			}
		default:
			c.mu.Unlock()
			return
		}
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
		c.setNotes("Connection(%s) shutting down due to InvalidDisconnectPeer result from parcel. Previous notes: %s.", c.peer.AddressPort(), c.notes)
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
		if Notes <= CurrentLoggingLevel() {
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
		c.setNotes("Connection.isValidParcel(), failed due to loopback!: %+v", parcel.Header)
		c.peer.QualityScore = MinimumQualityScore - 50 // Ban ourselves for a week
		return InvalidDisconnectPeer
	case parcel.Header.Network != CurrentNetwork:
		parcel.Trace("Connection.isValidParcel()-network", "H")
		c.setNotes("Connection.isValidParcel(), failed due to wrong network. Remote: %0x Us: %0x", parcel.Header.Network, CurrentNetwork)
		return InvalidDisconnectPeer
	case parcel.Header.Version < ProtocolVersionMinimum:
		parcel.Trace("Connection.isValidParcel()-version", "H")
		c.setNotes("Connection.isValidParcel(), failed due to wrong version: %+v", parcel.Header)
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
		significant(c.peer.PeerIdent(), "!!!!!!!!!!!!!!!!!! Alert: Alert feature not implemented.")
	case TypePing:
		// Send Pong
		pong := NewParcel(CurrentNetwork, []byte("Pong"))
		pong.Header.Type = TypePong
		BlockFreeChannelSend(c.SendChannel, ConnectionParcel{Parcel: *pong})
	case TypePong: // all we need is the timestamp which is set already
		return
	case TypePeerRequest:
		BlockFreeChannelSend(c.ReceiveChannel, ConnectionParcel{Parcel: parcel}) // Controller handles these.
	case TypePeerResponse:
		BlockFreeChannelSend(c.ReceiveChannel, ConnectionParcel{Parcel: parcel}) // Controller handles these.
	case TypeMessage:
		c.peer.QualityScore = c.peer.QualityScore + 1
		// Store our connection ID so the controller can direct response to us.
		parcel.Header.TargetPeer = c.peer.Hash
		parcel.Header.NodeID = NodeID
		BlockFreeChannelSend(c.ReceiveChannel, ConnectionParcel{Parcel: parcel}) // Controller handles these.
	case TypeMessagePart:
		c.peer.QualityScore = c.peer.QualityScore + 1
		// Store our connection ID so the controller can direct response to us.
		parcel.Header.TargetPeer = c.peer.Hash
		parcel.Header.NodeID = NodeID
		BlockFreeChannelSend(c.ReceiveChannel, ConnectionParcel{Parcel: parcel}) // Controller handles these.
	default:
		significant(c.peer.PeerIdent(), "!!!!!!!!!!!!!!!!!! Got message of unknown type?")
	}
}

func (c *Connection) pingPeer() {
	durationLastContact := time.Since(c.peer.LastContact)
	durationLastPing := time.Since(c.timeLastPing)
	if PingInterval < durationLastContact && PingInterval < durationLastPing {
		if MaxNumberOfRedialAttempts < c.attempts {
			c.goOffline()
			return
		} else {
			parcel := NewParcel(CurrentNetwork, []byte("Ping"))
			parcel.Header.Type = TypePing
			c.timeLastPing = time.Now()
			c.attempts++
			BlockFreeChannelSend(c.SendChannel, ConnectionParcel{Parcel: *parcel})
		}
	}
}

func (c *Connection) updatePeer() {
	now := time.Now() // never do now() in a locked context. It can cause context switches and be gone for a long time (>3s)
	c.mu.Lock()
	c.timeLastUpdate = now
	rc := c.ReceiveChannel
	peer := c.peer
	c.mu.Unlock()
	BlockFreeChannelSend(rc, ConnectionCommand{Command: ConnectionUpdatingPeer, Peer: peer})
}

func (c *Connection) updateStats() {
	c.metricsMutex.Lock()
	lm := c.timeLastMetrics // L metricsMutex
	c.metricsMutex.Unlock()
	if time.Second < time.Since(lm) {
		now := time.Now()
		statusString := c.StatusString()
		notes := c.Notes()

		c.metricsMutex.Lock()
		c.timeLastMetrics = now                                // L metricsMutex
		c.metrics.PeerAddress = c.peer.Address                 // L metricsMutex
		c.metrics.PeerQuality = c.peer.QualityScore            // L metricsMutex
		c.metrics.ConnectionState = statusString               // L metricsMutex
		c.metrics.ConnectionNotes = notes                      // L metricsMutex
		bs, br := c.metrics.BytesSent, c.metrics.BytesReceived // L metricsMutex
		m := c.metrics                                         // L metricsMutex Copy the metrics
		c.metricsMutex.Unlock()
		BlockFreeChannelSend(c.ReceiveChannel, ConnectionCommand{Command: ConnectionUpdateMetrics, Metrics: m}) // L metricsMutex
		verbose(c.peer.PeerIdent(), "updatePeer() SENDING ConnectionUpdateMetrics - Bytes Sent: %d Bytes Received: %d", bs, br)
	}
}

func (c *Connection) connectionStatusReport() {
	c.mu.Lock()
	reportDuration := time.Since(c.timeLastStatus)
	c.mu.Unlock()

	if reportDuration > ConnectionStatusInterval {
		c.metricsMutex.Lock()
		metrics := c.metrics // copy the metrics
		c.metricsMutex.Unlock()
		c.mu.Lock()
		peerHash := c.peer.Hash
		rl := len(c.ReceiveChannel) // snapshot in time
		sl := len(c.SendChannel)    // snapshot in time
		timeLastAttempt := c.timeLastAttempt
		c.mu.Unlock()

		isPersistent := c.IsPersistent()
		isOutgoing := c.IsOutGoing()
		notes := c.Notes()
		address := c.peer.AddressPort()
		status := c.StatusString()

		// TODO: No need to do locking on the copy ... But the copy has it's own locks so it just wasted time ... -- clay
		cm := metrics.MessagesReceived + metrics.MessagesSent
		bm := metrics.BytesSent + metrics.BytesReceived

		// TODO: Really should segregate this into static info (never changes once the connection is created) and dynamic and do the static when the connection is created ... -- clay
		significant("connection-report",
			"\n\n===============================================================================\n     Connection: %s\n          State: %s\n          Notes: %s\n           Hash: %s\n     Persistent: %t\n       Outgoing: %t\n ReceiveChannel: %d\n    SendChannel: %d\n\tConnStatusInterval:\t%s\n\treportDuration:\t\t%s\n\tTime Online:\t\t%s \nMsgs/Bytes: %d / %d \n==============================================================================\n\n",
			address, status, notes, peerHash, isPersistent, isOutgoing, rl, sl, ConnectionStatusInterval.String(), reportDuration.String(), time.Since(timeLastAttempt), cm, bm)
	}
}
