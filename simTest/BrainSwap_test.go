package simtest

import (
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"

	"github.com/FactomProject/factomd/fnode"
)

/*
Test brainswapping F <-> L  and F <-> A

follower and a leader + follower and an audit
at the same height in the same build
*/
func TestBrainSwap(t *testing.T) {
	simulation.ResetSimHome(t) // clear out old test home
	for i := 0; i < 6; i++ {   // build config files for the test
		simulation.WriteConfigFile(i, i, "", t) // just write the minimal config
	}

	params := map[string]string{"--blktime": "15"}
	state0 := simulation.SetupSim("LLLAFF", params, 15, 0, 0, t)
	state3 := fnode.Get(3).State // Get node 3

	simulation.WaitForBlock(state0, 6)
	simulation.WaitForAllNodes(state0)

	// rewrite the config to orchestrate brainSwaps
	simulation.WriteConfigFile(2, 4, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between L2 and F4
	simulation.WriteConfigFile(4, 2, "ChangeAcksHeight = 10\n", t)
	simulation.WriteConfigFile(3, 5, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between A3 and F5
	simulation.WriteConfigFile(5, 3, "ChangeAcksHeight = 10\n", t)

	simulation.WaitForBlock(state0, 9)
	simulation.RunCmd("5") // make sure the follower is lagging the audit so he doesn't beat the auditor to the ID change and produce a heartbeat that will kill him
	simulation.RunCmd("x")
	simulation.WaitForBlock(state3, 10) // wait till should have 3 has brainswapped
	simulation.RunCmd("x")
	simulation.WaitBlocks(state0, 1)

	simulation.WaitForAllNodes(state0)
	simulation.AssertAuthoritySet(t, "LLFFLA")
	simulation.ShutDownEverything(t)
}
