package simtest

import (
	"github.com/FactomProject/factomd/common/globals"
	"testing"

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
		peers := "127.0.0.1:37003" // follower peer uses standard unit testing port
		givenNodes := "LLLLAAA"
		outputNodes := "LLLAAFF"

		t.Run("Setup Config Files", func(t *testing.T) {
			ResetFactomHome(t, "network")

			// build config files for the test
			for i := 0; i < len(givenNodes); i++ {
				WriteConfigFile(i, i, "ChangeAcksHeight = 1\n", t)
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
			"--logPort":             "38000",
			"--port":                "38001",
			"--controlpanelport":    "38002",
			"--networkport":         "38003", // Listen on 'non-standard' unit test port
			"--peers":               peers,
			"--factomhome":          globals.Params.FactomHome,
		}

		state0 := SetupSim(givenNodes, params, int(maxBlocks), 0, 0, t)

		t.Run("Wait For Identity Swap", func(t *testing.T) {
			WaitForAllNodes(state0)
			WriteConfigFile(9, 1, "ChangeAcksHeight = 10\n", t)
			WriteConfigFile(8, 4, "ChangeAcksHeight = 10\n", t)
			WaitForBlock(state0, 10)
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
