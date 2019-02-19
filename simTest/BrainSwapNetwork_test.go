package simtest

import (
	"testing"

	"github.com/FactomProject/factomd/common/globals"
	. "github.com/FactomProject/factomd/testHelper"
)

/*
This test is part of a Network/Follower pair of tests used to test
brainswapping between 2 different versions of factomd

If you boot this simulator by itself - the tests will fail
*/
func TestBrainSwapNetwork(t *testing.T) {

	t.Run("Network Sim", func(t *testing.T) {
		maxBlocks := 30
		peers := "127.0.0.1:37003"
		// nodes usage  0123456 nodes 8 and 9 are in a separate sim of TestBrainSwapFollower
		given_Nodes := "LLLLAAA"
		outputNodes := "LLLAAFF"

		t.Run("Setup Config Files", func(t *testing.T) {
			ResetFactomHome(t, "network")

			// build config files for the test
			for i := 0; i < len(given_Nodes); i++ {
				WriteConfigFile(i, i, "", t)
			}
		})

		params := map[string]string{
			"--prefix":              "v0",
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
			"--logPort":          "38000",
			"--port":             "38001",
			"--controlpanelport": "38002",
			"--networkport":      "38003", // Listen on 'non-standard' unit test port
			"--peers":            peers,
			"--factomhome":       globals.Params.FactomHome,
		}

		state0 := SetupSim(given_Nodes, params, int(maxBlocks), 0, 0, t)

		t.Run("Wait For Identity Swap", func(t *testing.T) {
			WaitForAllNodes(state0)
			WriteConfigFile(9, 1, "ChangeAcksHeight = 10\n", t)
			WriteConfigFile(8, 4, "ChangeAcksHeight = 10\n", t)
			WaitForBlock(state0, 10)
			AdjustAuthoritySet(outputNodes)
		})

		t.Run("Verify Network", func(t *testing.T) {
			WaitBlocks(state0, 3)
			CheckAuthoritySet(t)
			WaitBlocks(state0, 1)
			Halt(t)
		})
	})
}
