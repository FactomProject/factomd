// +build all 

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"fmt"
	"testing"

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
