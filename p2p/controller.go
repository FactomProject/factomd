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
	"unicode"
)

// Controller manages the peer to peer network.
type Controller struct {
	keepRunning bool // Indicates its time to shut down when false.

	listenPort           string                // port we listen on for new connections
	connections          map[string]Connection // map of the connections indexed by peer hash
	connectionsByAddress map[string]Connection // map of the connections indexed by peer address

	// After launching the network, the management is done via these channels.
	commandChannel chan interface{} // Application use controller public API to send commands on this channel to controllers goroutines.

	ToNetwork   chan interface{} // Parcels from the application for us to route
	FromNetwork chan interface{} // Parcels from the network for the application

	connectionMetricsChannel chan interface{} // Channel on which we put the connection metrics map, periodically.

	connectionMetrics           map[string]ConnectionMetrics // map of the metrics indexed by peer hash
	lastConnectionMetricsUpdate time.Time                    // update once a second.

	discovery Discovery // Our discovery structure

	numberIncommingConnections int       // In PeerManagmeent we track this and refuse incomming connections when we have too many.
	lastPeerManagement         time.Time // Last time we ran peer management.
	lastDiscoveryRequest       time.Time
	NodeID                     uint64
	lastStatusReport           time.Time
	lastPeerRequest            time.Time // Last time we asked peers about the peers they know about.
	specialPeersString         string    // configuration set special peers
}

type ControllerInit struct {
	Port                     string           // Port to listen on
	PeersFile                string           // Path to file to find / save peers
	Network                  NetworkID        // Network - eg MainNet, TestNet etc.
	Exclusive                bool             // flag to indicate we should only connect to trusted peers
	SeedURL                  string           // URL to a source of peer info
	SpecialPeers             string           // Peers to always connect to at startup, and stay persistent
	ConnectionMetricsChannel chan interface{} // Channel on which we put the connection metrics map, periodically.
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

// CommandDisconnect is used to instruct the Controller to disconnect from a peer
type CommandDisconnect struct {
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
	note("ctrlr", "\n\n\n\n\nController.Init(%s) %#x", ci.Port, ci.Network)
	note("ctrlr", "\n\n\n\n\nController.Init(%s) ci: %+v\n\n", ci.Port, ci)
	RandomGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))
	NodeID = uint64(RandomGenerator.Int63()) // This is a global used by all connections
	c.keepRunning = true
	c.commandChannel = make(chan interface{}, StandardChannelSize) // Commands from App
	c.FromNetwork = make(chan interface{}, StandardChannelSize)    // Channel to the app for network data
	c.ToNetwork = make(chan interface{}, StandardChannelSize)      // Parcels from the app for the network
	c.connections = make(map[string]Connection)
	c.connectionMetrics = make(map[string]ConnectionMetrics)
	c.connectionMetricsChannel = ci.ConnectionMetricsChannel
	c.listenPort = ci.Port
	NetworkListenPort = ci.Port
	c.lastPeerManagement = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	c.lastPeerRequest = time.Now()
	CurrentNetwork = ci.Network
	OnlySpecialPeers = ci.Exclusive
	c.specialPeersString = ci.SpecialPeers
	c.lastDiscoveryRequest = time.Now() // Discovery does its own on startup.
	c.lastConnectionMetricsUpdate = time.Now()
	discovery := new(Discovery).Init(ci.PeersFile, ci.SeedURL)
	c.discovery = *discovery
	// Set this to the past so we will do peer management almost right away after starting up.
	note("ctrlr", "\n\n\n\n\nController.Init(%s) Controller is: %+v\n\n", ci.Port, c)
	return c
}

// StartNetwork configures the network, starts the runloop
func (c *Controller) StartNetwork() {
	verbose("ctrlr", "Controller.StartNetwork(%s)", " ")
	c.lastStatusReport = time.Now()
	// start listening on port given
	c.listen()
	// Dial the peers in from configuration
	c.DialSpecialPeersString(c.specialPeersString)
	// Dial out to peers
	c.fillOutgoingSlots()
	// Start the runloop
	go c.runloop()
}

// DialSpecialPeersString lets us pass in a string of special peers to dial
func (c *Controller) DialSpecialPeersString(peersString string) {
	note("ctrlr", "DialSpecialPeersString() Dialing Special Peers %s", peersString)
	parseFunc := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsPunct(c)
	}
	peerAddresses := strings.FieldsFunc(peersString, parseFunc)
	for _, peerAddress := range peerAddresses {
		fmt.Println("Dialing Peer: ", peerAddress)
		ipPort := strings.Split(peerAddress, ":")
		if len(ipPort) == 2 {
			peer := new(Peer).Init(ipPort[0], ipPort[1], 0, SpecialPeer, 0)
			peer.Source["Local-Configuration"] = time.Now()
			c.DialPeer(*peer, true) // these are persistent connections
		} else {
			logfatal("Controller", "Error: %s is not a valid peer, use format: 127.0.0.1:8999", peerAddress)
		}
	}
}

func (c *Controller) StartLogging(level uint8) {
	note("ctrlr", "StartLogging() Changing log level to %s", LoggingLevels[level])
	BlockFreeChannelSend(c.commandChannel, CommandChangeLogging{level: level})
}
func (c *Controller) StopLogging() {
	level := Silence
	note("ctrlr", "StopLogging() Changing log level to %s", LoggingLevels[level])
	BlockFreeChannelSend(c.commandChannel, CommandChangeLogging{level: level})
}
func (c *Controller) ChangeLogLevel(level uint8) {
	note("ctrlr", "Changing log level to %s", LoggingLevels[level])
	BlockFreeChannelSend(c.commandChannel, CommandChangeLogging{level: level})
}

func (c *Controller) DialPeer(peer Peer, persistent bool) {
	debug("ctrlr", "DialPeer message for %s", peer.PeerIdent())
	BlockFreeChannelSend(c.commandChannel, CommandDialPeer{peer: peer, persistent: persistent})
}

func (c *Controller) AddPeer(conn net.Conn) {
	debug("ctrlr", "CommandAddPeer for %+v", conn)
	BlockFreeChannelSend(c.commandChannel, CommandAddPeer{conn: conn})
}

func (c *Controller) NetworkStop() {
	debug("ctrlr", "NetworkStop %+v", c)
	if c != nil && c.commandChannel != nil {
		BlockFreeChannelSend(c.commandChannel, CommandShutdown{})
	}
}

func (c *Controller) AdjustPeerQuality(peerHash string, adjustment int32) {
	debug("ctrlr", "AdjustPeerQuality ")
	BlockFreeChannelSend(c.commandChannel, CommandAdjustPeerQuality{peerHash: peerHash, adjustment: adjustment})
}

func (c *Controller) Ban(peerHash string) {
	debug("ctrlr", "Ban %s ", peerHash)
	BlockFreeChannelSend(c.commandChannel, CommandBan{peerHash: peerHash})
}

func (c *Controller) Disconnect(peerHash string) {
	debug("ctrlr", "Ban %s ", peerHash)
	BlockFreeChannelSend(c.commandChannel, CommandDisconnect{peerHash: peerHash})
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
		if err := conn.(*net.TCPConn).SetNoDelay(true); err != nil {
			panic("error, nodelay didn't take")

		}
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
		// if 0 < CurrentLoggingLevel {
		// 	significant("ctrlr", "%+v", c)
		// }
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

	// startDelay := 24
	// i := 1
	// note("ctrlr", "Controller.runloop() @@@@@@@@@@ starting up in %d seconds", startDelay)
	// for i <= startDelay {
	// 	time.Sleep(time.Second * 1)
	// 	note("ctrlr", "Controller.runloop() @@@@@@@@@@ starting up in %d seconds", startDelay-i)
	// 	i = i + 1
	// }
	note("ctrlr", "Controller.runloop() @@@@@@@@@@ starting up in %d seconds", 2)
	time.Sleep(time.Second * time.Duration(2)) // Wait a few seconds to let the system come up.

	for c.keepRunning { // Run until we get the exit command
		dot("@@1\n")
		progress := false
		for 0 < len(c.commandChannel) {
			command := <-c.commandChannel
			c.handleCommand(command)
			progress = true
		}
		if !progress {
			time.Sleep(time.Millisecond * 10) // This can be a tight loop, don't want to starve the application
		}
		dot("@@3\n")
		// route messages to and from application
		c.route() // Route messages
		dot("@@4\n")
		// Manage peers
		c.managePeers()
		dot("@@5\n")
		if CurrentLoggingLevel > 0 {
			dot("@@6\n")
			c.networkStatusReport()
		}
		dot("@@7\n")
		c.updateMetrics()
		dot("@@11\n")
	}
	silence("ctrlr", "Controller.runloop() has exited. Shutdown command recieved?")
	significant("ctrlr", "runloop() - Final network statistics: TotalMessagesRecieved: %d TotalMessagesSent: %d", TotalMessagesRecieved, TotalMessagesSent)
}

// Route pulls all of the messages from the application and sends them to the appropriate
// peer. Broadcast messages go to everyone, directed messages go to the named peer.
// route also passes incomming messages on to the application.
func (c *Controller) route() {
	dot("&&a\n")
	note("ctrlr", "ctrlr.route() called. Number peers: %d", len(c.connections))
	// Recieve messages from the peers & forward to application.
	for peerHash, connection := range c.connections {
		// Empty the recieve channel, stuff the application channel.
		dot("&&b\n")
		note(peerHash, "ctrlr.route() size of recieve channel: %d", len(connection.ReceiveChannel))

	loop:
		for {
			select {
			case message := <-connection.ReceiveChannel:
				switch message.(type) {
				case ConnectionCommand:
					note(peerHash, "ctrlr.route() ConnectionCommand")
					c.handleConnectionCommand(message.(ConnectionCommand), connection)
				case ConnectionParcel:
					note(peerHash, "ctrlr.route() ConnectionParcel")
					c.handleParcelReceive(message, peerHash, connection)
				default:
					logfatal("ctrlr", "route() unknown message?: %+v ", message)
				}
			default:
				break loop
			}
		}
	}
	// For each message, see if it is directed, if so, send to the
	// specific peer, otherwise, broadcast.
	// significant("ctrlr", "Controller.route() size of ToNetwork channel: %d", len(c.ToNetwork))
	dot("&&e\n")
	for 0 < len(c.ToNetwork) { // effectively "While there are messages"
		dot("&&f\n")
		message := <-c.ToNetwork
		dot("&&g\n")
		parcel := message.(Parcel)
		TotalMessagesSent++
		note("ctrlr", "Controller.route() got parcel from APPLICATION %+v", parcel.Header)
		if "" != parcel.Header.TargetPeer { // directed send
			dot("&&h\n")
			connection, present := c.connections[parcel.Header.TargetPeer]
			if present { // We're still connected to the target
				significant("ctrlr", "Controller.route() SUCCESS Directed send to %+v", parcel.Header.TargetPeer)
				dot("&&i\n")
				BlockFreeChannelSend(connection.SendChannel, ConnectionParcel{parcel: parcel})
			} else {
				significant("ctrlr", "Controller.route() FAILED! Target not present in connections! Directed send to %+v", parcel.Header.TargetPeer)
				for key, _ := range c.connections {
					significant("ctrlr", "Controller.route() %+v", key)
				}
			}
		} else { // broadcast
			dot("&&j\n")
			note("ctrlr", "Controller.route() Broadcast send to %d peers", len(c.connections))
			for _, connection := range c.connections {
				dot("&&k\n")
				verbose("ctrlr", "Controller.route() Send to peer %s ", connection.peer.Hash)
				BlockFreeChannelSend(connection.SendChannel, ConnectionParcel{parcel: parcel})
			}
		}
	}
}

// handleParcelReceive takes a parcel from the network and annotates it for the application then routes it.
func (c *Controller) handleParcelReceive(message interface{}, peerHash string, connection Connection) {
	TotalMessagesRecieved++
	parameters := message.(ConnectionParcel)
	parcel := parameters.parcel
	note("ctrlr", "Controller.route() got parcel from NETWORK %+v", parcel.MessageType())
	dot("&&l\n")
	parcel.Header.TargetPeer = peerHash // Set the connection ID so the application knows which peer the message is from.
	switch parcel.Header.Type {
	case TypeMessage: // Application message, send it on.
		dot("&&m\n")
		ApplicationMessagesRecieved++
		BlockFreeChannelSend(c.FromNetwork, parcel)
	case TypePeerRequest: // send a response to the connection over its connection.SendChannel
		dot("&&n\n")
		// Get selection of peers from discovery
		response := NewParcel(CurrentNetwork, c.discovery.SharePeers())
		response.Header.Type = TypePeerResponse
		// Send them out to the network - on the connection that requested it!
		BlockFreeChannelSend(connection.SendChannel, ConnectionParcel{parcel: *response})
		note("ctrlr", "Controller.route() sent the SharePeers response: %+v", response.MessageType())
	case TypePeerResponse:
		dot("&&o\n")
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
		dot("&&p\n")
		note("ctrlr", "handleConnectionCommand() Got ConnectionUpdateMetrics")
	case ConnectionIsClosed:
		dot("&&q\n")
		note("ctrlr", "handleConnectionCommand() Got ConnectionIsShutdown from  %s", connection.peer.Hash)
		delete(c.connectionsByAddress, connection.peer.Address)
		delete(c.connections, connection.peer.Hash)
		delete(c.connectionMetrics, connection.peer.Hash)
	case ConnectionUpdatingPeer:
		dot("&&r\n")
		note("ctrlr", "handleConnectionCommand() Got ConnectionUpdatingPeer from  %s", connection.peer.Hash)
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
	case CommandDisconnect:
		verbose("ctrlr", "handleCommand() Processing command: CommandDisconnect")
		parameters := command.(CommandDisconnect)
		peerHash := parameters.peerHash
		connection, present := c.connections[peerHash]
		if present {
			BlockFreeChannelSend(connection.SendChannel, ConnectionCommand{command: ConnectionShutdownNow})
		}
	default:
		logfatal("ctrlr", "Unkown p2p.Controller command recieved: %+v", commandType)
	}
}
func (c *Controller) applicationPeerUpdate(qualityDelta int32, peerHash string) {
	connection, present := c.connections[peerHash]
	if present {
		BlockFreeChannelSend(connection.SendChannel, ConnectionCommand{command: ConnectionAdjustPeerQuality, delta: qualityDelta})
	}
}

func (c *Controller) managePeers() {
	managementDuration := time.Since(c.lastPeerManagement)
	if PeerSaveInterval < managementDuration {
		dot("&&s\n")
		c.lastPeerManagement = time.Now()
		debug("ctrlr", "managePeers() time since last peer management: %s", managementDuration.String())
		// If it's been awhile, update peers from the DNS seed.
		discoveryDuration := time.Since(c.lastDiscoveryRequest)
		if PeerDiscoveryInterval < discoveryDuration {
			note("ctrlr", "calling c.discovery.DiscoverPeersFromSeed()")
			c.discovery.DiscoverPeersFromSeed()
			note("ctrlr", "back from c.discovery.DiscoverPeersFromSeed()")
		}
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
		note("ctrlr", "managePeers() NumberPeersToConnect: %d outgoing: %d", NumberPeersToConnect, outgoing)
		dot("&&t\n")

		if NumberPeersToConnect > outgoing {
			// Get list of peers ordered by quality from discovery
			c.fillOutgoingSlots()
		}
		duration := time.Since(c.discovery.lastPeerSave)
		// Every so often, tell the discovery service to save peers.
		if PeerSaveInterval < duration {
			note("controller", "Saving peers")
			c.discovery.SavePeers()
			c.discovery.PrintPeers() // No-op if debugging off.
		}
		dot("&&u\n")
		duration = time.Since(c.lastPeerRequest)
		if PeerRequestInterval < duration {
			c.lastPeerRequest = time.Now()
			parcelp := NewParcel(CurrentNetwork, []byte("Peer Request"))
			parcel := *parcelp
			parcel.Header.Type = TypePeerRequest
			for _, connection := range c.connections {
				BlockFreeChannelSend(connection.SendChannel, ConnectionParcel{parcel: parcel})
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
	note("controller", "Connected peers:")
	for _, v := range c.connectionsByAddress {
		note("controller", "%s : %s", v.peer.Address, v.peer.Port)
	}
	peers := c.discovery.GetOutgoingPeers()

	for _, peer := range peers {
		if c.weAreNotAlreadyConnectedTo(peer) {
			note("controller", "We think we are not already connected to: %s so dialing.", peer.AddressPort())
			c.DialPeer(peer, false)
		}
	}
	c.discovery.PrintPeers()
}

func (c *Controller) updateMetrics() {
	if time.Second < time.Since(c.lastConnectionMetricsUpdate) {
		dot("@@8\n")
		c.lastConnectionMetricsUpdate = time.Now()
		// Apparently golang doesn't make a deep copy when sending structs over channels. Bad golang.
		newMetrics := make(map[string]ConnectionMetrics)
		for key, value := range c.connections {
			metrics, present := c.connectionMetrics[value.peer.Hash]
			if present {
				newMetrics[key] = ConnectionMetrics{
					MomentConnected:  metrics.MomentConnected,
					BytesSent:        metrics.BytesSent,
					BytesReceived:    metrics.BytesReceived,
					MessagesSent:     metrics.MessagesSent,
					MessagesReceived: metrics.MessagesReceived,
					PeerAddress:      metrics.PeerAddress,
					PeerQuality:      metrics.PeerQuality,
					ConnectionState:  metrics.ConnectionState,
					ConnectionNotes:  metrics.ConnectionNotes,
				}
			}
		}
		dot("@@9\n")
		BlockFreeChannelSend(c.connectionMetricsChannel, newMetrics)
		dot("@@10\n")
	}
}

func (c *Controller) shutdown() {
	debug("ctrlr", "Controller.shutdown() ")
	// Go thru peer list and shut down connections.
	for _, connection := range c.connections {
		BlockFreeChannelSend(connection.SendChannel, ConnectionCommand{command: ConnectionShutdownNow})
	}
	c.keepRunning = false
}

func (c *Controller) networkStatusReport() {
	durationSinceLastReport := time.Since(c.lastStatusReport)
	note("ctrlr", "networkStatusReport() NetworkStatusInterval: %s durationSinceLastReport: %s c.lastStatusReport: %s", NetworkStatusInterval.String(), durationSinceLastReport.String(), c.lastStatusReport.String())
	if durationSinceLastReport > NetworkStatusInterval {
		c.lastStatusReport = time.Now()
		silence("ctrlr", "\n\n\n\n")
		silence("ctrlr", "###################################")
		silence("ctrlr", " Network Controller Status Report:")
		silence("ctrlr", "===================================")
		c.updateConnectionAddressMap()
		silence("ctrlr", "     # Connections: %d", len(c.connections))
		silence("ctrlr", "Unique Connections: %d", len(c.connectionsByAddress))
		silence("ctrlr", "        Total RECV: %d", TotalMessagesRecieved)
		silence("ctrlr", "  Application RECV: %d", ApplicationMessagesRecieved)
		silence("ctrlr", "        Total XMIT: %d", TotalMessagesSent)
		silence("ctrlr", " ")
		silence("ctrlr", "\tPeer\t\t\t\tDuration\tStatus\t\tNotes")
		silence("ctrlr", "-------------------------------------------------------------------------------")
		for _, v := range c.connections {
			metrics, present := c.connectionMetrics[v.peer.Hash]
			if !present {
				metrics = ConnectionMetrics{MomentConnected: time.Now(), ConnectionState: "No Metrics", ConnectionNotes: "No Metrics"}
			}
			silence("ctrlr", "%s\t%s\t%s\t%s", v.peer.PeerFixedIdent(), time.Since(metrics.MomentConnected), metrics.ConnectionState, metrics.ConnectionNotes)
			silence("ctrlr", "\t\tSent/Recv: %d / %d\t\t Chan Send/Recv: %d / %d", metrics.MessagesSent, metrics.MessagesReceived, len(v.SendChannel), len(v.ReceiveChannel))
		}
		silence("ctrlr", "\tChannels:")
		silence("ctrlr", "          commandChannel: %d", len(c.commandChannel))
		silence("ctrlr", "               ToNetwork: %d", len(c.ToNetwork))
		silence("ctrlr", "             FromNetwork: %d", len(c.FromNetwork))
		silence("ctrlr", "connectionMetricsChannel: %d", len(c.connectionMetricsChannel))
		silence("ctrlr", "===================================")
		silence("ctrlr", "###################################\n\n\n")
	}
}
