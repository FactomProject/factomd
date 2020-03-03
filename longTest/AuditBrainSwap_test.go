package longtest

import (
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"

	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/fnode"
)

/*
Test brainswapping a F <-> A

follower and an audit when the audit is lagging behind

This test is useful for verifying that Leaders can swap without rebooting
And that Audits can reboot with lag (to prevent a panic if 2 nodes see the same audit heartbeat)
*/
func TestAuditBrainSwap(t *testing.T) {
	simulation.ResetSimHome(t) // clear out old test home
	for i := 0; i < 6; i++ {   // build config files for the test
		simulation.WriteConfigFile(i, i, "", t) // just write the minimal config
	}

	params := map[string]string{"--factomhome": globals.Params.FactomHome}
	state0 := simulation.SetupSim("LLLAFF", params, 15, 0, 0, t)
	state5 := fnode.Get(5).State // Get node 5
	_ = state5

	simulation.WaitForBlock(state0, 6)
	simulation.WaitForAllNodes(state0)

	// rewrite the config to have brainswaps
	simulation.WriteConfigFile(3, 5, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between A3 and F5
	simulation.WriteConfigFile(5, 3, "ChangeAcksHeight = 10\n", t)
	simulation.WaitForBlock(state0, 9)
	simulation.RunCmd("3") // make sure the Audit is lagging the audit if the heartbeats conflict one will panic
	simulation.RunCmd("x")
	simulation.WaitForBlock(state5, 10) // wait till 5 should have have brainswapped
	simulation.RunCmd("x")
	simulation.WaitBlocks(state0, 1)
	simulation.WaitForAllNodes(state0)
	simulation.CheckAuthoritySet(t)

	simulation.WaitForAllNodes(state0)
	simulation.AssertAuthoritySet(t, "LLLFFA")
	simulation.ShutDownEverything(t)
}
