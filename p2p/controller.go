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
	"math/rand"
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

	listenPort         string                // port we listen on for new connections
	connections        map[string]Connection // map of the peers indexed by peer hash
	discovery          Discovery             // Our discovery structure
	lastPeerManagement time.Time             // Last time we ran peer management.
	NodeID             uint64
	lastStatusReport   time.Time
	lastPeerRequest    time.Time // Last time we asked peers about the peers they know about.
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
	c.FromNetwork = make(chan Parcel, 10000)        // Channel to the app for network data
	c.ToNetwork = make(chan Parcel, 10000)          // Parcels from the app for the network
	c.listenPort = port
	c.connections = make(map[string]Connection)
	discovery := new(Discovery).Init(peersFile)
	c.discovery = *discovery
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	NodeID = uint64(r.Int63()) // This is a global used by all connections
	c.lastPeerManagement = time.Now()
	c.lastPeerRequest = time.Now()
	return c
}

// StartNetwork configures the network, starts the runloop
func (c *Controller) StartNetwork(exclusive bool) {
	verbose("controller", "Controller.StartNetwork(%s)", " ")
	// exclusive means we only connect to special peers
	OnlySpecialPeers = exclusive
	// start listening on port given
	c.listen()
	// Get a list of peers from discovery
	peers := c.discovery.GetStartupPeers()
	// dial into the peers
	for _, peer := range peers {
		c.DialPeer(peer.Address)
	}
	/// start heartbeat process
	c.lastStatusReport = time.Now()
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
	address := fmt.Sprintf(":%s", c.listenPort)
	note("controller", "Controller.listen(%s) got address %s", c.listenPort, address)
	listener, err := net.Listen("tcp", address)
	if nil != err {
		logfatal("controller", "Controller.listen() Error: %+v", err)
	} else {
		go c.acceptLoop(listener)
	}
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
			connection := new(Connection).InitWithConn(conn, peer)
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
	// time.Sleep(time.Second * 5) // Wait a few seconds to let the system come up.

	for c.keepRunning { // Run until we get the exit command
		time.Sleep(time.Millisecond * 1) // This can be a tight loop, don't want to starve the application
		// time.Sleep(time.Second * 1) // This can be a tight loop, don't want to starve the application
		// Process commands...
		// verbose("controller", "Controller.runloop() About to process commands. Commands in channel: %d", len(c.commandChannel))
		for 0 < len(c.commandChannel) {
			command := <-c.commandChannel
			c.handleCommand(command)
		}
		// route messages to and from application
		// verbose("controller", "Controller.runloop() Calling router")
		c.route() // Route messages
		// Manage peers
		// verbose("controller", "Controller.runloop() Calling managePeers")
		c.managePeers()

		if CurrentLoggingLevel > 0 {
			c.networkStatusReport()
		}

	}
	note("controller", "Controller.runloop() has exited. Shutdown command recieved?")
}

// Route pulls all of the messages from the application and sends them to the appropriate
// peer. Broadcast messages go to everyone, directed messages go to the named peer.
// route also passes incomming messages on to the application.
func (c *Controller) route() {
	// verbose("controller", "Controller.route() called. Number peers: %d", len(c.connections))
	// Recieve messages from the peers & forward to application.
	for peerHash, connection := range c.connections {
		// Empty the recieve channel, stuff the application channel.
		verbose(peerHash, "Controller.route() size of recieve channel: %d", len(connection.ReceiveChannel))
		for 0 < len(connection.ReceiveChannel) { // effectively "While there are messages"
			message := <-connection.ReceiveChannel
			switch message.(type) {
			case ConnectionCommand:
				debug(peerHash, "Controller.route() ConnectionCommand")
				c.handleConnectionCommand(message.(ConnectionCommand), connection)
			case ConnectionParcel:
				debug(peerHash, "Controller.route() ConnectionParcel")
				c.handleParcelReceive(message, peerHash, connection)
			default:
				logfatal("controller", "route() unknown message?: %+v ", message)
			}
		}
	}
	// For each message, see if it is directed, if so, send to the
	// specific peer, otherwise, broadcast.
	verbose("controller", "Controller.route() size of ToNetwork channel: %d", len(c.ToNetwork))
	for 0 < len(c.ToNetwork) { // effectively "While there are messages"
		parcel := <-c.ToNetwork
		verbose("controller", "Controller.route() got parcel from APPLICATION %+v", parcel.Header)
		if "" != parcel.Header.TargetPeer { // directed send
			verbose("controller", "Controller.route() Directed send to %+v", parcel.Header.TargetPeer)
			connection, present := c.connections[parcel.Header.TargetPeer]
			if present { // We're still connected to the target
				connection.SendChannel <- ConnectionParcel{parcel: parcel}
			}
		} else { // broadcast
			verbose("controller", "Controller.route() Broadcast send to %d peers", len(c.connections))
			for _, connection := range c.connections {
				verbose("controller", "Controller.route() Send to peer %s ", connection.peer.Hash)
				connection.SendChannel <- ConnectionParcel{parcel: parcel}
			}
		}
	}

}

// handleParcelReceive takes a parcel from the network and annotates it for the application then routes it.
func (c *Controller) handleParcelReceive(message interface{}, peerHash string, connection Connection) {
	parameters := message.(ConnectionParcel)
	parcel := parameters.parcel
	verbose("controller", "Controller.route() got parcel from NETWORK %+v", parcel.MessageType())
	parcel.Header.TargetPeer = peerHash // Set the connection ID so the application knows which peer the message is from.
	switch parcel.Header.Type {
	case TypeMessage: // Application message, send it on.
		c.FromNetwork <- parcel
	case TypePeerRequest: // send a response to the connection over its connection.SendChannel
		// Get selection of peers from discovery
		response := NewParcel(CurrentNetwork, c.discovery.SharePeers())
		response.Header.Type = TypePeerResponse
		// Send them out to the network - on the connection that requested it!
		connection.SendChannel <- ConnectionParcel{parcel: *response}
		verbose("controller", "Controller.route() sent the SharePeers response: %+v", response.MessageType())
	case TypePeerResponse:
		// Add these peers to our known peers
		c.discovery.LearnPeers(parcel.Payload)
	default:
		logfatal("controller", "handleParcelReceive() unknown parcel.Header.Type?: %+v ", parcel)
	}

}

func (c *Controller) handleConnectionCommand(command ConnectionCommand, connection Connection) {
	switch command.command {
	case ConnectionIsShutdown:
		debug("controller", "handleConnectionCommand() Got ConnectionIsShutdown from  %s", connection.peer.Hash)
		delete(c.connections, connection.peer.Hash)
	case ConnectionUpdatingPeer:
		debug("controller", "handleConnectionCommand() Got ConnectionUpdatingPeer from  %s", connection.peer.Hash)
		c.discovery.UpdatePeer(command.peer)
	default:
		logfatal("controller", "handleParcelReceive() unknown command.command?: %+v ", command.command)
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
			c.connections[connection.peer.Hash] = connection
			debug("controller", "Controller.handleCommand(CommandDialPeer) got peer %s", peer.Address)
		} else {
			debug("controller", "Controller.handleCommand(CommandDialPeer) ALREADY CONNECTED TO PEER %s", peer.Address)
		}
	case CommandAddPeer: // parameter is a Connection. This message is sent by the accept loop which is in a different goroutine
		parameters := command.(CommandAddPeer)
		connection := parameters.connection
		_, present := c.connections[connection.peer.Hash] // check if we are already connected to the peer
		if !present {                                     // we are not connected to the peer
			c.connections[connection.peer.Hash] = connection
		}
		debug("controller", "Controller.handleCommand(CommandAddPeer) got peer %+v", parameters.connection)
	case CommandShutdown:
		verbose("controller", "handleCommand() Processing command: CommandShutdown")
		c.shutdown()
	case CommandChangeLogging:
		parameters := command.(CommandChangeLogging)
		CurrentLoggingLevel = parameters.level
		debug("controller", "Controller.handleCommand(CommandChangeLogging) new logging level %s", LoggingLevels[parameters.level])
	case CommandDemerit:
		verbose("controller", "handleCommand() Processing command: CommandDemerit")
		parameters := command.(CommandDemerit)
		peerHash := parameters.peerHash
		c.applicationPeerUpdate(-1, peerHash)
	case CommandMerit:
		verbose("controller", "handleCommand() Processing command: CommandMerit")
		parameters := command.(CommandMerit)
		peerHash := parameters.peerHash
		c.applicationPeerUpdate(1, peerHash)
	case CommandBan:
		verbose("controller", "handleCommand() Processing command: CommandBan")
		parameters := command.(CommandBan)
		peerHash := parameters.peerHash
		c.applicationPeerUpdate(BannedQualityScore, peerHash)
	default:
		logfatal("controller", "Unkown p2p.Controller command recieved: %+v", commandType)
	}
}
func (c *Controller) applicationPeerUpdate(qualityDelta int32, peerHash string) {
	connection, present := c.connections[peerHash]
	if present {
		connection.SendChannel <- ConnectionCommand{command: ConnectionAdjustPeerQuality, delta: qualityDelta}
	}
}

func (c *Controller) managePeers() {
	managementDuration := time.Since(c.lastPeerManagement)
	if PeerSaveInterval < managementDuration {
		c.lastPeerManagement = time.Now()
		debug("controller", "managePeers() time since last peer management: %s", managementDuration.String())
		// If we are low on peers, attempt to connect to some more.
		if NumberPeersToConnect > len(c.connections) {
			// Get list of peers ordered by quality from discovery
			peers := c.discovery.GetStartupPeers()
			// For each one, if we don't already have a connection, create command message.
			for _, peer := range peers {
				_, present := c.connections[peer.Hash]
				if !present {
					c.DialPeer(peer.Address)
				}
			}
		}
		duration := time.Since(c.discovery.lastPeerSave)
		// Every so often, tell the discovery service to save peers.
		if PeerSaveInterval < duration {
			c.discovery.SavePeers()
			c.discovery.PrintPeers() // No-op if debugging off.
		}
		duration = time.Since(c.lastPeerRequest)
		if PeerRequestInterval < duration {
			for _, connection := range c.connections {
				parcel := NewParcel(CurrentNetwork, []byte("Peer Request"))
				parcel.Header.Type = TypePeerRequest
				connection.SendChannel <- ConnectionParcel{parcel: *parcel}
			}
		}
	}
}
func (c *Controller) shutdown() {
	debug("controller", "Controller.shutdown() ")
	// Go thru peer list and shut down connections.
	for _, connection := range c.connections {
		connection.SendChannel <- ConnectionCommand{command: ConnectionShutdownNow}
	}
	c.keepRunning = false
}

func (c *Controller) networkStatusReport() {
	reportDuration := time.Since(c.lastStatusReport)
	if reportDuration > NetworkStatusInterval {
		silence("conroller", "networkStatusReport() NetworkStatusInterval: %s reportDuration: %s c.lastPeerManagement: %s", NetworkStatusInterval.String(), reportDuration.String(), c.lastPeerManagement.String())
		c.lastStatusReport = time.Now()
		silence("conroller", "###########################")
		silence("conroller", "Network Status Report:")
		silence("conroller", "===========================")
		for key, value := range c.connections {
			silence("conroller", "Connection Hash: %s", key)
			silence("conroller", "          State: %s", value.ConnectionState())
			silence("conroller", " ReceiveChannel: %d", len(value.ReceiveChannel))
			silence("conroller", "    SendChannel: %d", len(value.SendChannel))
			// silence("conroller", "     Connection: %+v", value)
			silence("conroller", "===========================")
		}
		silence("conroller", "   Command Queue: %d", len(c.commandChannel))
		silence("conroller", "       ToNetwork: %d", len(c.ToNetwork))
		silence("conroller", "     FromNetwork: %d", len(c.FromNetwork))
		silence("conroller", "###########################")
	}
}
