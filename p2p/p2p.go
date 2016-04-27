package p2p


// P2PController manages the P2PNetwork.
// It maintains the list of peers, and has a master run-loop that
// processes ingoing and outgoing messages.
// It is controlled via a command channel.
// Other than Init and NetworkStart, all administration is done via the channel.

import (
   
    
)

// P2PController manages the peer to peer network.
type P2PController struct {
    keepRunning bool    // Indicates its time to shut down when false.
    // After launching the network, the management is done via these channels.
    commandChannel  chan P2PCommandMessage // Commands from the Application
    // responseChannel chan P2PResponseMessage // Command responses to the Application
    peers map[uint64]P2PConnection // map of the peers indexed by peer id
}

// P2PCommandMessage are used to instruct the P2PController to takve various actions.
type P2PCommandMessage struct {
    Command uint16
    Parameters []interface{}
}

// JAYJAY - can't think of a situation where responses are needed.

// P2PCommandMessage are used to instruct the P2PController to takve various actions.
// type P2PResponseMessage struct {
//     Error   error   // Error == nil on success
// }

const uint16 (
    AddPeer = iota
    Shutdown
    StartLogging
    StopLogging
    ChangeLogLevel
)

//////////////////////////////////////////////////////////////////////
// Public (exported) methods. 
//
// The surface for interfacting with this is very minimal to avoid deadlocks
// and allow maximum concurrency.
// Other than setup, these API communicate with the controller via the
// command channel.
//////////////////////////////////////////////////////////////////////



func (p *P2PController) Init() P2PController  {
    p.loggingLevel = Silence
    p.keepRunning = true
    p.commandChannel = make(chan<- P2PCommandMessage, 1000)
    p.responseChannel = make(<-chan P2PResponseMessage, 1000)
    return p
}
// NetworkStart configures the network, starts the runloop
func (p * P2PController) StartNetwork()  {
    	// start listening on port given
	// dial into the peers
	/// start heartbeat process
    // Start the runloop
    go p.runloop()
}

func (p * P2PController) StartLogging(level uint8) {
    p.commandChannel <- P2PCommandMessage{Command: StartLogging, Parameters: [1]uint8{level}}
}
func (p * P2PController) StopLogging() {
    p.commandChannel <- P2PCommandMessage{Command: StopLogging, Parameters: nil}
}
func (p * P2PController) ChangeLogLevel(level uint8) {
    p.commandChannel <- P2PCommandMessage{Command: ChangeLogLevel, Parameters: [1]uint8{level}}
}

func (p * P2PController) AddPeer(address string)  {
    p.commandChannel <- P2PCommandMessage{Command: AddPeer, Parameters: [1]string{address}}
}

func (p * P2PController) NetworkStop()  {
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


// runloop is a goroutine that does all the heavy lifting
func (p * P2PController) runloop(address string)  {
 
    for p.keepRunning { // Run until we get the exit command
        // Process commands...
        for 0 > len(p.commandChannel) {
         handleCommand(command <- p.commandChannel )  
        }
        // For each peer, empty its network buffer (do recieves first)
        for _, connection := range peers {
            connection.ProcessNetworkMessages()
        }
        // For each peer, empty its outbound channel (and send to network.)
        for _, connection := range peers {
            connection.ProcessInChannel()
        }
    }   
}

func (p * P2PController) handleCommand(message P2PCommandMessage)  {
    switch message.Command {
    case AddPeer:
    BUGBUG
    case Shutdown:
        shutdown()
    case StartLogging:
        p.loggingLevel = uint8(message.Parameters[1])
    case StopLogging:
        p.loggingLevel = Silence
    case ChangeLogLevel:
        p.loggingLevel = uint8(message.Parameters[1])
    }
}

func (p * P2PController) shutdown()  {
    // Go thru peer list and shut down connections.
    BUGBUG
    p.keepRunning = false
}


