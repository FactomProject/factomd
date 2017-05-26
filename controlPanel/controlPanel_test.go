package controlPanel_test

import (
	"fmt"
	"testing"
	"time"

	. "github.com/FactomProject/factomd/controlPanel"
	"github.com/FactomProject/factomd/p2p"
	//"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/common/primitives"
	//"github.com/FactomProject/factomd/common/primitives/random"
	. "github.com/FactomProject/factomd/testHelper"
)

var _ = fmt.Sprintf("")

// Enable for long test
var LongTest bool = false

func TestFactoidHas(t *testing.T) {
	rc := new(LastDirectoryBlockTransactions)
	rc.FactoidTransactions = append(rc.FactoidTransactions)
	for i := 0; i < 10; i++ {
		addTrans(rc)
	}

	for i := 0; i < len(rc.FactoidTransactions); i++ {
		h, _ := primitives.HexToHash(rc.FactoidTransactions[i].TxID)
		if !rc.ContainsTrans(h) {
			t.Error("This should be true")
		}
	}

	for i := 0; i < len(rc.Entries); i++ {
		h, _ := primitives.HexToHash(rc.Entries[i].Hash)
		if !rc.ContainsEntry(h) {
			t.Error("This should be true")
		}
	}
}

func addTrans(rc *LastDirectoryBlockTransactions) {
	rc.FactoidTransactions = append(rc.FactoidTransactions, struct {
		TxID         string
		Hash         string
		TotalInput   string
		Status       string
		TotalInputs  int
		TotalOutputs int
	}{primitives.RandomHash().String(), primitives.RandomHash().String(), "1", "Confirmed", 1, 1})

	e := new(EntryHolder)
	e.Hash = primitives.RandomHash().String()
	rc.Entries = append(rc.Entries, *e)
}

func TestControlPanel(t *testing.T) {
	if LongTest {
		var i uint32
		connections := make(chan interface{})
		emptyState := CreateAndPopulateTestState()

		gitBuild := "Test Is Running"
		go ServeControlPanel(emptyState.ControlPanelChannel, emptyState, connections, nil, gitBuild)
		emptyState.CopyStateToControlPanel()
		for count := 0; count < 1000; count++ {
			for i = 0; i < 5; i++ {
				PopulateConnectionChan(i, connections)

			}
			for i = 5; i > 0; i-- {
				PopulateConnectionChan(i, connections)
			}
		}
	}
}

func PopulateConnectionChan(total uint32, connections chan interface{}) {
	time.Sleep(3 * time.Second)
	var i uint32
	temp := make(map[string]p2p.ConnectionMetrics)
	for i = 0; i < total; i++ {
		peer := NewSeededP2PConnection(i)
		if i%2 == 0 {
			peer.MomentConnected = time.Now().Add(-(time.Duration(i)) * time.Hour)
		} else {
			peer.MomentConnected = time.Now().Add(-(time.Duration(i)) * time.Minute)
		}
		temp["{"+peer.PeerAddress+"}"] = *peer
	}
	connections <- temp
}
