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
	"strings"
	"time"
)

// Controller manages the peer to peer network.
type Controller struct {
	keepRunning bool // Indicates its time to shut down when false.

	// After launching the network, the management is done via these channels.
	commandChannel chan interface{} // Application use controller public API to send commands on this channel to controllers goroutines.

	ToNetwork   chan Parcel // Parcels from the application for us to route
	FromNetwork chan Parcel // Parcels from the network for the application

	listenPort           string                       // port we listen on for new connections
	connections          map[string]Connection        // map of the connections indexed by peer hash
	connectionsByAddress map[string]Connection        // map of the connections indexed by peer address
	connectionMetrics    map[string]ConnectionMetrics // map of the metrics indexed by peer hash

	discovery                  Discovery // Our discovery structure
	numberIncommingConnections int       // In PeerManagmeent we track this and refuse incomming connections when we have too many.
	lastPeerManagement         time.Time // Last time we ran peer management.
	NodeID                     uint64
	lastStatusReport           time.Time
	lastPeerRequest            time.Time // Last time we asked peers about the peers they know about.
}

type ControllerInit struct {
	Port      string    // Port to listen on
	PeersFile string    // Path to file to find / save peers
	Network   NetworkID // Network - eg MainNet, TestNet etc.
	Exclusive bool      // flag to indicate we should only connect to trusted peers
	SeedURL   string    // URL to a source of peer info
}

// CommandDialPeer is used to instruct the Controller to dial a peer address
type CommandDialPeer struct {
	persistent bool
	peer       Peer
}

// CommandAddPeer is used to instruct the Controller to add a connection
// This connection can come from acceptLoop or some other way.
type CommandAddPeer struct {
	conn net.Conn
}

// CommandShutdown is used to instruct the Controller to takve various actions.
type CommandShutdown struct {
	_ uint8
}

// CommandAdjustPeerQuality is used to instruct the Controller to reduce a connections quality score
type CommandAdjustPeerQuality struct {
	peerHash   string
	adjustment int32
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

func (c *Controller) Init(ci ControllerInit) *Controller {
	significant("ctrlr", "Controller.Init(%s) %#x", ci.Port, ci.Network)
	silence("#################", "META: Last touched: WEDNESDAY JULY 6 7:45PM")
	c.keepRunning = true
	c.commandChannel = make(chan interface{}, 1000) // Commands from App
	c.FromNetwork = make(chan Parcel, 10000)        // Channel to the app for network data
	c.ToNetwork = make(chan Parcel, 10000)          // Parcels from the app for the network
	c.listenPort = ci.Port
	NetworkListenPort = ci.Port
	c.connections = make(map[string]Connection)
	c.connectionMetrics = make(map[string]ConnectionMetrics)
	discovery := new(Discovery).Init(ci.PeersFile)
	c.discovery = *discovery
	c.discovery.seedURL = ci.SeedURL
	RandomGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))
	NodeID = uint64(RandomGenerator.Int63()) // This is a global used by all connections
	c.lastPeerManagement = time.Now()
	c.lastPeerRequest = time.Now()
	CurrentNetwork = ci.Network
	OnlySpecialPeers = ci.Exclusive
	return c
}

// StartNetwork configures the network, starts the runloop
func (c *Controller) StartNetwork() {
	verbose("ctrlr", "Controller.StartNetwork(%s)", " ")
	c.lastStatusReport = time.Now()
	// start listening on port given
	c.listen()
	// Dial out to peers
	c.fillOutgoingSlots()
	// Start the runloop
	go c.runloop()
}

func (c *Controller) StartLogging(level uint8) {
	note("ctrlr", "StartLogging() Changing log level to %s", LoggingLevels[level])
	c.commandChannel <- CommandChangeLogging{level: level}
}
func (c *Controller) StopLogging() {
	level := Silence
	note("ctrlr", "StopLogging() Changing log level to %s", LoggingLevels[level])
	c.commandChannel <- CommandChangeLogging{level: level}
}
func (c *Controller) ChangeLogLevel(level uint8) {
	note("ctrlr", "Changing log level to %s", LoggingLevels[level])
	c.commandChannel <- CommandChangeLogging{level: level}
}

func (c *Controller) DialPeer(peer Peer, persistent bool) {
	debug("ctrlr", "DialPeer message for %s", peer.Address)
	c.commandChannel <- CommandDialPeer{peer: peer, persistent: persistent}
}

func (c *Controller) AddPeer(conn net.Conn) {
	debug("ctrlr", "CommandAddPeer for %+v", conn)
	c.commandChannel <- CommandAddPeer{conn: conn}
}

func (c *Controller) NetworkStop() {
	debug("ctrlr", "NetworkStop ")
	c.commandChannel <- CommandShutdown{}
}

func (c *Controller) AdjustPeerQuality(peerHash string, adjustment int32) {
	debug("ctrlr", "AdjustPeerQuality ")
	c.commandChannel <- CommandAdjustPeerQuality{peerHash: peerHash, adjustment: adjustment}
}

func (c *Controller) Ban(peerHash string) {
	debug("ctrlr", "Ban %s ", peerHash)
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
	debug("ctrlr", "Controller.listen(%s) got address %s", c.listenPort, address)
	listener, err := net.Listen("tcp", address)
	if nil != err {
		logfatal("ctrlr", "Controller.listen() Error: %+v", err)
	} else {
		go c.acceptLoop(listener)
	}
}

// Since this runs in its own goroutine we need to send a command when
// when we get a new connection.
func (c *Controller) acceptLoop(listener net.Listener) {
	note("ctrlr", "Controller.acceptLoop() starting up")
	for {
		conn, err := listener.Accept()
		switch err {
		case nil:
			switch {
			case c.numberIncommingConnections < MaxNumberIncommingConnections:
				c.AddPeer(conn) // Sends command to add the peer to the peers list
				note("ctrlr", "Controller.acceptLoop() new peer: %+v", conn)
			default:
				note("ctrlr", "Controller.acceptLoop() new peer, but too many incomming connections. %d", c.numberIncommingConnections)
				conn.Close()
			}
		default:
			logerror("ctrlr", "Controller.acceptLoop() Error: %+v", err)
		}
	}
}

//////////////////////////////////////////////////////////////////////
// Operations
//////////////////////////////////////////////////////////////////////

// runloop is a goroutine that does all the heavy lifting
func (c *Controller) runloop() {
	// In long running processes it seems the runloop is exiting.
	reportExit := func() {
		significant("ctrlr", "@@@@@@@@@@ Controller.runloop() has exited! Here's its final state:")
		if 0 < CurrentLoggingLevel {
			significant("ctrlr", "%+v", c)
		}
		significant("ctrlr", "###################################")
		significant("ctrlr", " Network Controller Status Report:")
		significant("ctrlr", "===================================")
		significant("ctrlr", "     # Connections: %d", len(c.connections))
		significant("ctrlr", "Unique Connections: %d", len(c.connectionsByAddress))
		significant("ctrlr", "     Command Queue: %d", len(c.commandChannel))
		significant("ctrlr", "         ToNetwork: %d", len(c.ToNetwork))
		significant("ctrlr", "       FromNetwork: %d", len(c.FromNetwork))
		significant("ctrlr", "        Total RECV: %d", TotalMessagesRecieved)
		significant("ctrlr", "  Application RECV: %d", ApplicationMessagesRecieved)
		significant("ctrlr", "        Total XMIT: %d", TotalMessagesSent)
		significant("ctrlr", "###################################")
		significant("ctrlr", "@@@@@@@@@@ Controller.runloop() is terminated!")
	}
	defer reportExit()

	note("ctrlr", "Controller.runloop() starting up")
	// time.Sleep(time.Second * 5) // Wait a few seconds to let the system come up.

	for c.keepRunning { // Run until we get the exit command
		time.Sleep(time.Millisecond * 5) // This can be a tight loop, don't want to starve the application
		if CurrentLoggingLevel > 1 {
			fmt.Printf("@")
		}
		verbose("ctrlr", "@@@@@@@@@@ Controller.runloop() About to process commands. Commands in channel: %d", len(c.commandChannel))
		for 0 < len(c.commandChannel) {
			command := <-c.commandChannel
			verbose("ctrlr", "@@@@@@@@@@ Controller.runloop() handleCommand()")
			c.handleCommand(command)
		}
		// route messages to and from application
		verbose("ctrlr", "@@@@@@@@@@ Controller.runloop() Calling router")
		c.route() // Route messages
		// Manage peers
		verbose("ctrlr", "@@@@@@@@@@ Controller.runloop() Calling managePeers")
		c.managePeers()
		verbose("ctrlr", "@@@@@@@@@@ Controller.runloop() Checking Logging level")
		if CurrentLoggingLevel > 0 {
			verbose("ctrlr", "@@@@@@@@@@ Controller.runloop() networkStatusReport()")
			c.networkStatusReport()
		}
	}
	silence("ctrlr", "Controller.runloop() has exited. Shutdown command recieved?")
	significant("ctrlr", "runloop() - Final network statistics: TotalMessagesRecieved: %d TotalMessagesSent: %d", TotalMessagesRecieved, TotalMessagesSent)
}

// Route pulls all of the messages from the application and sends them to the appropriate
// peer. Broadcast messages go to everyone, directed messages go to the named peer.
// route also passes incomming messages on to the application.
func (c *Controller) route() {
	// verbose("ctrlr", "Controller.route() called. Number peers: %d", len(c.connections))
	// Recieve messages from the peers & forward to application.
	for peerHash, connection := range c.connections {
		// Empty the recieve channel, stuff the application channel.
		// verbose(peerHash, "Controller.route() size of recieve channel: %d", len(connection.ReceiveChannel))
		for 0 < len(connection.ReceiveChannel) { // effectively "While there are messages"
			message := <-connection.ReceiveChannel
			switch message.(type) {
			case ConnectionCommand:
				verbose(peerHash, "Controller.route() ConnectionCommand")
				c.handleConnectionCommand(message.(ConnectionCommand), connection)
			case ConnectionParcel:
				verbose(peerHash, "Controller.route() ConnectionParcel")
				c.handleParcelReceive(message, peerHash, connection)
			default:
				logfatal("ctrlr", "route() unknown message?: %+v ", message)
			}
		}
	}
	// For each message, see if it is directed, if so, send to the
	// specific peer, otherwise, broadcast.
	// significant("ctrlr", "Controller.route() size of ToNetwork channel: %d", len(c.ToNetwork))
	for 0 < len(c.ToNetwork) { // effectively "While there are messages"
		parcel := <-c.ToNetwork
		TotalMessagesSent++
		verbose("ctrlr", "Controller.route() got parcel from APPLICATION %+v", parcel.Header)
		if "" != parcel.Header.TargetPeer { // directed send
			debug("ctrlr", "Controller.route() Directed send to %+v", parcel.Header.TargetPeer)
			connection, present := c.connections[parcel.Header.TargetPeer]
			if present { // We're still connected to the target
				connection.SendChannel <- ConnectionParcel{parcel: parcel}
			}
		} else { // broadcast
			debug("ctrlr", "Controller.route() Broadcast send to %d peers", len(c.connections))
			for _, connection := range c.connections {
				verbose("ctrlr", "Controller.route() Send to peer %s ", connection.peer.Hash)
				connection.SendChannel <- ConnectionParcel{parcel: parcel}
			}
		}
	}
}

// handleParcelReceive takes a parcel from the network and annotates it for the application then routes it.
func (c *Controller) handleParcelReceive(message interface{}, peerHash string, connection Connection) {
	TotalMessagesRecieved++
	parameters := message.(ConnectionParcel)
	parcel := parameters.parcel
	verbose("ctrlr", "Controller.route() got parcel from NETWORK %+v", parcel.MessageType())
	parcel.Header.TargetPeer = peerHash // Set the connection ID so the application knows which peer the message is from.
	switch parcel.Header.Type {
	case TypeMessage: // Application message, send it on.
		ApplicationMessagesRecieved++
		c.FromNetwork <- parcel
	case TypePeerRequest: // send a response to the connection over its connection.SendChannel
		// Get selection of peers from discovery
		response := NewParcel(CurrentNetwork, c.discovery.SharePeers())
		response.Header.Type = TypePeerResponse
		// Send them out to the network - on the connection that requested it!
		connection.SendChannel <- ConnectionParcel{parcel: *response}
		verbose("ctrlr", "Controller.route() sent the SharePeers response: %+v", response.MessageType())
	case TypePeerResponse:
		// Add these peers to our known peers
		c.discovery.LearnPeers(parcel)
	default:
		logfatal("ctrlr", "handleParcelReceive() unknown parcel.Header.Type?: %+v ", parcel)
	}

}

func (c *Controller) handleConnectionCommand(command ConnectionCommand, connection Connection) {
	switch command.command {
	case ConnectionUpdateMetrics:
		c.connectionMetrics[connection.peer.Hash] = command.metrics
		note("ctrlr", "handleConnectionCommand() Got ConnectionUpdateMetrics command, all metrics are: %+v", c.connectionMetrics)
	case ConnectionIsClosed:
		debug("ctrlr", "handleConnectionCommand() Got ConnectionIsShutdown from  %s", connection.peer.Hash)
		delete(c.connectionsByAddress, connection.peer.Address)
		delete(c.connections, connection.peer.Hash)
	case ConnectionUpdatingPeer:
		debug("ctrlr", "handleConnectionCommand() Got ConnectionUpdatingPeer from  %s", connection.peer.Hash)
		c.discovery.updatePeer(command.peer)
	default:
		logfatal("ctrlr", "handleParcelReceive() unknown command.command?: %+v ", command.command)
	}
}

func (c *Controller) handleCommand(command interface{}) {
	switch commandType := command.(type) {
	case CommandDialPeer: // parameter is the peer address
		parameters := command.(CommandDialPeer)
		conn := new(Connection).Init(parameters.peer, parameters.persistent)
		connection := *conn
		connection.Start()
		c.connections[connection.peer.Hash] = connection
		c.connectionsByAddress[connection.peer.Address] = connection
		debug("ctrlr", "Controller.handleCommand(CommandDialPeer) got peer %s", parameters.peer.Address)
	case CommandAddPeer: // parameter is a Connection. This message is sent by the accept loop which is in a different goroutine
		parameters := command.(CommandAddPeer)
		conn := parameters.conn // net.Conn
		addPort := strings.Split(conn.RemoteAddr().String(), ":")
		debug("ctrlr", "Controller.handleCommand(CommandAddPeer) got rconn.RemoteAddr().String() %s and parsed IP: %s and Port: %s",
			conn.RemoteAddr().String(), addPort[0], addPort[1])
		// Port initially stored will be the connection port (not the listen port), but peer will update it on first message.
		peer := new(Peer).Init(addPort[0], addPort[1], 0, RegularPeer, 0)
		peer.Source["Accept()"] = time.Now()
		connection := new(Connection).InitWithConn(conn, *peer)
		connection.Start()
		c.connections[connection.peer.Hash] = *connection
		c.connectionsByAddress[connection.peer.Address] = *connection
		debug("ctrlr", "Controller.handleCommand(CommandAddPeer) got peer %+v", *peer)
	case CommandShutdown:
		silence("ctrlr", "handleCommand() Processing command: CommandShutdown")
		c.shutdown()
	case CommandChangeLogging:
		parameters := command.(CommandChangeLogging)
		CurrentLoggingLevel = parameters.level
		significant("ctrlr", "Controller.handleCommand(CommandChangeLogging) new logging level %s", LoggingLevels[parameters.level])
	case CommandAdjustPeerQuality:
		verbose("ctrlr", "handleCommand() Processing command: CommandDemerit")
		parameters := command.(CommandAdjustPeerQuality)
		peerHash := parameters.peerHash
		c.applicationPeerUpdate(parameters.adjustment, peerHash)
	case CommandBan:
		verbose("ctrlr", "handleCommand() Processing command: CommandBan")
		parameters := command.(CommandBan)
		peerHash := parameters.peerHash
		c.applicationPeerUpdate(BannedQualityScore, peerHash)
	default:
		logfatal("ctrlr", "Unkown p2p.Controller command recieved: %+v", commandType)
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
		debug("ctrlr", "managePeers() time since last peer management: %s", managementDuration.String())
		// If we are low on outgoing onnections, attempt to connect to some more.
		// If the connection is not online, we don't count it as connected.
		outgoing := 0
		c.numberIncommingConnections = 0
		for _, connection := range c.connections {
			if connection.IsOutGoing() && connection.IsOnline() {
				outgoing++
			} else {
				c.numberIncommingConnections++
			}
		}
		debug("ctrlr", "managePeers() NumberPeersToConnect: %d outgoing: %d", NumberPeersToConnect, outgoing)

		if NumberPeersToConnect > outgoing {
			// Get list of peers ordered by quality from discovery
			c.fillOutgoingSlots()
		}
		duration := time.Since(c.discovery.lastPeerSave)
		// Every so often, tell the discovery service to save peers.
		if PeerSaveInterval < duration {
			significant("controller", "Saving peers")
			c.discovery.SavePeers()
			c.discovery.PrintPeers() // No-op if debugging off.
		}
		duration = time.Since(c.lastPeerRequest)
		if PeerRequestInterval < duration {
			c.lastPeerRequest = time.Now()
			parcelp := NewParcel(CurrentNetwork, []byte("Peer Request"))
			parcel := *parcelp
			parcel.Header.Type = TypePeerRequest
			for _, connection := range c.connections {
				connection.SendChannel <- ConnectionParcel{parcel: parcel}
			}
		}
	}
}

// updateConnectionAddressMap() updates the address index map to reflect all current connections
func (c *Controller) updateConnectionAddressMap() {
	c.connectionsByAddress = map[string]Connection{}
	for _, value := range c.connections {
		c.connectionsByAddress[value.peer.Address] = value
	}
}

func (c *Controller) weAreNotAlreadyConnectedTo(peer Peer) bool {
	_, present := c.connectionsByAddress[peer.Address]
	return !present
}

func (c *Controller) fillOutgoingSlots() {
	c.updateConnectionAddressMap()
	significant("controller", "\n##############\n##############\n##############\n##############\n##############\n")
	significant("controller", "Connected peers:")
	for _, v := range c.connectionsByAddress {
		significant("controller", "%s : %s", v.peer.Address, v.peer.Port)
	}
	peers := c.discovery.GetOutgoingPeers()
	if len(peers) < NumberPeersToConnect*2 {
		c.discovery.GetOutgoingPeers()
		peers = c.discovery.GetOutgoingPeers()
	}
	// dial into the peers
	for _, peer := range peers {
		if c.weAreNotAlreadyConnectedTo(peer) {
			significant("controller", "We think we are not already connected to: %s so dialing.", peer.AddressPort())
			c.DialPeer(peer, false)
		}
	}
	c.discovery.PrintPeers()
	significant("controller", "\n##############\n##############\n##############\n##############\n##############\n")
}

func (c *Controller) shutdown() {
	debug("ctrlr", "Controller.shutdown() ")
	// Go thru peer list and shut down connections.
	for _, connection := range c.connections {
		connection.SendChannel <- ConnectionCommand{command: ConnectionShutdownNow}
	}
	//BUGBUG Make sure connetions are actually shut down.
	c.keepRunning = false
}

func (c *Controller) networkStatusReport() {
	durationSinceLastReport := time.Since(c.lastStatusReport)
	note("ctrlr", "networkStatusReport() NetworkStatusInterval: %s durationSinceLastReport: %s c.lastStatusReport: %s", NetworkStatusInterval.String(), durationSinceLastReport.String(), c.lastStatusReport.String())
	if durationSinceLastReport > NetworkStatusInterval {
		c.lastStatusReport = time.Now()
		silence("ctrlr", "###################################")
		silence("ctrlr", " Network Controller Status Report:")
		silence("ctrlr", "===================================")
		c.updateConnectionAddressMap()
		silence("ctrlr", "     # Connections: %d", len(c.connections))
		silence("ctrlr", "Unique Connections: %d", len(c.connectionsByAddress))
		silence("ctrlr", "     Command Queue: %d", len(c.commandChannel))
		silence("ctrlr", "         ToNetwork: %d", len(c.ToNetwork))
		silence("ctrlr", "       FromNetwork: %d", len(c.FromNetwork))
		silence("ctrlr", "        Total RECV: %d", TotalMessagesRecieved)
		silence("ctrlr", "  Application RECV: %d", ApplicationMessagesRecieved)
		silence("ctrlr", "        Total XMIT: %d", TotalMessagesSent)
		silence("ctrlr", "###################################")
	}
}
