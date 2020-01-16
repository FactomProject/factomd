package simtest

import (
	"fmt"
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"

	"github.com/FactomProject/factomd/fnode"
)

/*
Test brainswapping F <-> L with no auditors

This test is useful for catching a failure scenario where the timing between
identity swap is off leading to a stall
*/
func TestLeaderBrainSwap(t *testing.T) {
	simulation.ResetSimHome(t) // clear out old test home
	for i := 0; i < 6; i++ {   // build config files for the test
		simulation.WriteConfigFile(i, i, "", t) // just write the minimal config
	}

	params := map[string]string{"--blktime": "10"}
	state0 := simulation.SetupSim("LLLFFF", params, 30, 0, 0, t)
	state3 := fnode.Get(3).State // Get node 2

	simulation.WaitForAllNodes(state0)
	simulation.WaitForBlock(state0, 6)

	// FIXME https://factom.atlassian.net/browse/FD-950 - setting batch > 1 can occasionally cause failure
	batches := 1 // use odd number to fulfill LFFFLL as end condition

	for batch := 0; batch < batches; batch++ {

		target := batch + 7

		change := fmt.Sprintf("ChangeAcksHeight = %v\n", target)

		if batch%2 == 0 {
			simulation.WriteConfigFile(1, 5, change, t) // Setup A brain swap between L1 and F5
			simulation.WriteConfigFile(5, 1, change, t)

			simulation.WriteConfigFile(2, 4, change, t) // Setup A brain swap between L2 and F4
			simulation.WriteConfigFile(4, 2, change, t)

		} else {
			simulation.WriteConfigFile(5, 5, change, t) // Un-Swap
			simulation.WriteConfigFile(1, 1, change, t)

			simulation.WriteConfigFile(4, 4, change, t)
			simulation.WriteConfigFile(2, 2, change, t)
		}

		simulation.WaitForBlock(state3, target)
		simulation.WaitMinutes(state3, 1)
	}

	simulation.WaitBlocks(state0, 1)
	simulation.AssertAuthoritySet(t, "LFFFLL")
	simulation.WaitForAllNodes(state0)
	simulation.ShutDownEverything(t)
}
