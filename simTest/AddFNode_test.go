package simtest

import (
	"os"
	"testing"

	"github.com/FactomProject/factomd/engine"

	"github.com/FactomProject/factomd/common/globals"
	. "github.com/FactomProject/factomd/testHelper"
)

/*
 */
func TestAddingFNode(t *testing.T) {

	t.Run("Run Sim", func(t *testing.T) {

		t.Run("Setup Config Files", func(t *testing.T) {
			dir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			globals.Params.FactomHome = dir + "/TestAddFnode"
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

		// FIXME: can we replace w/ this?
		// somehow this causes a brainswap not to happen
		//ResetFactomHome(t, "TestAddingFNode")

		params := map[string]string{
			"--db":                  "LDB", // NOTE: using MAP causes an occasional error see FD-825
			"--network":             "LOCAL",
			"--net":                 "alot+",
			"--enablenet":           "false",
			"--blktime":             "15",
			"--startdelay":          "1",
			"--stdoutlog":           "out.txt",
			"--stderrlog":           "out.txt",
			"--checkheads":          "false",
			"--controlpanelsetting": "readwrite",
			"--debuglog":            ".",
			"--logPort":             "37000",
			"--port":                "37001",
			"--controlpanelport":    "37002",
			"--networkport":         "37003",
			"--peers":               "127.0.0.1:38003",
			"--factomhome":          globals.Params.FactomHome,
		}

		state0 := SetupSim("LLLLLAA", params, 25, 1, 1, t)

		t.Run("Cause auditor to be promoted", func(t *testing.T) {
			// FIXME actually do this
			RunCmd("1")
			RunCmd("w")
			RunCmd("s")
			apiRegex := "EOM.*9/.*minute 1"
			SetOutputFilter(apiRegex)
		})

		t.Run("Create additional FNode", func(t *testing.T) {
			WaitForBlock(state0, 7)
			CloneFnodeData(2, 7, t)
			AddFNode() // REVIEW: somehow the way the new node is added causes it to lag
		})

		t.Run("Filter out Target ACK from new node", func(t *testing.T) {
			// FIXME actually do this
			//RunCmd("7")
			//RunCmd("w")
			//RunCmd("s")
			//apiRegex := "EOM.*9/.*minute 1"
			//SetOutputFilter(apiRegex)
		})

		state7 := engine.GetFnodes()[7].State // Get new node

		t.Run("Wait For Identity Swap", func(t *testing.T) { // REVIEW: setting changeAcksHeight to 7 causes a stall leader swap fails b/c new follower  is not up-to-date
			//WaitForBlock(state7, 7)
			// FIXME: swap the proper nodes
			WriteConfigFile(2, 7, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between L2 and F4
			WriteConfigFile(7, 2, "ChangeAcksHeight = 10\n", t)
			WaitForBlock(state7, 12)
		})

		t.Run("Verify Network", func(t *testing.T) {
			WaitForAllNodes(state0)
			// FIXME: somehow brainswap isn't working
			//AssertAuthoritySet(t, "LLLLLAFA")
			//ShutDownEverything(t)
		})

	})
}
