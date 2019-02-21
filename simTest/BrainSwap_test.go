package simtest

import (
	"os"
	"testing"

	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
)

/*
Test brainswapping F <-> L  and F <-> A

follower and a leader + follower and an audit
at the same height in the same build
*/
func TestBrainSwap(t *testing.T) {

	t.Run("Run Sim", func(t *testing.T) {

		t.Run("Setup Config Files", func(t *testing.T) {
			dir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			globals.Params.FactomHome = dir + "/TestBrainSwap"
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
		state0 := SetupSim("LLLAFF", params, 15, 0, 0, t)
		state2 := engine.GetFnodes()[2].State // Get node 2
		state3 := engine.GetFnodes()[3].State // Get node 3
		state4 := engine.GetFnodes()[4].State // Get node 4
		state5 := engine.GetFnodes()[5].State // Get node 5

		t.Run("Wait For Identity Swap", func(t *testing.T) {
			WaitForBlock(state0, 6)
			WaitForAllNodes(state0)
			// rewrite the config to have brainswaps

			WriteConfigFile(2, 4, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between L2 and F4
			WriteConfigFile(4, 2, "ChangeAcksHeight = 10\n", t)
			WriteConfigFile(3, 5, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between A3 and F5
			WriteConfigFile(5, 3, "ChangeAcksHeight = 10\n", t)
			WaitForBlock(state0, 9)
			RunCmd("5") // make sure the follower is lagging the audit so he doesn't beat the auditor to the ID change and produce a heartbeat that will kill him
			RunCmd("x")
			WaitForBlock(state3, 10) // wait till should have 3 has brainswapped
			RunCmd("x")
			WaitBlocks(state0, 1)
			WaitForAllNodes(state0)
			CheckAuthoritySet(t)
		})

		t.Run("Verify Network", func(t *testing.T) {

			if state2.Leader {
				t.Error("Node 2 did not become a follower")
			}
			if state3.Leader {
				t.Error("Node 3 did not become a follower")
			}
			if !state4.Leader {
				t.Error("Node 4 did not become a leader")
			}

			list := state0.ProcessLists.Get(state0.LLeaderHeight)
			foundAudit, _ := list.GetAuditServerIndexHash(state5.GetIdentityChainID())
			if !foundAudit {
				t.Error("Node 5 did not become an audit server")
			}

			Halt(t)
		})

	})
}
