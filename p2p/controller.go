// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

// Controller manages the P2P Network.
// It maintains the list of peers, and has a master run-loop that
// processes ingoing and outgoing messages.
// It is controlled via a command channel.
// Other than Init and NetworkStart, all administration is done via the channel.

import (
	"fmt"
	"net"
	"time"
)

// Controller manages the peer to peer network.
type Controller struct {
	keepRunning bool // Indicates its time to shut down when false.

	// After launching the network, the management is done via these channels.
	commandChannel chan interface{} // Application use controller public API to send commands on this channel to controllers goroutines.

	ToNetwork   chan Parcel // Parcels from the application for us to route
	FromNetwork chan Parcel // Parcels from the network for the application

	listenPort  string                // port we listen on for new connections
	connections map[string]Connection // map of the peers indexed by peer hash
	discovery   Discovery             // Our discovery structure
}

// CommandDialPeer is used to instruct the Controller to dial a peer address
type CommandDialPeer struct {
	address string
}

// CommandAddPeer is used to instruct the Controller to add a connection
// This connection can come from acceptLoop or some other way.
type CommandAddPeer struct {
	connection Connection
}

// CommandShutdown is used to instruct the Controller to takve various actions.
type CommandShutdown struct {
	_ uint8
}

// CommandDemerit is used to instruct the Controller to reduce a connections quality score
type CommandDemerit struct {
	peerHash string
}

// CommandMerit is used to instruct the Controller to increase a connections quality score
type CommandMerit struct {
	peerHash string
}

// CommandBan is used to instruct the Controller to disconnect and ban a peer
type CommandBan struct {
	peerHash string
}

// CommandChangeLogging is used to instruct the Controller to takve various actions.
type CommandChangeLogging struct {
	level uint8
}

//////////////////////////////////////////////////////////////////////
// Public (exported) methods.
//
// The surface for interfacting with this is very minimal to avoid deadlocks
// and allow maximum concurrency.
// Other than setup, these API communicate with the controller via the
// command channel.
//////////////////////////////////////////////////////////////////////

func (c *Controller) Init(port string, peersFile string) *Controller {
	verbose("controller", "Controller.Init(%s)", port)
	c.keepRunning = true
	c.commandChannel = make(chan interface{}, 1000) // Commands from App
	c.FromNetwork = make(chan Parcel, 1000)         // Channel to the app for network data
	c.ToNetwork = make(chan Parcel, 1000)           // Parcels from the app for the network
	c.listenPort = port
	c.connections = make(map[string]Connection)
	discovery := new(Discovery).Init(peersFile)
	c.discovery = *discovery
	return c
}

// StartNetwork configures the network, starts the runloop
func (c *Controller) StartNetwork(exclusive bool) {
	verbose("controller", "Controller.StartNetwork(%s)", " ")
	// exclusive means we only connect to special peers
	OnlySpecialPeers = exclusive
	// start listening on port given
	c.listen()
	// Start the discovery service?
	c.discovery.LoadPeers()
	// Get a list of peers from discovery
	peers := c.discovery.GetStartupPeers()
	// dial into the peers
	for _, peer := range peers {
		c.DialPeer(peer.Address)
	}
	/// start heartbeat process
	// Start the runloop
	go c.runloop()
}

func (c *Controller) StartLogging(level uint8) {
	note("controller", "Changing log level to %s", LoggingLevels[level])
	c.commandChannel <- CommandChangeLogging{level: level}
}
func (c *Controller) StopLogging() {
	note("controller", "Changing log level to %s", LoggingLevels[Silence])
	c.commandChannel <- CommandChangeLogging{level: Silence}
}
func (c *Controller) ChangeLogLevel(level uint8) {
	note("controller", "Changing log level to %s", LoggingLevels[level])
	c.commandChannel <- CommandChangeLogging{level: level}
}

func (c *Controller) DialPeer(address string) {
	debug("controller", "DialPeer message for %s", address)
	c.commandChannel <- CommandDialPeer{address: address}
}

func (c *Controller) AddPeer(connection *Connection) {
	debug("controller", "CommandAddPeer for %+v", connection)
	c.commandChannel <- CommandAddPeer{connection: *connection}
}

func (c *Controller) NetworkStop() {
	debug("controller", "NetworkStop ")
	c.commandChannel <- CommandShutdown{}
}

func (c *Controller) Demerit(peerHash string) {
	debug("controller", "NetworkStop ")
	c.commandChannel <- CommandDemerit{peerHash: peerHash}
}

func (c *Controller) Merit(peerHash string) {
	debug("controller", "NetworkStop ")
	c.commandChannel <- CommandMerit{peerHash: peerHash}
}

func (c *Controller) Ban(peerHash string) {
	debug("controller", "NetworkStop ")
	c.commandChannel <- CommandBan{peerHash: peerHash}
}

//////////////////////////////////////////////////////////////////////
//
// Private API (unexported)
//
//  These functions happen in the runloop goroutine, and deal with the
//  outside world via the command channels and the connections.
//
//////////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////////
// Network management
//////////////////////////////////////////////////////////////////////

func (c *Controller) listen() {
	fmt.Printf("Controller.listen() %+v\n", " DEBUG statement immediately follows!")
	address := fmt.Sprintf(":%s", c.listenPort)
	note("controller", "Controller.listen(%s) got address %s", c.listenPort, address)
	listener, err := net.Listen("tcp", address)
	if nil != err {
		logfatal("controller", "Controller.listen() Error: %+v", err)
	}
	go c.acceptLoop(listener)
}

// Since this runs in its own goroutine we need to send a command when
// when we get a new connection.
func (c *Controller) acceptLoop(listener net.Listener) {
	note("controller", "Controller.acceptLoop() starting up")
	for {
		conn, err := listener.Accept()
		if nil != err {
			logerror("controller", "Controller.acceptLoop() Error: %+v", err)
		} else {
			address := conn.RemoteAddr().String()
			peer := c.discovery.GetPeerByAddress(address)
			connection := new(Connection).Init(peer)
			connection.Configure(conn)
			c.AddPeer(connection) // Sends command to add the peer to the peers list
			note("controller", "Controller.acceptLoop() new peer: %+v", peer.Address)
		}
	}
}

//////////////////////////////////////////////////////////////////////
// Operations
//////////////////////////////////////////////////////////////////////

// runloop is a goroutine that does all the heavy lifting
func (c *Controller) runloop() {
	note("controller", "Controller.runloop() starting up")

	for c.keepRunning { // Run until we get the exit command
		time.Sleep(time.Millisecond * 100)

		// Process commands...
		verbose("controller", "Controller.runloop() About to process commands. Commands in channel: %d", len(c.commandChannel))
		for 0 < len(c.commandChannel) {
			command := <-c.commandChannel
			verbose("controller", "Controller.runloop() Processing command: %+v", command)
			c.handleCommand(command)
		}
		// route messages to and from application
		verbose("controller", "Controller.runloop() Calling router")
		c.route() // Route messages
		// Manage peers (reconnect, etc.)
		c.managePeers()
	}
	note("controller", "Controller.runloop() has exited. Shutdown command recieved?")
}

// Route pulls all of the messages from the application and sends them to the appropriate
// peer. Broadcast messages go to everyone, directed messages go to the named peer.
// route also passes incomming messages on to the application.
func (c *Controller) route() {
	verbose("controller", "Controller.route() called. Number peers: %d", len(c.connections))

	// Recieve messages from the peers & forward to application.
	for id, peer := range c.connections {
		// Empty the recieve channel, stuff the application channel.
		verbose("controller", "Controller.route() size of recieve channel: %d", len(peer.ReceiveChannel))
		for 0 < len(peer.ReceiveChannel) { // effectively "While there are messages"
			parcel := <-peer.ReceiveChannel
			verbose("controller", "Controller.route() got parcel from NETWORK %+v", parcel)
			parcel.Header.TargetPeer = id // Set the connection ID so the application knows which peer the message is from.
			switch parcel.Header.Type {
			case TypeMessage: // Application message, send it on.
				c.FromNetwork <- parcel
			case TypePeerRequest:
				// Get selection of peers from discovery
				response := NewParcel(CurrentNetwork, c.discovery.SharePeers())
				response.Header.Type = TypePeerResponse
				// Send them out to the network - on the connection that requested it!
				peer.SendChannel <- *response
			case TypePeerResponse:
				// Add these peers to our known peers
				c.discovery.LearnPeers(parcel.Payload)
			}
		}
	}
	// For each message, see if it is directed, if so, send to the
	// specific peer, otherwise, broadcast.
	verbose("controller", "Controller.route() size of ToNetwork channel: %d", len(c.ToNetwork))
	for 0 < len(c.ToNetwork) { // effectively "While there are messages"
		parcel := <-c.ToNetwork
		verbose("controller", "Controller.route() got parcel from APPLICATION %+v", parcel)
		if "" != parcel.Header.TargetPeer { // directed send
			verbose("controller", "Controller.route() Directed send to %+v", parcel.Header.TargetPeer)
			connection, present := c.connections[parcel.Header.TargetPeer]
			if present { // We're still connected to the target
				connection.SendChannel <- parcel
			}
		} else { // broadcast
			verbose("controller", "Controller.route() Boadcast send to %d peers", len(c.connections))
			for _, connection := range c.connections {
				verbose("controller", "Controller.route() Send to peer %s ", connection.peer.Hash)
				connection.SendChannel <- parcel
			}
		}
	}
}

func (c *Controller) handleCommand(command interface{}) {
	switch commandType := command.(type) {
	case CommandDialPeer: // parameter is the peer address
		parameters := command.(CommandDialPeer)
		peer := c.discovery.GetPeerByAddress(parameters.address)
		_, present := c.connections[peer.Hash]
		if !present { // we are not connected to the peer
			conn := new(Connection).Init(peer)
			connection := *conn
			connection.dial()
			c.connections[peer.Hash] = connection
			debug("controller", "Controller.handleCommand(CommandDialPeer) got peer %s", peer.Address)
		} else {
			debug("controller", "Controller.handleCommand(CommandDialPeer) ALREADY CONNECTED TO PEER %s", peer.Address)
		}
	case CommandAddPeer: // parameter is a Connection
		parameters := command.(CommandAddPeer)
		connection := parameters.connection
		c.connections[connection.peer.Hash] = connection
		debug("controller", "Controller.handleCommand(CommandAddPeer) got peer %+v", parameters.connection)
	case CommandShutdown:
		c.shutdown()
		debug("controller", "Controller.handleCommand(CommandAddPeer) ")
	case CommandChangeLogging:
		parameters := command.(CommandChangeLogging)
		CurrentLoggingLevel = parameters.level
		debug("controller", "Controller.handleCommand(CommandChangeLogging) new logging level %s", LoggingLevels[parameters.level])
	case CommandDemerit:
		parameters := command.(CommandDemerit)
		peerHash := parameters.peerHash
		connection := c.connections[peerHash]
		connection.peer.demerit()
		c.discovery.UpdatePeer(connection.peer)
	case CommandMerit:
		parameters := command.(CommandDemerit)
		peerHash := parameters.peerHash
		connection := c.connections[peerHash]
		connection.peer.merit()
		c.discovery.UpdatePeer(connection.peer)
	case CommandBan:
		parameters := command.(CommandDemerit)
		peerHash := parameters.peerHash
		connection := c.connections[peerHash]
		connection.peer.QualityScore = BannedQualityScore
		connection.connectionDropped() // hang up on the peer
		c.discovery.UpdatePeer(connection.peer)
	default:
		note("controller", "Unkown p2p.Controller command recieved: %+v", commandType)
	}
}

func (c *Controller) managePeers() {
	var duration time.Duration
	// check for and remove disconnected peers or peers offline after awhile
	for key, connection := range c.connections {
		if false == connection.Online {
			duration := time.Now().Sub(connection.timeLastAttempt)
			if MaxNumberOfRedialAttempts > connection.attempts && TimeBetweenRedials < duration {
				connection.dial()
			}
			if MaxNumberOfRedialAttempts <= connection.attempts { // give up on the connection
				connection.shutdown()
				delete(c.connections, key)
			}
		}
		// If it's been more than PingInterval since we last heard from a connection, send them a ping
		duration = time.Now().Sub(connection.timeLastContact)
		if PingInterval < duration {
			parcel := NewParcel(CurrentNetwork, []byte("Ping"))
			parcel.Header.Type = TypePing
			connection.SendChannel <- *parcel
		}
		// Go thru an update peers in discovery using discovery.UpdatePeer()
		// so the known peers list is kept relatively up to date with peer score.
		c.discovery.UpdatePeer(connection.peer)
	}
	duration = time.Now().Sub(c.discovery.lastPeerSave)
	// Every so often, tell the discovery service to save peers.
	if PeerSaveInterval < duration {
		c.discovery.SavePeers()
	}

}

func (c *Controller) shutdown() {
	debug("controller", "Controller.shutdown() ")
	// Go thru peer list and shut down connections.
	for key, connection := range c.connections {
		connection.shutdown()
		delete(c.connections, key)
	}
	c.keepRunning = false
}
