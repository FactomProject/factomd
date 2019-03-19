package simtest

import (
	"os"
	"testing"

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

			globals.Params.FactomHome = dir + "/TestAddFNode"
			os.Setenv("FACTOM_HOME", globals.Params.FactomHome)

			t.Logf("Removing old run in %s", globals.Params.FactomHome)
			if err := os.RemoveAll(globals.Params.FactomHome); err != nil {
				t.Fatal(err)
			}

		})

		params := map[string]string{
			"--db":                  "LDB", // NOTE: using MAP causes an occasional error see FD-825
			"--network":             "LOCAL",
			"--net":                 "alot+",
			"--enablenet":           "true",
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

		state0 := SetupSim("LF", params, 15, 0, 0, t)

		t.Run("Create additional Fnode02", func(t *testing.T) {
			AddFNode()
			WaitBlocks(state0, 1)
		})

		t.Run("Verify Network", func(t *testing.T) {
			WaitForAllNodes(state0)
			AssertAuthoritySet(t, "LFF")
			ShutDownEverything(t)
		})

	})
}
