package controlPanel_test

import (
	"testing"
	"time"

	. "github.com/FactomProject/factomd/controlPanel"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
)

// Enable for long test
var LongTest bool = false

func TestControlPanel(t *testing.T) {
	var i uint32
	connections := make(chan interface{})
	emptyState := CreateAndPopulateTestState()
	fnodes := make([]*state.State, 1)
	fnodes[0] = emptyState
	gitBuild := "Test Is Running"
	if LongTest {
		go ServeControlPanel(fnodes, connections, nil, gitBuild)
		for count := 0; count < 10; count++ {
			for i = 0; i < 2; i++ {
				PopulateConnectionChan(i, connections)
			}
			for i = 2; i > 0; i-- {
				PopulateConnectionChan(i, connections)
			}
		}
	}
}

func PopulateConnectionChan(total uint32, connections chan interface{}) {
	time.Sleep(5 * time.Second)
	var i uint32
	temp := make(map[string]p2p.ConnectionMetrics)
	for i = 0; i < total; i++ {
		peer := NewSeededP2PConnection(i)
		temp["{"+peer.PeerAddress+"}"] = *peer
	}
	connections <- temp
}
