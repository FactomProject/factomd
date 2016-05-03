// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

// P2PController manages the P2PNetwork.
// It maintains the list of peers, and has a master run-loop that
// processes ingoing and outgoing messages.
// It is controlled via a command channel.
// Other than Init and NetworkStart, all administration is done via the channel.

import (
	"fmt"
	"net"
)

// P2PController manages the peer to peer network.
type P2PController struct {
	keepRunning bool // Indicates its time to shut down when false.

	// After launching the network, the management is done via these channels.
	commandChannel chan  // Commands from the Application
	// responseChannel chan P2PResponseMessage // Command responses to the Application

	ToNetwork   chan Parcel // Parcels from the application for us to route
	FromNetwork chan Parcel // Parcels from the network for the application

	listenPort string                   // port we listen on for new connections
	peers      map[uint64]P2PConnection // map of the peers indexed by peer id
}

// JAYJAY - can't think of a situation where responses are needed.
// type P2PResponseMessage struct {
//     Error   error   // Error == nil on success
// }
type P2PCommand uint16

const (
	DialPeer P2PCommand = iota // parameter is an address of the peer
	Shutdown
	StartLogging // parameter is the log level
	StopLogging
	ChangeLogLevel
	AddPeer // Parameter is a p2pConnection
)

// P2PCommands are used to instruct the P2PController to takve various actions.
type P2PCommandDialPeer struct {
	Address    string
}

//////////////////////////////////////////////////////////////////////
// Public (exported) methods.
//
// The surface for interfacting with this is very minimal to avoid deadlocks
// and allow maximum concurrency.
// Other than setup, these API communicate with the controller via the
// command channel.
//////////////////////////////////////////////////////////////////////

func (p *P2PController) Init(port string) *P2PController {
	p.keepRunning = true
	p.commandChannel = make(chan, 1000) // Commands from App
	// p.responseChannel = make(<-chan P2PResponseMessage, 1000)
	p.FromNetwork = make(chan Parcel, 1000) // Channel to the app for network data
	p.ToNetwork = make(chan Parcel, 1000)   // Parcels from the app for the network
	p.listenPort = port
	p.peers = make(map[uint64]P2PConnection)
	return p
}

// NetworkStart configures the network, starts the runloop
func (p *P2PController) StartNetwork() {
	// start listening on port given
	p.listen()
	// dial into the peers
	/// start heartbeat process
	// Start the runloop
	go p.runloop()
}

func (p *P2PController) StartLogging(level uint8) {
	p.commandChannel <- P2PCommandMessage{Command: StartLogging, Parameters: [1]uint8{level}}
}
func (p *P2PController) StopLogging() {
	p.commandChannel <- P2PCommandMessage{Command: StopLogging, Parameters: nil}
}
func (p *P2PController) ChangeLogLevel(level uint8) {
	p.commandChannel <- P2PCommandMessage{Command: ChangeLogLevel, Parameters: [1]uint8{level}}
}

func (p *P2PController) DialPeer(address string) {
	p.commandChannel <- P2PCommandMessage{Command: AddPeer, Parameters: [1]string{address}}
}

func (p *P2PController) AddPeer(connection *P2PConnection) {
	p.commandChannel <- P2PCommandMessage{Command: AddPeer, Parameters: [1]P2PConnection{*connection}}
}

func (p *P2PController) NetworkStop() {
	p.commandChannel <- P2PCommandMessage{Command: Shutdown, Parameters: nil}
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

func (p *P2PController) listen() {
	address := fmt.Sprintf(":%s", p.listenPort)
	// debug(true, "P2PController.listen(%s) got address %s", p.listenPort, address)
	listener, err := net.Listen("tcp", address)
	if nil != err {
		log(Fatal, true, "P2PController.listen() Error: %+v", err)
	}
	go p.acceptLoop(listener)
}

// Since this runs in its own goroutine we need to send a command when
// when we get a new connection.
func (p *P2PController) acceptLoop(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if nil != err {
			log(Notes, true, "P2PController.acceptLoop() Error: %+v", err)
		} else {
			peer := new(P2PConnection).Init() //.(*P2PConnection)
			peer.Configure(conn)
			p.AddPeer(peer) // Sends command to add the peer to the peers list
			log(Notes, true, "P2PController.acceptLoop() new peer: %+v", peer.ConnectionID)
		}
	}
}

//////////////////////////////////////////////////////////////////////
// Operations
//////////////////////////////////////////////////////////////////////

// runloop is a goroutine that does all the heavy lifting
func (p *P2PController) runloop() {
	for p.keepRunning { // Run until we get the exit command
		// Process commands...
		for command := range p.commandChannel {
			p.handleCommand(command)
		}
		// route messages to and from application
		p.route() // Route messages
	}
	note(true, "P2PController.runloop() has exited. Shutdown command recieved?")
}

// Route pulls all of the messages from the application and sends them to the appropriate
// peer. Broadcast messages go to everyone, directed messages go to the named peer.
// route also passes incomming messages on to the application.
func (p *P2PController) route() {
	// Recieve messages from the peers & forward to application.
	for id, peer := range p.peers {
		// Empty the recieve channel, stuff the application channel.
		for parcel := range peer.ReceiveChannel {
			parcel.Header.ConnectionID = id // Set the connection ID so the application knows which peer the message is from.
			p.FromNetwork <- parcel
		}
	}
	// For each message, see if it is directed, if so, send to the
	// specific peer, otherwise, broadcast.
	for parcel := range p.ToNetwork {
		if 0 != parcel.Header.ConnectionID { // directed send
			peer := p.peers[parcel.Header.ConnectionID]
			peer.SendChannel <- parcel
		} else { // broadcast
			for _, peer := range p.peers {
				peer.SendChannel <- parcel
			}
		}
	}
}

func (p *P2PController) handleCommand(message P2PCommandMessage) {
	switch message.Command {
	case DialPeer: // parameter is the peer address
		connection := new(P2PConnection).Init() //.(*P2PConnection)
		parameterList := message.Parameters.([]string)
		connection.dial(parameterList[0])
		p.peers[connection.ConnectionID] = *connection
		// add connection to peers list
	case AddPeer: // parameter is a P2PConnection
		connection := parameterList[0].(P2PConnection)
		p.peers[connection.ConnectionID] = connection
	case Shutdown:
		p.shutdown()
	case StartLogging:
		P2PCurrentLoggingLevel = parameterList[0].(uint8)
	case StopLogging:
		P2PCurrentLoggingLevel = Silence
	case ChangeLogLevel:
		P2PCurrentLoggingLevel = parameterList[0].(uint8)
	}
}

func (p *P2PController) shutdown() {
	// Go thru peer list and shut down connections.
	// BUGBUG
	p.keepRunning = false
}
