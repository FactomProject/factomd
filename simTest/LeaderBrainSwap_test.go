package simtest

import (
	"fmt"
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
			"--blktime":             "30",
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

		batchCount := 50

		// start the 6 nodes running  012345
		state0 := SetupSim("LLLFFF", params, batchCount+10, 0, 0, t)
		state3 := engine.GetFnodes()[3].State // Get node 2

		WaitForAllNodes(state0)
		WaitForBlock(state0, 9)

		for batch := 0; batch < batchCount; batch++ {

			t.Run(fmt.Sprintf("Wait For Identity Swap %v", batch), func(t *testing.T) {
				target := batch + 10

				change := fmt.Sprintf("ChangeAcksHeight = %v\n", target)

				if batch%2 == 0 {

					WriteConfigFile(1, 5, change, t) // Setup A brain swap between L1 and F5
					WriteConfigFile(5, 1, change, t)

					WriteConfigFile(2, 4, change, t) // Setup A brain swap between L2 and F4
					WriteConfigFile(4, 2, change, t)

				} else {
					WriteConfigFile(5, 5, change, t) // Un-Swap
					WriteConfigFile(1, 1, change, t)

					WriteConfigFile(4, 4, change, t)
					WriteConfigFile(2, 2, change, t)

				}
				WaitForBlock(state3, target)
				WaitMinutes(state3, 1)
			})
		}

		t.Run("Verify Network", func(t *testing.T) {
			WaitBlocks(state0, 1)
			// FIXME
			//AssertAuthoritySet(t, "LFFFLL")
			WaitForAllNodes(state0)
			ShutDownEverything(t)
		})

	})
}
