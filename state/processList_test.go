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
	pl := NewProcessList(state, nil, 1)
	pl.VMs[0].List = append(pl.VMs[0].List, nil)
	pl.AddFedServer(primitives.NewHash([]byte("one")))
	pl.AddAuditServer(primitives.NewHash([]byte("two")))
	pl.AddFedServer(primitives.NewHash([]byte("three")))

	for _, f := range pl.FedServers {
		f.SetOnline(false)
	}

	var _ = pl.String()
}
