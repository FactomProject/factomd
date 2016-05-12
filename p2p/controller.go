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
	Peer Peer
}

// CommandAddPeer is used to instruct the Controller to add a connection 
// This connection can come from acceptLoop or some other way.
type CommandAddPeer struct {
	Peer Connection
}

// CommandShutdown is used to instruct the Controller to takve various actions.
type CommandShutdown struct {
	_ uint8
}
// CommandDemerit is used to instruct the Controller to reduce a connections quality score
type CommandDemerit struct {
	ConnectionID string
}

// CommandMerit is used to instruct the Controller to increase a connections quality score
type CommandMerit struct {
	ConnectionID string
}

// CommandBan is used to instruct the Controller to disconnect and ban a peer
type CommandBan struct {
	ConnectionID string
}

// CommandChangeLogging is used to instruct the Controller to takve various actions.
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

func (c *Controller) Init(port string, peersFile string) *Controller {
	verbose(true, "Controller.Init(%s)", port)
	c.keepRunning = true
	c.commandChannel = make(chan interface{}, 1000) // Commands from App
	c.FromNetwork = make(chan Parcel, 1000)         // Channel to the app for network data
	c.ToNetwork = make(chan Parcel, 1000)           // Parcels from the app for the network
	c.listenPort = port
	c.connections = make(map[string]Connection)
	c.discovery = new(Discovery).Init(peersFile)
	return c
}

// StartNetwork configures the network, starts the runloop
func (c *Controller) StartNetwork() {
	verbose(true, "Controller.StartNetwork(%s)", " ")
	// start listening on port given
	c.listen()
	// Start the discovery service?
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
	note("Changing log level to %s", LoggingLevels[level])
	c.commandChannel <- CommandChangeLogging{Level: level}
}
func (c *Controller) StopLogging() {
	note("Changing log level to %s", LoggingLevels[Silence])
	c.commandChannel <- CommandChangeLogging{Level: Silence}
}
func (c *Controller) ChangeLogLevel(level uint8) {
	note("Changing log level to %s", LoggingLevels[level])
	c.commandChannel <- CommandChangeLogging{Level: level}
}

func (c *Controller) DialPeer(address string) {
	debug("DialPeer message for %s", address)
	c.commandChannel <- CommandDialPeer{Address: address}
}

func (c *Controller) AddPeer(connection *Connection) {
	debug("CommandAddPeer for %+v", connection)
	c.commandChannel <- CommandAddPeer{Peer: *connection}
}

func (c *Controller) NetworkStop() {
	debug("NetworkStop ")
	c.commandChannel <- CommandShutdown{}
}

func (c *Controller) Demerit(connection uint64) {
	debug("NetworkStop ")
	c.commandChannel <- CommandDemerit{ConnectionID: connection}
}

func (c *Controller) Merit(connection uint64) {
	debug("NetworkStop ")
	c.commandChannel <- CommandMerit{ConnectionID: connection}
}

func (c *Controller) Ban(connection uint64) {
	debug("NetworkStop ")
	c.commandChannel <- CommandBan{ConnectionID: connection}
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
	note("Controller.listen(%s) got address %s", c.listenPort, address)
	listener, err := net.Listen("tcp", address)
	if nil != err {
		logfatal(true, "Controller.listen() Error: %+v", err)
	}
	go c.acceptLoop(listener)
}

// Since this runs in its own goroutine we need to send a command when
// when we get a new connection.
func (c *Controller) acceptLoop(listener net.Listener) {
	note("Controller.acceptLoop() starting up")
	for {
		conn, err := listener.Accept()
		if nil != err {
			logerror(true, "Controller.acceptLoop() Error: %+v", err)
		} else {
			address := conn.RemoteAddr().String()
			peer := c.discovery.GetPeerByHash(PeerHashFromAddress(address))
			connection := new(Connection).Init(*peer)
			connection.Configure(conn)
			c.AddPeer(connection) // Sends command to add the peer to the peers list
			note("Controller.acceptLoop() new peer: %+v", peer.address)
		}
	}
}

//////////////////////////////////////////////////////////////////////
// Operations
//////////////////////////////////////////////////////////////////////

// runloop is a goroutine that does all the heavy lifting
func (c *Controller) runloop() {
	note("Controller.runloop() starting up")

	for c.keepRunning { // Run until we get the exit command
		time.Sleep(time.Millisecond * 100)

		// Process commands...
		verbose(true, "Controller.runloop() About to process commands. Commands in channel: %d", len(c.commandChannel))
		for 0 < len(c.commandChannel) {
			command := <-c.commandChannel
			verbose(true, "Controller.runloop() Processing command: %+v", command)
			c.handleCommand(command)
		}
		// route messages to and from application
		verbose(true, "Controller.runloop() Calling router")
		c.router() // Route messages
		// Manage peers (reconnect, etc.)
		c.managePeers()
	}
	note("Controller.runloop() has exited. Shutdown command recieved?")
}

// Route pulls all of the messages from the application and sends them to the appropriate
// peer. Broadcast messages go to everyone, directed messages go to the named peer.
// route also passes incomming messages on to the application.
func (c *Controller) router() {
	verbose(true, "Controller.route() called. Number peers: %d", len(c.connections))

	// Recieve messages from the peers & forward to application.
	for id, peer := range c.connections {
		// Empty the recieve channel, stuff the application channel.
		verbose(true, "Controller.route() size of recieve channel: %d", len(peer.ReceiveChannel))
		for 0 < len(peer.ReceiveChannel) { // effectively "While there are messages"
			parcel := <-peer.ReceiveChannel
			verbose(true, "Controller.route() got parcel from NETWORK %+v", parcel)
			parcel.Header.ConnectionID = id // Set the connection ID so the application knows which peer the message is from.
			switch parcel.Header.Type {
			case TypeMessage: // Application message, send it on.
			c.FromNetwork <- parcel
				case TypePeerRequest:
				// Get selection of peers from discovery
				response := NewParcel(CurrentNetwork, c.discovery.SharePeers())
				response.Header.Type = TypePeerResponse
				// Send them out to the network - on the connection that requested it!
				peer.SendChannel <- response
					case TypePeerResponse:
					// Add these peers to our known peers
					c.discovery.LearnPeers(parcel.Payload)
			}
		}
	}
	// For each message, see if it is directed, if so, send to the
	// specific peer, otherwise, broadcast.
	verbose(true, "Controller.route() size of ToNetwork channel: %d", len(c.ToNetwork))
	for 0 < len(c.ToNetwork) { // effectively "While there are messages"
		parcel := <-c.ToNetwork
		verbose(true, "Controller.route() got parcel from APPLICATION %+v", parcel)
		if 0 != parcel.Header.ConnectionID { // directed send
			verbose(true, "Controller.route() Directed send to %+v", parcel.Header.ConnectionID)
			// BUGBUG Shouldn't this be the hash not the connectionID?
			peer := c.connections[parcel.Header.ConnectionID]
			peer.SendChannel <- parcel
		} else { // broadcast
			verbose(true, "Controller.route() Boadcast send to %d peers", len(c.connections))
			for _, peer := range c.connections {
				verbose(true, "Controller.route() Send to peer %d ", peer.ConnectionID)
				peer.SendChannel <- parcel
			}
		}
	}
	verbose(true, "Controller.route() Leaving Router Number peers: %d", len(c.connections))

}

func (c *Controller) handleCommand(command interface{}) {
	switch commandType := command.(type) {
	case CommandDialPeer: // parameter is the peer address
		parameters := command.(CommandDialPeer)
		if nil == c.connections[parameters.Peer] {
		connection := new(Connection).Init(*peer)
		connection.dial()
		c.connections[connection.ConnectionID] = *connection
		debug("Controller.handleCommand(CommandDialPeer) got peer %s", parameters.Address)
			
		} else {
		debug("Controller.handleCommand(CommandDialPeer) ALREADY CONNECTED TO PEER %s", parameters.Address)
			
		}

	case CommandAddPeer: // parameter is a Connection
		parameters := command.(CommandAddPeer)
		connection := parameters.Peer
		c.connections[connection.ConnectionID] = connection
		debug("Controller.handleCommand(CommandAddPeer) got peer %+v", parameters.Peer)
	case CommandShutdown:
		c.shutdown()
		debug("Controller.handleCommand(CommandAddPeer) ")
	case CommandChangeLogging:
		parameters := command.(CommandChangeLogging)
		CurrentLoggingLevel = parameters.Level
		debug("Controller.handleCommand(CommandChangeLogging) new logging level %s", LoggingLevels[parameters.Level])
	case CommandDemerit:
		parameters := command.(CommandDemerit)
		connectionID := parameters.ConnectionID
		connection := c.connections[connectionID]
		connection.peer.demerit()
	case CommandMerit:
		parameters := command.(CommandDemerit)
		connectionID := parameters.ConnectionID
		connection := c.connections[connectionID]
		connection.peer.merit()
	case CommandBan:
		parameters := command.(CommandDemerit)
		connectionID := parameters.ConnectionID
		connection := c.connections[connectionID]
		connection.peer.QualityScore = BannedQualityScore
		connection.connectionDropped() // hang up on the peer
	default:
		note("Unkown p2p.Controller command recieved: %+v", commandType)
	}
}

func (c *Controller) managePeers() {
	// check for and remove disconnected peers or peers offline after awhile
	for key, connection := range c.connections {
		if false == connection.Online {
			duration := time.Now().Sub(connection.timeLastAttempt)
			if MaxNumberOfRedialAttempts > connection.attempts && TimeBetweenRedials < duration {
				connection.dial()
			}
			if MaxNumberOfRedialAttempts <= connection.attempts { // give up on the connection
				//TODO BUGBUG Update discovery about this peer (eg: score)
				connection.shutdown()
				delete(c.connections, key)
			}
		} 
		// If it's been more than PingInterval since we last heard from a connection, send them a ping
		duration := time.Now().Sub(connection.timeLastContact)
		if PingInterval < duration {
		ping := NewParcel(CurrentNetwork, []byte("Pong"))
		pong.Header.Type = TypePing
		connection.SendChannel <- parcel 
			
			
		}
	}
	// Go thru an update peers in discovery using discovery.UpdatePeer()
	// so the known peers list is kept relatively up to date with peer score.
	BUGBUG IMPLEMENT
}

func (c *Controller) shutdown() {
	debug("Controller.shutdown() ")
	// Go thru peer list and shut down connections.
	for key, connection := range c.connections {
		connection.shutdown()
		delete(c.connections, key)
	}
	c.keepRunning = false
}
