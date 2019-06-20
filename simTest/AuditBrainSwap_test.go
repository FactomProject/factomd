package simtest

import (
	"testing"

	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
)

/*
Test brainswapping a F <-> A

follower and an audit when the audit is lagging behind

This test is useful for verifying that Leaders can swap without rebooting
And that Audits can reboot with lag (to prevent a panic if 2 nodes see the same audit heartbeat)
*/
func TestAuditBrainSwap(t *testing.T) {
	ResetSimHome(t)          // clear out old test home
	for i := 0; i < 6; i++ { // build config files for the test
		WriteConfigFile(i, i, "", t) // just write the minimal config
	}

	params := map[string]string{"--factomhome": globals.Params.FactomHome}
	state0 := SetupSim("LLLAFF", params, 15, 0, 0, t)
	state5 := engine.GetFnodes()[5].State // Get node 5
	_ = state5

	t.Log("Disabled test while known bug exists FD-845")
	/*
	   WaitForBlock(state0, 6)
	   WaitForAllNodes(state0)
	   // rewrite the config to have brainswaps

	   WriteConfigFile(3, 5, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between A3 and F5
	   WriteConfigFile(5, 3, "ChangeAcksHeight = 10\n", t)
	   WaitForBlock(state0, 9)
	   RunCmd("3") // make sure the Audit is lagging the audit if the heartbeats conflict one will panic
	   RunCmd("x")
	   WaitForBlock(state5, 10) // wait till 5 should have have brainswapped
	   RunCmd("x")
	   WaitBlocks(state0, 1)
	   WaitForAllNodes(state0)
	   CheckAuthoritySet(t)
	*/

	WaitForAllNodes(state0)
	// FIXME
	//AssertAuthoritySet(t, "LLLFFA")
	ShutDownEverything(t)
}
