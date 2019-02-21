package simtest

import (
	"os"
	"testing"

	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
)

/*
Test brainswapping F <-> L with no auditors

This test is useful for catching a failure scenario where the timing between
identity swap is off leading to a stall
*/
func TestLeaderBrainSwap(t *testing.T) {

	t.Run("Run Sim", func(t *testing.T) {

		t.Run("Setup Config Files", func(t *testing.T) {
			dir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			globals.Params.FactomHome = dir + "/TestLeadersOnlyBrainSwap"
			os.Setenv("FACTOM_HOME", globals.Params.FactomHome)

			t.Logf("Removing old run in %s", globals.Params.FactomHome)
			if err := os.RemoveAll(globals.Params.FactomHome); err != nil {
				t.Fatal(err)
			}

			// build config files for the test
			for i := 0; i < 6; i++ {
				WriteConfigFile(i, i, "", t) // just write the minimal config
			}
		})

		params := map[string]string{
			"--db":                  "LDB", // NOTE: using MAP causes an occasional error see FD-825
			"--network":             "LOCAL",
			"--net":                 "alot+",
			"--enablenet":           "true",
			"--blktime":             "10",
			"--startdelay":          "1",
			"--stdoutlog":           "out.txt",
			"--stderrlog":           "out.txt",
			"--checkheads":          "false",
			"--controlpanelsetting": "readwrite",
			"--debuglog":            ".",
			"--logPort":             "38000",
			"--port":                "38001",
			"--controlpanelport":    "38002",
			"--networkport":         "38003",
			"--peers":               "127.0.0.1:37003",
			"--factomhome":          globals.Params.FactomHome,
		}

		// start the 6 nodes running  012345
		state0 := SetupSim("LLLFFF", params, 15, 0, 0, t)
		state1 := engine.GetFnodes()[1].State // Get node 1
		state2 := engine.GetFnodes()[2].State // Get node 2
		state3 := engine.GetFnodes()[3].State // Get node 2
		state4 := engine.GetFnodes()[4].State // Get node 4
		state5 := engine.GetFnodes()[5].State // Get node 5

		t.Run("Wait For Identity Swap", func(t *testing.T) {
			WaitForBlock(state0, 6)
			WaitForAllNodes(state0)
			// rewrite the config to have brainswaps

			WriteConfigFile(1, 5, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between L1 and F5
			WriteConfigFile(5, 1, "ChangeAcksHeight = 10\n", t)

			WriteConfigFile(2, 4, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between L2 and F4
			WriteConfigFile(4, 2, "ChangeAcksHeight = 10\n", t)

			WaitForBlock(state3, 10)

			WaitBlocks(state0, 1)
			WaitForAllNodes(state0)
			CheckAuthoritySet(t)
		})

		t.Run("Verify Network", func(t *testing.T) {

			if state1.Leader {
				t.Error("Node 1 did not become a follower")
			}
			if state2.Leader {
				t.Error("Node 2 did not become a follower")
			}
			if !state4.Leader {
				t.Error("Node 4 did not become a leader")
			}
			if !state5.Leader {
				t.Error("Node 5 did not become a leader")
			}

			Halt(t)
		})

	})
}
