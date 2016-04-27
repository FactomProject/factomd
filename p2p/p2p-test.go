package main

import (
	"fmt"
	"github.com/FactomProject/factomd/p2p"
)

// Note - we only need to expose the Listen port, so this can be tested
// on Caladan by simply connecting to the Caladan Cluster at a URL and Port.
// It will be effectively a random connection (we might connect to ourselves)
// So will need to check and release if we end up talking to ourself.

const (
	fullPeerManagement = iota
	simpleMinimalTest
)

func main() {
	P2PCurrentLoggingLevel = Verbose
	mode := simpleMinimalTest
	switch mode {
	case simpleMinimalTest:
		simpleMinimalTest()
	case fullPeerManagement:
		fullPeerManagement()
	}
}

func fullPeerManagement() {
	p2pController := new(P2PController).Init().(*P2PController)
	p2pController.StartLogging(Verbose)
	p2pController.StartNetwork()
	// Send some custom messages.

}

func simpleMinimalTest() {
	nodeA := new(P2PConnection).Init()
	nodeB := new(P2PConnection).Init()
	count :=0
	var	message string

	for {
			message = fmt.Sprintf("Heartbeat #%d", count)
			nodeA.SimpleSend([]byte(message))
			nodeB.SimpleSend([]byte(message))
			nodeA.ProcessNetworkMessages()
			nodeA.ProcessInChannel()
			nodeB.ProcessNetworkMessages()
			nodeB.ProcessInChannel()
		}

	}
}
