// +build all 

package simtest

import (
	"testing"

	"github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
)

/*
Test brainswapping F <-> L  and F <-> A

follower and a leader + follower and an audit
at the same height in the same build
*/
func TestBrainSwap(t *testing.T) {
	ResetSimHome(t) // clear out old test home
	for i := 0; i < 6; i++ { // build config files for the test
		WriteConfigFile(i, i, "", t) // just write the minimal config
	}

	params := map[string]string{"--blktime": "15"}
	state0 := SetupSim("LLLAFF", params, 15, 0, 0, t)
	state3 := engine.GetFnodes()[3].State // Get node 3

	WaitForBlock(state0, 6)
	WaitForAllNodes(state0)

	// rewrite the config to orchestrate brainSwaps
	WriteConfigFile(2, 4, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between L2 and F4
	WriteConfigFile(4, 2, "ChangeAcksHeight = 10\n", t)
	WriteConfigFile(3, 5, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between A3 and F5
	WriteConfigFile(5, 3, "ChangeAcksHeight = 10\n", t)

	WaitForBlock(state0, 9)
	RunCmd("5") // make sure the follower is lagging the audit so he doesn't beat the auditor to the ID change and produce a heartbeat that will kill him
	RunCmd("x")
	WaitForBlock(state3, 10) // wait till should have 3 has brainswapped
	RunCmd("x")
	WaitBlocks(state0, 1)

	WaitForAllNodes(state0)
	AssertAuthoritySet(t, "LLFFLA")
	ShutDownEverything(t)
}
