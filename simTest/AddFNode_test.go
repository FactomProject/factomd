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
func TestAddFNode(t *testing.T) {

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

		t.Run("Create additional FNode", func(t *testing.T) {
			WaitForBlock(state0, 7)
			CloneFnodeData(2, 7, t)
			AddFNode()
		})

		state7 := engine.GetFnodes()[7].State // Get new node

		t.Run("Verify Network", func(t *testing.T) {
			WaitForBlock(state7, 7)
			AssertAuthoritySet(t, "LLLLLAAF")
			ShutDownEverything(t)
		})

	})
}
