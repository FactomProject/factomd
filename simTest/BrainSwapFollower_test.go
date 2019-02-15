package simtest

import (
	. "github.com/FactomProject/factomd/testHelper"
	"os"
	"strconv"
	"testing"
)

var logName string = "simTest"

func TestBrainSwap(t *testing.T) {

	t.Run("Run sim to create entries", func(t *testing.T) {
		givenNodes := os.Getenv("GIVEN_NODES")
		factomHome := os.Getenv("FACTOM_HOME")
		maxBlocks, _ := strconv.ParseInt(os.Getenv("MAX_BLOCKS"), 10, 64)
		peers := os.Getenv("PEERS")

		if factomHome == "" {
			factomHome = ".sim/follower"
		}

		if maxBlocks == 0 {
			maxBlocks = 30
		}

		if peers == "" {
			peers = "127.0.0.1:38003"
		}

		if givenNodes == "" {
			givenNodes = "F"
		}

		// FIXME update to match test data
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
			//"--debuglog":            ".",
			"--logPort":          "37000",
			"--port":             "37001",
			"--controlpanelport": "37002",
			"--networkport":      "37003",
			"--peers":            peers,
			"--factomhome":       factomHome,
		}

		state0 := SetupSim(givenNodes, params, int(maxBlocks), 0, 0, t)
		state0.LogPrintf(logName, "GIVEN_NODES:%v", givenNodes)

		t.Run("Wait For Identity Swap", func(t *testing.T) {
			// NOTE: external scripts swap config files
			// during this time
			WaitForBlock(state0, 12)
			Followers--
			Leaders++
			WaitForAllNodes(state0)
			CheckAuthoritySet(t)
		})

		t.Run("Verify Network", func(t *testing.T) {
			WaitBlocks(state0, 3)
			Halt(t)
		})

	})
}
