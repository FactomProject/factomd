// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"

	"github.com/FactomProject/factomd/common/messages"

	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

var _ = fmt.Print

func TestProcessListString(t *testing.T) {
	// The string function is called in some unit tests, and lines that show offline nodes is sometimes hit. This
	// ensures coverage is consistent, despite it just being a String() call
	state := testHelper.CreateEmptyTestState()
	pl := NewProcessList(state, nil, 0)
	pl.VMs[0].List = append(pl.VMs[0].List, nil)
	pl.AddFedServer(primitives.NewHash([]byte("one")))
	pl.AddAuditServer(primitives.NewHash([]byte("two")))
	pl.AddFedServer(primitives.NewHash([]byte("three")))

	for _, f := range pl.FedServers {
		f.SetOnline(false)
	}

	var _ = pl.String()
}

func TestProcessListMisc(t *testing.T) {
	// The string function is called in some unit tests, and lines that show offline nodes is sometimes hit. This
	// ensures coverage is consistent, despite it just being a String() call
	state := testHelper.CreateEmptyTestState()
	pl := NewProcessList(state, nil, 0)
	pl.VMs[0].List = append(pl.VMs[0].List, nil)
	pl.AddFedServer(primitives.NewHash([]byte("one")))
	pl.AddAuditServer(primitives.NewHash([]byte("two")))
	pl.AddFedServer(primitives.NewHash([]byte("three")))

	vmi := pl.VMIndexFor([]byte("one"))
	if vmi != 0 {
		t.Error("VMIndex should be 0")
	}

	fs := pl.FedServerFor(0, []byte("one"))
	if fs == nil {
		t.Error("No fed server associated with minute 0 byte slice")
	}

	wasReset := pl.Reset()
	if !wasReset {
		t.Error("Process List Reset did not work")
	}

	pl.TrimVMList(0, 0)
}

func TestProcessListTrim(t *testing.T) {
	state := testHelper.CreateEmptyTestState()
	pl := NewProcessList(state, nil, 0)

	// Test various PL trim scenarios
	//	We cannot trim below the processed height.
	testPLListTrim(t, pl, 100, 50, 45, 100)
	// Cannot trim above the total height
	testPLListTrim(t, pl, 2, 0, 5, 2)

	testPLListTrim(t, pl, 50, 35, 35, 35)
	testPLListTrim(t, pl, 50, 0, 0, 0)
	testPLListTrim(t, pl, 26, 25, 25, 25)
}

func testPLListTrim(t *testing.T, pl *ProcessList, total, processedHeight, trimHeight, expHeight int) {
	pl.VMs[0].List = []interfaces.IMsg{}
	for i := 0; i < total; i++ {
		pl.VMs[0].List = append(pl.VMs[0].List, &messages.Bounce{})
	}
	pl.VMs[0].Height = processedHeight // Set a height that is "processed"

	pl.TrimVMList(uint32(trimHeight), 0)
	if len(pl.VMs[0].List) != expHeight {
		t.Errorf("PLTrim left height of %d. Expected %d", len(pl.VMs[0].List), expHeight)
	}
}

func TestServerMap(t *testing.T) {
	// The string function is called in some unit tests, and lines that show offline nodes is sometimes hit. This
	// ensures coverage is consistent, despite it just being a String() call
	state := testHelper.CreateEmptyTestState()
	pl := NewProcessList(state, nil, 1)
	pl.FedServers = append(pl.FedServers, nil)
	pl.FedServers = append(pl.FedServers, nil)
	pl.FedServers = append(pl.FedServers, nil)
	pl.FedServers = append(pl.FedServers, nil)
	pl.FedServers = append(pl.FedServers, nil)
	pl.FedServers = append(pl.FedServers, nil)
	pl.FedServers = append(pl.FedServers, nil)
	pl.FedServers = append(pl.FedServers, nil)
	pl.FedServers = append(pl.FedServers, nil)
	pl.FedServers = append(pl.FedServers, nil)
	pl.FedServers = append(pl.FedServers, nil)
	pl.FedServers = append(pl.FedServers, nil)

	pl.DBHeight = 100000
	pl.MakeMap()
	pl.PrintMap()

	vmIdx := FedServerVM(pl.ServerMap, len(pl.FedServers), 3, 10)
	//fmt.Println("VM Index", vmIdx)
	if vmIdx != 2 {
		t.Error("Unexpected VM index in ServerMap")
	}
}

/*
 * A panic shouldn't occurs when a node receives a future ACK messages with the same ID.
 * This can occurs during a brain swap.
 */
func TestBrainSwapAck(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()

	// init ack message
	m := new(messages.Ack)
	m.SetLeaderChainID(primitives.RandomHash())
	m.DBHeight = state.GetLLeaderHeight() + 1
	m.Timestamp = primitives.NewTimestampNow()
	m.MessageHash = primitives.RandomHash()
	m.SerialHash = primitives.RandomHash()

	ack := state.NewAck(m, state.Balancehash).(*messages.Ack)

	// Change the timestamp of the ack message so the SaltNumber if different, i.e. the secret number used to detect multiple servers with the same ID
	ack.Timestamp = primitives.NewTimestampFromSeconds(primitives.NewTimestampNow().GetTimeSecondsUInt32() + 10)

	// change the db height such that the ack of the brain swap will be processed in the wrong process list.
	ack.DBHeight = state.LLeaderHeight + 1

	state.ProcessLists.Get(ack.DBHeight).AddToProcessList(state, ack, m)
}
