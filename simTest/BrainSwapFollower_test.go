package simtest

import (
	"testing"

	"github.com/FactomProject/factomd/common/globals"

	. "github.com/FactomProject/factomd/testHelper"
)

/*
This test is part of a Network/Follower pair of tests used to test
brainswapping between 2 different versions of factomd

If you boot this simulator by itself - the simulation will not progress and will eventually time out
 */
func TestBrainSwapFollower(t *testing.T) {

	t.Run("Followers Sim", func(t *testing.T) {
		maxBlocks := 30
		peers := "127.0.0.1:38003"
		// thsi sim is  8 9
		given_Nodes := "FF"
		outputNodes := "LA"

		t.Run("Setup Config Files", func(t *testing.T) {
			ResetFactomHome(t, "follower")
			WriteConfigFile(9, 0, "", t)
			WriteConfigFile(8, 1, "", t)
		})

		params := map[string]string{
			"--prefix":              "v1",
			"--db":                  "LDB", // NOTE: using MAP causes an occasional error see FD-825
			"--network":             "LOCAL",
			"--net":                 "alot+",
			"--enablenet":           "true",
			"--blktime":             "30",
			"--startdelay":          "1",
			"--stdoutlog":           "out.txt",
			"--stderrlog":           "out.txt",
			"--checkheads":          "false",
			"--controlpanelsetting": "readwrite",
			//"--debuglog":            ".",
			"--logPort":             "37000",
			"--port":                "37001",
			"--controlpanelport":    "37002",
			"--networkport":         "37003",
			"--peers":               peers,
			"--factomhome":          globals.Params.FactomHome,
		}

		state0 := SetupSim(given_Nodes, params, int(maxBlocks), 0, 0, t)

		t.Run("Wait For Identity Swap", func(t *testing.T) {
			WaitForAllNodes(state0)
			WriteConfigFile(1, 0, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between L2 and F4
			WriteConfigFile(4, 1, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between L2 and F4
			WaitForBlock(state0, 9)
			RunCmd("1") // make sure the follower is lagging the audit so he doesn't beat the auditor to the ID change and produce a heartbeat that will kill him
			RunCmd("x")
			WaitForBlock(state0, 10) // wait till should have brainswapped
			RunCmd("x")
			AdjustAuthoritySet(outputNodes)
		})

		t.Run("Verify Network", func(t *testing.T) {
			WaitBlocks(state0, 1)
			CheckAuthoritySet(t)
			WaitBlocks(state0, 2)
			Halt(t)
		})
	})
}
