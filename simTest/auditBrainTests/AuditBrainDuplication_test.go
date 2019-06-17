package auditBrainTests_test

import (
	"fmt"
	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"

	"strconv"
	"testing"
)

// Simulate bad node upgrade procedure where and audit server is duplicated / the initial one is left online or comes back online
// In this test one of the nodes does not have ChangeAcksHeight set
func TestAuditBrainDuplication1(t *testing.T) {
	t.Run("Run Brain Duplication Sim 1", func(t *testing.T) {
		t.Run("Setup Config Files", SetupConfigFiles)
		states := SetupNodes(t, "LLLAFF")
		duplicateIdentities(t, states, 0, 10)
		verifyNetworkAfterDup(t, states, 3)
	})
}

// In this test both of the nodes have ChangeAcksHeight set to the same value
// We need an extra audit node for this test otherwise we will get a stall
func TestAuditBrainDuplication2(t *testing.T) {
	t.Run("Run Brain Duplication Sim 2", func(t *testing.T) {
		t.Run("Setup Config Files", SetupConfigFiles)
		states := SetupNodes(t, "LLLAFFA")
		duplicateIdentities(t, states, 10, 10)
		verifyNetworkAfterDup(t, states, 5)
	})
}

// In this test both of the nodes have ChangeAcksHeight set to the same value
// We need an extra audit node for this test otherwise we will get a stall
func TestAuditBrainDuplication3(t *testing.T) {
	t.Run("Run Brain Duplication Sim 3", func(t *testing.T) {
		t.Run("Setup Config Files", SetupConfigFiles)
		states := SetupNodes(t, "LLLAFFA")
		duplicateIdentities(t, states, 10, 11)
		verifyNetworkAfterDup(t, states, 5)
	})
}

func duplicateIdentities(t *testing.T, states map[int]*state.State, node3ChangeAckHeight int, node5ChangeAckHeight int) {
	t.Run("Wait For Identity Duplication", func(t *testing.T) {
		WaitForBlock(states[0], 6)
		WaitForAllNodes(states[0])
		CheckAuthoritySet(t)

		// rewrite the config to have brainswaps
		changeAckHeight := ""
		if node3ChangeAckHeight > 0 {
			changeAckHeight = "ChangeAcksHeight = " + strconv.Itoa(node3ChangeAckHeight) + "\n"
		}
		WriteConfigFile(3, 3, changeAckHeight, t) // Setup A brain duplication from A3 to F5/A

		changeAckHeight = ""
		if node5ChangeAckHeight > 0 {
			changeAckHeight = "ChangeAcksHeight = " + strconv.Itoa(node5ChangeAckHeight) + "\n"
		}
		WriteConfigFile(3, 5, changeAckHeight, t) // Setup A brain duplication from A3 to F5/A
		WaitForBlock(states[0], 9)
		RunCmd("3") // make sure the Audit is lagging the audit if the heartbeats conflicts one will panic
		RunCmd("x")
		WaitForBlock(states[5], 10) // wait till 5 should have been brainswapped
		RunCmd("x")
		WaitBlocks(states[0], 2)
	})
}

func verifyNetworkAfterDup(t *testing.T, states map[int]*state.State, nodeExpectedToFail int) {
	t.Run("Verify Network", func(t *testing.T) {

		if states[nodeExpectedToFail].RunState < runstate.Stopping {
			t.Error(fmt.Sprintf("Node %d did didn't shut down", nodeExpectedToFail))
		}

		if nodeExpectedToFail == 3 {
			AdjustAuthoritySet("LLLFFA")
		} else {
			if states[3].RunState == runstate.Running {
				AdjustAuthoritySet("LLLAFFA")
			} else {
				// A rare but possible and observed outcome where both audit nodes shut down
				AdjustAuthoritySet("LLLFFFA")
				Halt(t)
				return
			}
		}
		CheckAuthoritySet(t)

		RunCmd("2") // Kill a leader to force an election
		RunCmd("x")
		WaitBlocks(states[0], 1)

		if nodeExpectedToFail == 3 {
			AdjustAuthoritySet("LLLFFL") // Node 5 should have been the only audit server and should be leader now
			CheckAuthoritySet(t)
		} else {
			AdjustAuthoritySet("LLLAFFL") // Node 3 is abstaining from the election so node 6 will now be a leader
			CheckAuthoritySet(t)

			RunCmd("6")
			RunCmd("x")
			states[6].ShutdownNode(1) // Shut down node 6

			WaitBlocks(states[0], 2)
			AdjustAuthoritySet("LLLLFFF") // Node 3 should have been elected now
			CheckAuthoritySet(t)
		}

		Halt(t)
	})
}
