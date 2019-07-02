// +build simtest

package simtest

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants/servertype"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
)

// Test brainswapping a follower and an audit when the audit is lagging behind
func TestAuditBrainSwap(t *testing.T) {
	t.Run("Run Brain Swap Sim", func(t *testing.T) {
		t.Run("Setup Config Files", SetupConfigFiles)
		states := SetupNodes(t, "LLLAFF")
		swapIdentities(t, states)
		verifyNetworkAfterSwap(t, states)
	})
}

func swapIdentities(t *testing.T, states map[int]*state.State) bool {
	return t.Run("Wait For Identity Swap", func(t *testing.T) {
		WaitForBlock(states[0], 6)
		WaitForAllNodes(states[0])

		// rewrite the config to have brainswaps
		WriteConfigFile(3, 5, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between A3 and F5
		WriteConfigFile(5, 3, "ChangeAcksHeight = 10\n", t)
		WaitForBlock(states[0], 9)
		RunCmd("3") // make sure the Audit is lagging the audit if the heartbeats conflicts one will panic
		RunCmd("x")
		WaitForBlock(states[5], 10) // wait till 5 should have have brainswapped
		RunCmd("x")
		WaitBlocks(states[0], 1)
		WaitForAllNodes(states[0])
		CheckAuthoritySet(t)
	})
}

func verifyNetworkAfterSwap(t *testing.T, states map[int]*state.State) {
	t.Run("Verify Network", func(t *testing.T) {
		list := states[0].ProcessLists.Get(states[0].LLeaderHeight)

		serverType := servertype.GetServerType(list, states[3])
		if serverType != servertype.Follower {
			t.Error("Node 3 did not become a follower but a " + serverType)
		}

		serverType = servertype.GetServerType(list, states[5])
		if servertype.GetServerType(list, states[5]) != servertype.AuditServer {
			t.Error("Node 5 did not become an audit server but a " + serverType)
		}

		Halt(t)
	})
}
