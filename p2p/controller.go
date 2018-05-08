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

	"github.com/FactomProject/factomd/common/primitives"

	log "github.com/sirupsen/logrus"
)

// packageLogger is the general logger for all p2p related logs. You can add additional fields,
// or create more context loggers off of this
var packageLogger = log.WithFields(log.Fields{
	"package":   "p2p",
	"component": "networking"})

var controllerLogger = packageLogger.WithField("subpack", "controller")

// Controller manages the peer to peer network.
type Controller struct {
	keepRunning bool // Indicates its time to shut down when false.

	listenPort  string             // port we listen on for new connections
	connections *ConnectionManager // current connections

	// After launching the network, the management is done via these channels.
	commandChannel chan interface{} // Application use controller public API to send commands on this channel to controllers goroutines.

	ToNetwork   chan interface{} // Parcels from the application for us to route
	FromNetwork chan interface{} // Parcels from the network for the application

	connectionMetricsChannel chan interface{} // Channel on which we put the connection metrics map, periodically.

	connectionMetrics           map[string]ConnectionMetrics // map of the metrics indexed by peer hash
	lastConnectionMetricsUpdate time.Time                    // update once a second.

	discovery Discovery // Our discovery structure

	lastPeerManagement   time.Time // Last time we ran peer management.
	lastDiscoveryRequest time.Time
	NodeID               uint64
	lastStatusReport     time.Time
	lastPeerRequest      time.Time        // Last time we asked peers about the peers they know about.
	specialPeers         map[string]*Peer // special peers (from config file and from the command line params) by peer address
	partsAssembler       *PartsAssembler  // a data structure that assembles full messages from received message parts

	// logging
	logger *log.Entry
}

type ControllerInit struct {
	NodeName                 string           // Name of the current node
	Port                     string           // Port to listen on
	PeersFile                string           // Path to file to find / save peers
	Network                  NetworkID        // Network - eg MainNet, TestNet etc.
	Exclusive                bool             // flag to indicate we should only connect to trusted peers
	ExclusiveIn              bool             // flag to indicate we should only connect to trusted peers and disallow incoming connections
	SeedURL                  string           // URL to a source of peer info
	ConfigPeers              string           // Peers to always connect to at startup, and stay persistent, passed from the config file
	CmdLinePeers             string           // Additional special peers passed from the command line
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

//////////////////////////////////////////////////////////////////////
// Public (exported) methods.
//
// The surface for interfacting with this is very minimal to avoid deadlocks
// and allow maximum concurrency.
// Other than setup, these API communicate with the controller via the
// command channel.
//////////////////////////////////////////////////////////////////////

func (c *Controller) Init(ci ControllerInit) *Controller {
	c.logger = controllerLogger.WithFields(log.Fields{
		"node":    ci.NodeName,
		"port":    ci.Port,
		"network": fmt.Sprintf("%#x", ci.Network)})
	c.logger.WithField("controller_init", ci).Debugf("Initializing network controller")
	RandomGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))
	NodeID = uint64(RandomGenerator.Int63()) // This is a global used by all connections
	c.keepRunning = true
	c.commandChannel = make(chan interface{}, StandardChannelSize) // Commands from App
	c.FromNetwork = make(chan interface{}, StandardChannelSize)    // Channel to the app for network data
	c.ToNetwork = make(chan interface{}, StandardChannelSize)      // Parcels from the app for the network
	c.connections = new(ConnectionManager).Init()
	c.connectionMetrics = make(map[string]ConnectionMetrics)
	c.connectionMetricsChannel = ci.ConnectionMetricsChannel
	c.listenPort = ci.Port
	NetworkListenPort = ci.Port
	// Set this to the past so we will do peer management almost right away after starting up.
	c.lastPeerManagement = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	c.lastPeerRequest = time.Now()
	CurrentNetwork = ci.Network
	OnlySpecialPeers = ci.Exclusive || ci.ExclusiveIn
	AllowUnknownIncomingPeers = !ci.ExclusiveIn
	c.initSpecialPeers(ci)
	c.lastDiscoveryRequest = time.Now() // Discovery does its own on startup.
	c.lastConnectionMetricsUpdate = time.Now()
	c.partsAssembler = new(PartsAssembler).Init()
	discovery := new(Discovery).Init(ci.PeersFile, ci.SeedURL)
	c.discovery = *discovery
	return c
}

// StartNetwork configures the network, starts the runloop
func (c *Controller) StartNetwork() {
	c.logger.Info("Starting network")
	c.lastStatusReport = time.Now()
	// start listening on port given
	c.listen()
	// Dial all the gathered special peers
	c.dialSpecialPeers()
	// Start the runloop
	go c.runloop()
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

func (c *Controller) GetNumberOfConnections() int {
	return c.connections.Count()
}

func (c *Controller) ReloadSpecialPeers(newPeersConfig string) {
	c.logger.Info("Reloading special peers after config file change")
	newPeers := make(map[string]*Peer)
	for _, newPeer := range c.parseSpecialPeers(newPeersConfig, SpecialPeerConfig) {
		newPeers[newPeer.Address] = newPeer
	}

	toBeAdded := make([]*Peer, 0, len(newPeers))
	toBeRemoved := make([]*Peer, 0, len(c.specialPeers))

	for address, newPeer := range newPeers {
		_, exists := c.specialPeers[address]
		if !exists {
			c.logger.Infof("Detected a new peer in the config file: %s", address)
			toBeAdded = append(toBeAdded, newPeer)
		}
	}

	for address, oldPeer := range c.specialPeers {
		_, exists := newPeers[address]
		if exists {
			if oldPeer.Type == SpecialPeerCmdLine {
				c.logger.Warnf(
					"Detected a peer removed from the config file,"+
						" but it was earlier defined in the command line, ignoring: %s",
					address,
				)
				continue
			}
			c.logger.Infof("Detected a peer removed from the config file: %s")
			toBeRemoved = append(toBeRemoved, oldPeer)
		}
	}

	for _, peer := range toBeRemoved {
		delete(c.specialPeers, peer.Address)
		c.Disconnect(peer.Hash)
	}

	for _, peer := range toBeAdded {
		c.specialPeers[peer.Address] = peer
		c.DialPeer(*peer, true)
	}
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

func (c *Controller) dialSpecialPeers() {
	for _, peer := range c.specialPeers {
		c.DialPeer(*peer, true) // these are persistent connections
	}
}

func (c *Controller) listen() {
	address := fmt.Sprintf(":%s", c.listenPort)
	c.logger.WithFields(log.Fields{"address": address, "port": c.listenPort}).Infof("Listening for new connections")
	listener, err := net.Listen("tcp", address)
	if nil != err {
		c.logger.Errorf("Controller.listen() Error: %+v", err)
	} else {
		go c.acceptLoop(listener)
	}
}

// Since this runs in its own goroutine we need to send a command when
// when we get a new connection.
func (c *Controller) acceptLoop(listener net.Listener) {
	c.logger.Debug("Controller.acceptLoop() starting up")
	for {
		conn, err := listener.Accept()
		if err != nil {
			c.logger.Warnf("Controller.acceptLoop() Error: %+v", err)
			continue
		}

		connLogger := c.logger.WithField("remote_address", conn.RemoteAddr())

		if ok, reason := c.canConnectTo(conn); !ok {
			connLogger.Infof("Rejecting new connection request: %s", reason)
			_ = conn.Close()
			continue
		}

		c.AddPeer(conn) // Sends command to add the peer to the peers list
		connLogger.Infof("Accepting new incoming connection")
	}
}

func (c *Controller) canConnectTo(conn net.Conn) (bool, string) {
	incoming := c.connections.CountIf(func(c *Connection) bool {
		return !c.IsOutGoing() && c.IsOnline()
	})
	if incoming >= MaxNumberIncomingConnections {
		return false, "too many incoming connections"
	}

	if !AllowUnknownIncomingPeers && !c.isSpecialPeer(conn) {
		return false, "not a special peer and unknown incoming connections are not allowed"
	}

	return true, ""
}

func (c *Controller) isSpecialPeer(conn net.Conn) bool {
	for _, peer := range c.specialPeers {
		if peer.IsSamePeerAs(conn.RemoteAddr()) {
			return true
		}
	}
	return false
}

func (c *Controller) initSpecialPeers(ci ControllerInit) {
	c.specialPeers = make(map[string]*Peer)
	configPeers := c.parseSpecialPeers(ci.ConfigPeers, SpecialPeerConfig)
	cmdLinePeers := c.parseSpecialPeers(ci.CmdLinePeers, SpecialPeerCmdLine)

	// command line peers overwrite config peers
	for _, peer := range configPeers {
		c.specialPeers[peer.Address] = peer
	}
	for _, peer := range cmdLinePeers {
		c.specialPeers[peer.Address] = peer
	}
}

func (c *Controller) parseSpecialPeers(peersString string, peerType uint8) []*Peer {
	parseFunc := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsPunct(c)
	}
	peerAddresses := strings.FieldsFunc(peersString, parseFunc)
	peers := make([]*Peer, 0, len(peerAddresses))
	for _, peerAddress := range peerAddresses {
		address, port, err := net.SplitHostPort(peerAddress)
		if err != nil {
			c.logger.Errorf("%s is not a valid peer (%v), use format: 127.0.0.1:8999", peersString, err)
		} else {
			peer := new(Peer).Init(address, port, 0, peerType, 0)
			peer.Source["Local-Configuration"] = time.Now()
			peers = append(peers, peer)
		}
	}

	return peers
}

//////////////////////////////////////////////////////////////////////
// Operations
//////////////////////////////////////////////////////////////////////

// runloop is a goroutine that does all the heavy lifting
func (c *Controller) runloop() {
	// In long running processes it seems the runloop is exiting.
	c.logger.Debugf("Controller.runloop() @@@@@@@@@@ starting up in %d seconds", 2)
	time.Sleep(time.Second * time.Duration(2)) // Wait a few seconds to let the system come up.

	for c.keepRunning { // Run until we get the exit command
		c.connections.UpdatePrometheusMetrics()
		p2pControllerNumMetrics.Set(float64(len(c.connectionMetrics)))

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
		// route messages to and from application
		c.route() // Route messages
		// Manage peers
		c.managePeers()
		c.updateMetrics()
	}
}

// Route pulls all of the messages from the application and sends them to the appropriate
// peer. Broadcast messages go to everyone, directed messages go to the named peer.
// route also passes incoming messages on to the application.
func (c *Controller) route() {
	// Receive messages from the peers & forward to application.
	for peerHash, connection := range c.connections.All() {
		// Empty the receive channel, stuff the application channel.
		for 0 < len(connection.ReceiveChannel) { // effectively "While there are messages"
			message := <-connection.ReceiveChannel
			switch message.(type) {
			case ConnectionCommand:
				c.handleConnectionCommand(message.(ConnectionCommand), connection)
			case ConnectionParcel:
				c.handleParcelReceive(message, peerHash, connection)
			default:
				c.logger.Warnf("route() unknown message?: %+v ", message)
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
		case FullBroadcastFlag: // Send to all peers
			c.broadcast(parcel, true)
		case BroadcastFlag: // Send to many peers
			c.broadcast(parcel, false)
		case RandomPeerFlag: // Find a random peer, send to that peer.
			c.sendToRandomPeer(parcel)
		default: // Check if we're connected to the peer, if not drop message.
			c.logger.Debugf("Controller.route() Directed Neither Random nor Broadcast: %s Type: %s ", parcel.Header.TargetPeer, parcel.Header.AppType)
			c.doDirectedSend(parcel)
		}
	}
}

func (c *Controller) doDirectedSend(parcel Parcel) {
	connection, present := c.connections.GetByHash(parcel.Header.TargetPeer)
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
		c.logger.Warnf("handleParcelReceive() unknown parcel.Header.Type?: %+v ", parcel)
	}

}

func (c *Controller) handleConnectionCommand(command ConnectionCommand, connection *Connection) {
	switch command.Command {
	case ConnectionUpdateMetrics:
		c.connectionMetrics[connection.peer.Hash] = command.Metrics
	case ConnectionIsClosed:
		c.connections.Remove(connection)
		delete(c.connectionMetrics, connection.peer.Hash)
		go connection.goShutdown()
	case ConnectionUpdatingPeer:
		c.discovery.updatePeer(command.Peer)
	default:
		c.logger.Errorf("handleParcelReceive() unknown command.command?: %+v ", command.Command)
	}
}

func (c *Controller) handleCommand(command interface{}) {
	switch commandType := command.(type) {
	case CommandDialPeer: // parameter is the peer address
		parameters := command.(CommandDialPeer)
		conn := new(Connection).Init(parameters.peer, parameters.persistent)
		conn.Start()
		c.connections.Add(conn)
	case CommandAddPeer: // parameter is a Connection. This message is sent by the accept loop which is in a different goroutine

		parameters := command.(CommandAddPeer)
		conn := parameters.conn // net.Conn
		addPort := strings.Split(conn.RemoteAddr().String(), ":")
		// Port initially stored will be the connection port (not the listen port), but peer will update it on first message.
		peer := new(Peer).Init(addPort[0], addPort[1], 0, RegularPeer, 0)
		peer.Source["Accept()"] = time.Now()
		connection := new(Connection).InitWithConn(conn, *peer)
		connection.Start()
		c.connections.Add(connection)
	case CommandShutdown:
		c.shutdown()
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
		connection, present := c.connections.GetByHash(parameters.PeerHash)
		if present {
			BlockFreeChannelSend(connection.SendChannel, ConnectionCommand{Command: ConnectionShutdownNow})
		}
	default:
		c.logger.Errorf("Unknown p2p.Controller command received: %+v", commandType)
	}
}

func (c *Controller) applicationPeerUpdate(qualityDelta int32, peerHash string) {
	connection, present := c.connections.GetByHash(peerHash)
	if present {
		BlockFreeChannelSend(connection.SendChannel, ConnectionCommand{Command: ConnectionAdjustPeerQuality, Delta: qualityDelta})
	}
}

func (c *Controller) managePeers() {
	managementDuration := time.Since(c.lastPeerManagement)
	if PeerSaveInterval < managementDuration {
		c.lastPeerManagement = time.Now()
		c.logger.Debugf("managePeers() time since last peer management: %s", managementDuration.String())
		// If it's been awhile, update peers from the DNS seed.
		discoveryDuration := time.Since(c.lastDiscoveryRequest)
		if PeerDiscoveryInterval < discoveryDuration {
			c.logger.Debug("calling c.discovery.DiscoverPeersFromSeed()")
			c.discovery.DiscoverPeersFromSeed()
			c.logger.Debug("back from c.discovery.DiscoverPeersFromSeed()")
		}
		outgoingCount := c.connections.CountIf(func(c *Connection) bool {
			return c.IsOutGoing() && c.IsOnline()
		})
		c.logger.Debugf("managePeers() NumberPeersToConnect: %d outgoing: %d", NumberPeersToConnect, outgoingCount)
		if NumberPeersToConnect > outgoingCount {
			// Get list of peers ordered by quality from discovery
			c.fillOutgoingSlots(NumberPeersToConnect - outgoingCount)
		}
		duration := time.Since(c.discovery.lastPeerSave)
		// Every so often, tell the discovery service to save peers.
		if PeerSaveInterval < duration {
			c.logger.Debug("Saving peers")
			c.discovery.SavePeers()
		}
		duration = time.Since(c.lastPeerRequest)
		if PeerRequestInterval < duration {
			c.lastPeerRequest = time.Now()
			parcelp := NewParcel(CurrentNetwork, []byte("Peer Request"))
			parcel := *parcelp
			parcel.Header.Type = TypePeerRequest
			c.connections.SendToAll(ConnectionParcel{Parcel: parcel})
		}
	}
}

func (c *Controller) fillOutgoingSlots(openSlots int) {
	peers := c.discovery.GetOutgoingPeers()

	// To avoid dialing "too many" peers, we are keeping a count and only dialing the number of peers we need to add.
	newPeers := 0
	for _, peer := range peers {
		if c.connections.ConnectedTo(peer.Address) && newPeers < openSlots {
			c.logger.Debugf("newPeers: %d < openSlots: %d We think we are not already connected to: %s so dialing.", newPeers, openSlots, peer.AddressPort())
			newPeers = newPeers + 1
			c.DialPeer(peer, false)
		}
	}
}

func (c *Controller) updateMetrics() {
	if time.Second < time.Since(c.lastConnectionMetricsUpdate) {
		c.lastConnectionMetricsUpdate = time.Now()
		// Apparently golang doesn't make a deep copy when sending structs over channels. Bad golang.
		newMetrics := make(map[string]ConnectionMetrics)
		for key, value := range c.connections.All() {
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
					PeerType:         metrics.PeerType,
					ConnectionState:  metrics.ConnectionState,
					ConnectionNotes:  metrics.ConnectionNotes,
				}
			}
		}
		BlockFreeChannelSend(c.connectionMetricsChannel, newMetrics)
	}
}

func (c *Controller) shutdown() {
	c.logger.Debug("Controller.shutdown()")
	c.connections.SendToAll(ConnectionCommand{Command: ConnectionShutdownNow})
	c.keepRunning = false
}

// Broadcasts the parcel to a number of peers: all special peers and a random selection
// of regular peers (total max NumberPeersToBroadcast).
func (c *Controller) broadcast(parcel Parcel, full bool) {
	numSent := 0

	// always broadcast to special peers
	for _, peer := range c.specialPeers {
		connection, connected := c.connections.GetByHash(peer.Hash)
		if !connected {
			continue
		}
		numSent++
		BlockFreeChannelSend(connection.SendChannel, ConnectionParcel{Parcel: parcel})
	}

	// send also to a random selection of regular peers
	var randomSelection []*Connection
	if full {
		randomSelection = c.connections.GetAllRegular()
	} else {
		numToSendTo := NumberPeersToBroadcast - len(c.specialPeers)
		randomSelection = c.connections.GetRandomRegular(numToSendTo)
	}

	if len(randomSelection) == 0 {
		c.logger.Warn("Broadcast to random hosts failed: we don't have any peers to broadcast to")
		return
	}
	for _, connection := range randomSelection {
		BlockFreeChannelSend(connection.SendChannel, ConnectionParcel{Parcel: parcel})
	}

	SentToPeers.Set(float64(numSent))
}

func (c *Controller) sendToRandomPeer(parcel Parcel) {
	c.logger.Debugf("Controller.route() Directed FINDING RANDOM Target: %s Type: %s #Number Connections: %d", parcel.Header.TargetPeer, parcel.Header.AppType, c.connections.Count())
	randomConn := c.connections.GetRandom()

	if randomConn == nil {
		c.logger.Warn("Sending a parcel to a random peer failed: we don't have any peers to send to")
		return
	}

	parcel.Header.TargetPeer = randomConn.peer.Hash
	c.doDirectedSend(parcel)
}
