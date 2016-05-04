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
)

// Controller manages the peer to peer network.
type Controller struct {
	keepRunning bool // Indicates its time to shut down when false.

	// After launching the network, the management is done via these channels.
	commandChannel chan interface{} // Application use controller public API to send commands on this channel to controllers goroutines.

	ToNetwork   chan Parcel // Parcels from the application for us to route
	FromNetwork chan Parcel // Parcels from the network for the application

	listenPort string                // port we listen on for new connections
	peers      map[uint64]Connection // map of the peers indexed by peer id
}

// CommandDialPeer are used to instruct the Controller to takve various actions.
type CommandDialPeer struct {
	Address string
}

// CommandAddPeer are used to instruct the Controller to takve various actions.
type CommandAddPeer struct {
	Peer Connection
}

// CommandShutdown are used to instruct the Controller to takve various actions.
type CommandShutdown struct {
	_ uint8
}

// CommandChangeLogging are used to instruct the Controller to takve various actions.
type CommandChangeLogging struct {
	Level uint8
}

//////////////////////////////////////////////////////////////////////
// Public (exported) methods.
//
// The surface for interfacting with this is very minimal to avoid deadlocks
// and allow maximum concurrency.
// Other than setup, these API communicate with the controller via the
// command channel.
//////////////////////////////////////////////////////////////////////

func (c *Controller) Init(port string) *Controller {
	verbose(true, "Controller.Init(%s)", port)
	c.keepRunning = true
	c.commandChannel = make(chan interface{}, 1000) // Commands from App
	c.FromNetwork = make(chan Parcel, 1000)         // Channel to the app for network data
	c.ToNetwork = make(chan Parcel, 1000)           // Parcels from the app for the network
	c.listenPort = port
	c.peers = make(map[uint64]Connection)
	return c
}

// StartNetwork configures the network, starts the runloop
func (c *Controller) StartNetwork() {
	verbose(true, "Controller.StartNetwork(%s)", " ")
	// start listening on port given
	c.listen()
	// dial into the peers
	/// start heartbeat process
	// Start the runloop
	go c.runloop()
}

func (c *Controller) StartLogging(level uint8) {
	note(true, "Changing log level to %s", LoggingLevels[level])
	c.commandChannel <- CommandChangeLogging{Level: level}
}
func (c *Controller) StopLogging() {
	note(true, "Changing log level to %s", LoggingLevels[Silence])
	c.commandChannel <- CommandChangeLogging{Level: Silence}
}
func (c *Controller) ChangeLogLevel(level uint8) {
	note(true, "Changing log level to %s", LoggingLevels[level])
	c.commandChannel <- CommandChangeLogging{Level: level}
}

func (c *Controller) DialPeer(address string) {
	debug(true, "DialPeer message for %s", address)
	c.commandChannel <- CommandDialPeer{Address: address}
}

func (c *Controller) AddPeer(connection *Connection) {
	debug(true, "CommandAddPeer for %+v", connection)
	c.commandChannel <- CommandAddPeer{Peer: *connection}
}

func (c *Controller) NetworkStop() {
	debug(true, "NetworkStop ")
	c.commandChannel <- CommandShutdown{}
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
	note(true, "Controller.listen(%s) got address %s", c.listenPort, address)
	listener, err := net.Listen("tcp", address)
	if nil != err {
		logfatal(true, "Controller.listen() Error: %+v", err)
	}
	go c.acceptLoop(listener)
}

// Since this runs in its own goroutine we need to send a command when
// when we get a new connection.
func (c *Controller) acceptLoop(listener net.Listener) {
	note(true, "Controller.acceptLoop() starting up")
	for {
		conn, err := listener.Accept()
		if nil != err {
			logerror(true, "Controller.acceptLoop() Error: %+v", err)
		} else {
			peer := new(Connection).Init() //.(*Connection)
			peer.Configure(conn)
			c.AddPeer(peer) // Sends command to add the peer to the peers list
			note(true, "Controller.acceptLoop() new peer: %+v", peer.ConnectionID)
		}
	}
}

//////////////////////////////////////////////////////////////////////
// Operations
//////////////////////////////////////////////////////////////////////

// runloop is a goroutine that does all the heavy lifting
func (c *Controller) runloop() {
	note(true, "Controller.runloop() starting up")

	for c.keepRunning { // Run until we get the exit command
		verbose(true, "Controller.runloop() Calling router")
		c.router() // Route messages
		// Process commands...
		verbose(true, "Controller.runloop() About to process commands. Commands in channel: %d", len(c.commandChannel))
		for command := range c.commandChannel {
			verbose(true, "Controller.runloop() Processing command: %+v", command)
			c.handleCommand(command)
		}
		// route messages to and from application
	}
	note(true, "Controller.runloop() has exited. Shutdown command recieved?")
}

// Route pulls all of the messages from the application and sends them to the appropriate
// peer. Broadcast messages go to everyone, directed messages go to the named peer.
// route also passes incomming messages on to the application.
func (c *Controller) router() {
	note(true, "Controller.route() called. Number peers: %d", len(c.peers))

	// Recieve messages from the peers & forward to application.
	for id, peer := range c.peers {
		// Empty the recieve channel, stuff the application channel.
		note(true, "Controller.route() size of recieve channel: %d", len(peer.ReceiveChannel))

		for parcel := range peer.ReceiveChannel {
			debug(true, "Controller.route() got parcel from NETWORK %+v", parcel)
			parcel.Header.ConnectionID = id // Set the connection ID so the application knows which peer the message is from.
			c.FromNetwork <- parcel
		}
	}
	// For each message, see if it is directed, if so, send to the
	// specific peer, otherwise, broadcast.
	note(true, "Controller.route() size of ToNetwork channel: %d", len(c.ToNetwork))

	for parcel := range c.ToNetwork {
		debug(true, "Controller.route() got parcel from APPLICATION %+v", parcel)
		if 0 != parcel.Header.ConnectionID { // directed send
			debug(true, "Controller.route() Directed send to %+v", parcel.Header.ConnectionID)

			peer := c.peers[parcel.Header.ConnectionID]
			peer.SendChannel <- parcel
		} else { // broadcast
			debug(true, "Controller.route() Boadcast send to %d peers", len(c.peers))

			for _, peer := range c.peers {
				debug(true, "Controller.route() Send to peer %d ", peer.ConnectionID)

				peer.SendChannel <- parcel
			}
		}
	}
}

func (c *Controller) handleCommand(command interface{}) {
	switch commandType := command.(type) {
	case CommandDialPeer: // parameter is the peer address
		parameters := command.(CommandDialPeer)
		connection := new(Connection).Init() //.(*Connection)
		connection.dial(parameters.Address)
		c.peers[connection.ConnectionID] = *connection
		debug(true, "Controller.handleCommand(CommandDialPeer) got peer %s", parameters.Address)
	case CommandAddPeer: // parameter is a Connection
		parameters := command.(CommandAddPeer)
		connection := parameters.Peer
		c.peers[connection.ConnectionID] = connection
		debug(true, "Controller.handleCommand(CommandAddPeer) got peer %+v", parameters.Peer)
	case CommandShutdown:
		c.shutdown()
		debug(true, "Controller.handleCommand(CommandAddPeer) ")
	case CommandChangeLogging:
		parameters := command.(CommandChangeLogging)
		CurrentLoggingLevel = parameters.Level
		debug(true, "Controller.handleCommand(CommandChangeLogging) new logging level %s", LoggingLevels[parameters.Level])

	default:
		note(true, "Unkown p2p.Controller command recieved: %+v", commandType)
	}
}

func (c *Controller) shutdown() {
	debug(true, "Controller.shutdown() ")
	// Go thru peer list and shut down connections.
	// BUGBUG
	c.keepRunning = false
}
