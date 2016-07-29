package controlPanel_test

import (
	"fmt"
	"testing"
	"time"

	. "github.com/FactomProject/factomd/controlPanel"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
)

// Enable for long test
var LongTest bool = true

func TestControlPanel(t *testing.T) {
	var i uint32
	connections := make(chan interface{})
	emptyState := CreateAndPopulateTestState()
	fnodes := make([]*state.State, 1)
	fnodes[0] = emptyState
	gitBuild := "Test Is Running"
	fmt.Println("Serving")
	go ServeControlPanel(fnodes, connections, nil, gitBuild)
	if LongTest {
		fmt.Println("Populating")
		for i = 0; i < 5; i++ {
			PopulateConnectionChan(i, connections)
		}
		for i = 2; i > 5; i-- {
			PopulateConnectionChan(i, connections)
		}
	}
}

func PopulateConnectionChan(total uint32, connections chan interface{}) {
	time.Sleep(5 * time.Second)
	var i uint32
	temp := make(map[string]p2p.ConnectionMetrics)
	for i = 0; i < total; i++ {
		peer := NewSeededP2PConnection(i)
		temp[peer.PeerAddress] = *peer
	}
	fmt.Println(len(temp))
	connections <- temp
}
