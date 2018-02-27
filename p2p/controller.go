// Copyright 2017 Factom Foundation
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
	log "github.com/sirupsen/logrus"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/util/atomic"
)

// packageLogger is the general logger for all p2p related logs. You can add additional fields,
// or create more context loggers off of this
var packageLogger = log.WithFields(log.Fields{"package": "p2p"})

// Controller manages the peer to peer network.
type Controller struct {
	keepRunning bool // Indicates its time to shut down when false.

	listenPort              string                 // port we listen on for new connections
	connections             map[string]*Connection // map of the connections indexed by peer hash
	numConnections          atomic.AtomicUint32                 // Number of Connections we are managing.
	connectionsByAddress    map[string]*Connection // map of the connections indexed by peer address
	numConnectionsByAddress atomic.AtomicUint32                 // Number of Connections in the by address table

	// After launching the network, the management is done via these channels.
	commandChannel chan interface{} // Application use controller public API to send commands on this channel to controllers goroutines.

	ToNetwork   chan interface{} // Parcels from the application for us to route
	FromNetwork chan interface{} // Parcels from the network for the application

	connectionMetricsChannel chan interface{} // Channel on which we put the connection metrics map, periodically.

	connectionMetrics           map[string]ConnectionMetrics // map of the metrics indexed by peer hash
	lastConnectionMetricsUpdate time.Time                    // update once a second.

	discovery Discovery // Our discovery structure

	numberOutgoingConnections int       // In PeerManagement we track this to know when to dial out.
	numberIncomingConnections int       // In PeerManagement we track this and refuse incoming connections when we have too many.
	lastPeerManagement        time.Time // Last time we ran peer management.
	lastDiscoveryRequest      time.Time
	NodeID                    uint64
	lastStatusReport          time.Time
	lastPeerRequest           time.Time       // Last time we asked peers about the peers they know about.
	specialPeersString        string          // configuration set special peers
	partsAssembler            *PartsAssembler // a data structure that assembles full messages from received message parts
}

type ControllerInit struct {
	Port                     string           // Port to listen on
	PeersFile                string           // Path to file to find / save peers
	Network                  NetworkID        // Network - eg MainNet, TestNet etc.
	Exclusive                bool             // flag to indicate we should only connect to trusted peers
	SeedURL                  string           // URL to a source of peer info
	SpecialPeers             string           // Peers to always connect to at startup, and stay persistent
	ConnectionMetricsChannel chan interface{} // Channel on which we put the connection metrics map, periodically.
	LogPath                  string           // Path for logs
	LogLevel                 string           // Logging level
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
	PeerHash   string
	Adjustment int32
}

func (e *CommandAdjustPeerQuality) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *CommandAdjustPeerQuality) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *CommandAdjustPeerQuality) String() string {
	str, _ := e.JSONString()
	return str
}

// CommandBan is used to instruct the Controller to disconnect and ban a peer
type CommandBan struct {
	PeerHash string
}

func (e *CommandBan) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *CommandBan) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *CommandBan) String() string {
	str, _ := e.JSONString()
	return str
}

// CommandDisconnect is used to instruct the Controller to disconnect from a peer
type CommandDisconnect struct {
	PeerHash string
}

func (e *CommandDisconnect) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *CommandDisconnect) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *CommandDisconnect) String() string {
	str, _ := e.JSONString()
	return str
}

// CommandChangeLogging is used to instruct the Controller to takve various actions.
type CommandChangeLogging struct {
	Level uint8
}

func (e *CommandChangeLogging) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *CommandChangeLogging) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *CommandChangeLogging) String() string {
	str, _ := e.JSONString()
	return str
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
	c.connections = make(map[string]*Connection)
	c.connectionsByAddress = make(map[string]*Connection)
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
	c.partsAssembler = new(PartsAssembler).Init()
	discovery := new(Discovery).Init(ci.PeersFile, ci.SeedURL)
	c.discovery = *discovery
	// Set this to the past so we will do peer management almost right away after starting up.
	note("ctrlr", "\n\n\n\n\nController.Init(%s) Controller is: %+v\n\n", ci.Port, c)
	return c
}

// StartNetwork configures the network, starts the runloop
func (c *Controller) StartNetwork() {
	significant("ctrlr", "Controller.StartNetwork(%s)", " ")
	c.lastStatusReport = time.Now()
	// start listening on port given
	c.listen()
	// Dial the peers in from configuration
	c.DialSpecialPeersString(c.specialPeersString)
	// Start the runloop
	go c.runloop()
}

// DialSpecialPeersString lets us pass in a string of special peers to dial
func (c *Controller) DialSpecialPeersString(peersString string) {
	parseFunc := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsPunct(c)
	}
	peerAddresses := strings.FieldsFunc(peersString, parseFunc)
	for _, peerAddress := range peerAddresses {
		address, port, err := net.SplitHostPort(peerAddress)
		if err != nil {
			logerror("Controller", "DialSpecialPeersString: %s is not a valid peer (%v), use format: 127.0.0.1:8999", peersString, err)
		} else {
			peer := new(Peer).Init(address, port, 0, SpecialPeer, 0)
			peer.Source["Local-Configuration"] = time.Now()
			c.DialPeer(*peer, true) // these are persistent connections
		}
	}
}

func (c *Controller) StartLogging(level uint8) {
	BlockFreeChannelSend(c.commandChannel, CommandChangeLogging{Level: level})
}
func (c *Controller) StopLogging() {
	level := Silence
	BlockFreeChannelSend(c.commandChannel, CommandChangeLogging{Level: level})
}
func (c *Controller) ChangeLogLevel(level uint8) {
	BlockFreeChannelSend(c.commandChannel, CommandChangeLogging{Level: level})
}

func (c *Controller) DialPeer(peer Peer, persistent bool) {
	BlockFreeChannelSend(c.commandChannel, CommandDialPeer{peer: peer, persistent: persistent})
}

func (c *Controller) AddPeer(conn net.Conn) {
	BlockFreeChannelSend(c.commandChannel, CommandAddPeer{conn: conn})
}

func (c *Controller) NetworkStop() {
	if c != nil && c.commandChannel != nil {
		BlockFreeChannelSend(c.commandChannel, CommandShutdown{})
	}
}

func (c *Controller) AdjustPeerQuality(peerHash string, adjustment int32) {
	BlockFreeChannelSend(c.commandChannel, CommandAdjustPeerQuality{PeerHash: peerHash, Adjustment: adjustment})
}

func (c *Controller) Ban(peerHash string) {
	BlockFreeChannelSend(c.commandChannel, CommandBan{PeerHash: peerHash})
}

func (c *Controller) Disconnect(peerHash string) {
	BlockFreeChannelSend(c.commandChannel, CommandDisconnect{PeerHash: peerHash})
}

func (c *Controller) GetNumberConnections() int {
	return int(c.numConnections.Load())
}
func (c *Controller) getNumberConnectionsByAddress() int {
	return int(c.numConnectionsByAddress.Load())
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
			case c.numberIncomingConnections < MaxNumberIncomingConnections:
				c.AddPeer(conn) // Sends command to add the peer to the peers list
				note("ctrlr", "Controller.acceptLoop() new peer: %+v", conn)
			default:
				note("ctrlr", "Controller.acceptLoop() new peer, but too many incoming connections. %d", c.numberIncomingConnections)
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
		significant("ctrlr", "     # Connections: %d", c.GetNumberConnections())
		significant("ctrlr", "Unique Connections: %d", c.getNumberConnectionsByAddress())
		significant("ctrlr", "     Command Queue: %d", len(c.commandChannel))
		significant("ctrlr", "         ToNetwork: %d", len(c.ToNetwork))
		significant("ctrlr", "       FromNetwork: %d", len(c.FromNetwork))
		significant("ctrlr", "        Total RECV: %d", TotalMessagesReceived)
		significant("ctrlr", "  Application RECV: %d", ApplicationMessagesReceived)
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

		c.numConnections.Store(uint32(len(c.connections)))
		c.numConnectionsByAddress.Store(uint32(len(c.connectionsByAddress)))

		p2pControllerNumConnections.Set(float64(c.GetNumberConnections()))
		p2pControllerNumMetrics.Set(float64(len(c.connectionMetrics)))
		p2pControllerNumConnectionsByAddress.Set(float64(c.getNumberConnectionsByAddress()))

		dot("@@1\n")
	commandloop:
		for {
			select {
			case command := <-c.commandChannel:
				c.handleCommand(command)
			default:
				time.Sleep(time.Millisecond * 20)
				break commandloop
			}
		}
		dot("@@3\n")
		// route messages to and from application
		c.route() // Route messages
		dot("@@4\n")
		// Manage peers
		c.managePeers()
		dot("@@5\n")
		if CurrentLoggingLevel() > 0 {
			dot("@@6\n")
			c.networkStatusReport()
		}
		dot("@@7\n")
		c.updateMetrics()
		dot("@@11\n")

	}
	significant("ctrlr", "runloop() - Final network statistics: TotalMessagesReceived: %d TotalMessagesSent: %d", TotalMessagesReceived, TotalMessagesSent)
}

// Route pulls all of the messages from the application and sends them to the appropriate
// peer. Broadcast messages go to everyone, directed messages go to the named peer.
// route also passes incoming messages on to the application.
func (c *Controller) route() {
	// Receive messages from the peers & forward to application.
	for peerHash, connection := range c.connections {
		// Empty the receive channel, stuff the application channel.
		for 0 < len(connection.ReceiveChannel) { // effectively "While there are messages"
			message := <-connection.ReceiveChannel
			switch message.(type) {
			case ConnectionCommand:
				c.handleConnectionCommand(message.(ConnectionCommand), connection) // Used to pass a copy of the connection
			case ConnectionParcel:
				c.handleParcelReceive(message, peerHash, connection) // Used to pass a copy of the connection
			default:
				logfatal("ctrlr", "route() unknown message?: %+v ", message)
			}
		}
	}
	// For each message, see if it is directed, if so, send to the
	// specific peer, otherwise, broadcast.
	// significant("ctrlr", "Controller.route() size of ToNetwork channel: %d", len(c.ToNetwork))
	for 0 < len(c.ToNetwork) { // effectively "While there are messages"
		message := <-c.ToNetwork
		parcel := message.(Parcel)
		TotalMessagesSent++
		switch parcel.Header.TargetPeer {
		case BroadcastFlag: // Send to all peers

			// First off, how many nodes are we broadcasting to?  At least 4, if possible.  But 1/4 of the
			// number of connections if that is more than 4.
			num := NumberPeersToBroadcast
			clen := c.GetNumberConnections()
			if clen == 0 {
				return
			} else if clen < num {
				num = clen
			}
			quarter := clen / 4
			if quarter > num {
				num = quarter
			}

			// So at this point num <= clen, and we are going to send num sequentinial connections our message.
			// Note that if we run over the end of the connections, we wrap back to the start.  We don't assume
			// an order of connections, but we do assume that if we range over a map twice, we get the keys in
			// the same order both times.  (We do not modify the map)
			cnt := 0
			start := rand.Int() % clen
			spot := start
		broadcast:
			for i := 0; i < 2; i++ {
				loopcnt := 0
				for _, connection := range c.connections {
					if loopcnt == spot {
						BlockFreeChannelSend(connection.SendChannel, ConnectionParcel{Parcel: parcel})
						spot++
						if spot >= clen {
							spot = 0
						}
						cnt++
					}
					if cnt >= num {
						break broadcast
					}
					loopcnt++
				}
			}
			SentToPeers.Set(float64(cnt))
			StartingPoint.Set(float64(start))

		case RandomPeerFlag: // Find a random peer, send to that peer.
			debug("ctrlr", "Controller.route() Directed FINDING RANDOM Target: %s Type: %s #Number Connections: %d", parcel.Header.TargetPeer, parcel.Header.AppType, c.GetNumberConnections())
			bestKey := ""
		search:
			for i := 0; i < c.GetNumberConnections()*3; i++ {
				guess := (rand.Int() % c.GetNumberConnections())
				i := 0
				for key := range c.connections {
					if i == guess {
						connection := c.connections[key]
						connection.metricsMutex.Lock()
						bytes := connection.metrics.BytesReceived
						connection.metricsMutex.Unlock()
						if bytes > 0 {
							bestKey = key
							break search
						}
					}
					i++
				}
			}
			parcel.Header.TargetPeer = bestKey
			c.doDirectedSend(parcel)
		default: // Check if we're connected to the peer, if not drop message.
			debug("ctrlr", "Controller.route() Directed Neither Random nor Broadcast: %s Type: %s ", parcel.Header.TargetPeer, parcel.Header.AppType)
			c.doDirectedSend(parcel)
		}
	}
}

func (c *Controller) doDirectedSend(parcel Parcel) {
	connection, present := c.connections[parcel.Header.TargetPeer]
	if present { // We're still connected to the target
		BlockFreeChannelSend(connection.SendChannel, ConnectionParcel{Parcel: parcel})
	}
}

// handleParcelReceive takes a parcel from the network and annotates it for the application then routes it.
func (c *Controller) handleParcelReceive(message interface{}, peerHash string, connection *Connection) {
	TotalMessagesReceived++
	parameters := message.(ConnectionParcel)
	parcel := parameters.Parcel
	parcel.Header.TargetPeer = peerHash // Set the connection ID so the application knows which peer the message is from.
	switch parcel.Header.Type {
	case TypeMessage: // Application message, send it on.
		ApplicationMessagesReceived++
		BlockFreeChannelSend(c.FromNetwork, parcel)
	case TypeMessagePart: // A part of the application message, handle by assembler and if we have the full message, send it on.
		assembled := c.partsAssembler.handlePart(parcel)
		if assembled != nil {
			ApplicationMessagesReceived++
			BlockFreeChannelSend(c.FromNetwork, *assembled)
		}
	case TypePeerRequest: // send a response to the connection over its connection.SendChannel
		// Get selection of peers from discovery
		response := NewParcel(CurrentNetwork, c.discovery.SharePeers())
		response.Header.Type = TypePeerResponse
		// Send them out to the network - on the connection that requested it!
		BlockFreeChannelSend(connection.SendChannel, ConnectionParcel{Parcel: *response})
	case TypePeerResponse:
		// Add these peers to our known peers
		c.discovery.LearnPeers(parcel)
	default:
		logfatal("ctrlr", "handleParcelReceive() unknown parcel.Header.Type?: %+v ", parcel)
	}

}

func (c *Controller) handleConnectionCommand(command ConnectionCommand, connection *Connection) {
	switch command.Command {
	case ConnectionUpdateMetrics:
		c.connectionMetrics[connection.peer.Hash] = command.Metrics
	case ConnectionIsClosed:
		delete(c.connectionsByAddress, connection.peer.Address)
		delete(c.connections, connection.peer.Hash)
		delete(c.connectionMetrics, connection.peer.Hash)
		go connection.goShutdown()
	case ConnectionUpdatingPeer:
		c.discovery.updatePeer(command.Peer)
	default:
		logfatal("ctrlr", "handleParcelReceive() unknown command.command?: %+v ", command.Command)
	}
}

func (c *Controller) handleCommand(command interface{}) {
	switch commandType := command.(type) {
	case CommandDialPeer: // parameter is the peer address
		parameters := command.(CommandDialPeer)
		conn := new(Connection).Init(parameters.peer, parameters.persistent)
		conn.Start()

		c.connections[conn.peer.Hash] = conn
		c.connectionsByAddress[conn.peer.Address] = conn
	case CommandAddPeer: // parameter is a Connection. This message is sent by the accept loop which is in a different goroutine

		parameters := command.(CommandAddPeer)
		conn := parameters.conn // net.Conn
		addPort := strings.Split(conn.RemoteAddr().String(), ":")
		// Port initially stored will be the connection port (not the listen port), but peer will update it on first message.
		peer := new(Peer).Init(addPort[0], addPort[1], 0, RegularPeer, 0)
		peer.Source["Accept()"] = time.Now()
		connection := new(Connection).InitWithConn(conn, *peer)
		connection.Start()

		c.connections[connection.peer.Hash] = connection
		c.connectionsByAddress[connection.peer.Address] = connection
	case CommandShutdown:
		c.shutdown()
	case CommandChangeLogging:
		parameters := command.(CommandChangeLogging)
		CurrentLoggingLevelVar.Store(parameters.Level) // really a uint8 but still got reported as a race...
	case CommandAdjustPeerQuality:
		parameters := command.(CommandAdjustPeerQuality)
		peerHash := parameters.PeerHash
		c.applicationPeerUpdate(parameters.Adjustment, peerHash)
	case CommandBan:
		parameters := command.(CommandBan)
		peerHash := parameters.PeerHash
		c.applicationPeerUpdate(BannedQualityScore, peerHash)
	case CommandDisconnect:
		parameters := command.(CommandDisconnect)
		peerHash := parameters.PeerHash
		connection, present := c.connections[peerHash]
		if present {
			BlockFreeChannelSend(connection.SendChannel, ConnectionCommand{Command: ConnectionShutdownNow})
		}
	default:
		logfatal("ctrlr", "Unknown p2p.Controller command received: %+v", commandType)
	}
}

func (c *Controller) applicationPeerUpdate(qualityDelta int32, peerHash string) {
	connection, present := c.connections[peerHash]
	if present {
		BlockFreeChannelSend(connection.SendChannel, ConnectionCommand{Command: ConnectionAdjustPeerQuality, Delta: qualityDelta})
	}
}

func (c *Controller) managePeers() {
	managementDuration := time.Since(c.lastPeerManagement)
	if PeerSaveInterval < managementDuration {
		dot("&&s\n")
		c.lastPeerManagement = time.Now()
		significant("ctrlr", "managePeers() time since last peer management: %s", managementDuration.String())
		// If it's been awhile, update peers from the DNS seed.
		discoveryDuration := time.Since(c.lastDiscoveryRequest)
		if PeerDiscoveryInterval < discoveryDuration {
			note("ctrlr", "calling c.discovery.DiscoverPeersFromSeed()")
			c.discovery.DiscoverPeersFromSeed()
			note("ctrlr", "back from c.discovery.DiscoverPeersFromSeed()")
		}
		c.updateConnectionCounts()
		significant("ctrlr", "managePeers() NumberPeersToConnect: %d outgoing: %d", NumberPeersToConnect, c.numberOutgoingConnections)
		dot("&&t\n")
		if NumberPeersToConnect > c.numberOutgoingConnections {
			// Get list of peers ordered by quality from discovery
			c.fillOutgoingSlots(NumberPeersToConnect - c.numberOutgoingConnections)
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
				BlockFreeChannelSend(connection.SendChannel, ConnectionParcel{Parcel: parcel})
			}
		}
	}
}

func (c *Controller) updateConnectionCounts() {
	// If we are low on outgoing connections, attempt to connect to some more.
	// If the connection is not online, we don't count it as connected.
	c.numberOutgoingConnections = 0
	c.numberIncomingConnections = 0
	for _, connection := range c.connections {
		switch {
		case connection.IsOutGoing() && connection.IsOnline():
			c.numberOutgoingConnections++
		case !connection.IsOutGoing() && connection.IsOnline():
			c.numberIncomingConnections++
		default: // we don't count offline connections for these purposes.
		}
	}
}

// updateConnectionAddressMap() updates the address index map to reflect all current connections
func (c *Controller) updateConnectionAddressMap() {
	c.connectionsByAddress = map[string]*Connection{}
	for _, value := range c.connections {
		c.connectionsByAddress[value.peer.Address] = value
	}
}

func (c *Controller) weAreNotAlreadyConnectedTo(peer Peer) bool {
	_, present := c.connectionsByAddress[peer.Address]
	return !present
}

func (c *Controller) fillOutgoingSlots(openSlots int) {
	c.updateConnectionAddressMap()
	significant("controller", "Connected peers:")
	for _, v := range c.connectionsByAddress {
		significant("controller", "%s : %s", v.peer.Address, v.peer.Port)
	}
	peers := c.discovery.GetOutgoingPeers()

	// To avoid dialing "too many" peers, we are keeping a count and only dialing the number of peers we need to add.
	newPeers := 0
	for _, peer := range peers {
		if c.weAreNotAlreadyConnectedTo(peer) && newPeers < openSlots {
			note("controller", "newPeers: %d < openSlots: %d We think we are not already connected to: %s so dialing.", newPeers, openSlots, peer.AddressPort())
			newPeers = newPeers + 1
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
		BlockFreeChannelSend(connection.SendChannel, ConnectionCommand{Command: ConnectionShutdownNow})
	}
	c.keepRunning = false
}

func (c *Controller) networkStatusReport() {
	durationSinceLastReport := time.Since(c.lastStatusReport)
	note("ctrlr", "networkStatusReport() NetworkStatusInterval: %s durationSinceLastReport: %s c.lastStatusReport: %s", NetworkStatusInterval.String(), durationSinceLastReport.String(), c.lastStatusReport.String())
	if durationSinceLastReport > NetworkStatusInterval {
		c.lastStatusReport = time.Now()
		c.updateConnectionCounts()
		silence("ctrlr", "\n\n\n\n")
		silence("ctrlr", "###################################")
		silence("ctrlr", " Network Controller Status Report:")
		silence("ctrlr", "===================================")
		c.updateConnectionAddressMap()
		silence("ctrlr", "     # Connections: %d", c.GetNumberConnections())
		silence("ctrlr", "Unique Connections: %d", c.getNumberConnectionsByAddress())
		silence("ctrlr", "    In Connections: %d", c.numberIncomingConnections)
		silence("ctrlr", "   Out Connections: %d (only online are counted)", c.numberOutgoingConnections)
		silence("ctrlr", "        Total RECV: %d", TotalMessagesReceived)
		silence("ctrlr", "  Application RECV: %d", ApplicationMessagesReceived)
		silence("ctrlr", "        Total XMIT: %d", TotalMessagesSent)
		silence("ctrlr", " ")
		silence("ctrlr", "\tPeer\t\t\t\tDuration\tStatus\t\tNotes")
		silence("ctrlr", "-------------------------------------------------------------------------------")
		for _, v := range c.connections {
			metrics, present := c.connectionMetrics[v.peer.Hash]
			if !present {
				metrics = ConnectionMetrics{MomentConnected: time.Now(), ConnectionState: "No Metrics", ConnectionNotes: "No Metrics"}
			}
			silence("ctrlr", "Location: %d", v.peer.Location)
			silence("ctrlr", "%s\t%s\t%s\t%s", v.peer.PeerFixedIdent(), time.Since(metrics.MomentConnected), metrics.ConnectionState, metrics.ConnectionNotes)
			silence("ctrlr", "IsOutgoing: %t\tIsOnline: %t\tStatus: %s Quality: %d", v.IsOutGoing(), v.IsOnline(), v.StatusString(), metrics.PeerQuality)
			silence("ctrlr", "Sent/Recv: %d / %d\t\t Chan Send/Recv: %d / %d", metrics.MessagesSent, metrics.MessagesReceived, len(v.SendChannel), len(v.ReceiveChannel))
			silence("ctrlr", ".")
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
