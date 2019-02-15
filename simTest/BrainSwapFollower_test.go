package simtest

import (
	. "github.com/FactomProject/factomd/testHelper"
	"os"
	"strconv"
	"testing"
)

var logName string = "simTest"

func TestBrainSwapFollower(t *testing.T) {

	t.Run("Create Followers On Network", func(t *testing.T) {
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
			givenNodes = "FF"
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
			// FIXME: replace external scripts swap config files
			WaitForBlock(state0, 12)
			// brainswap leader
			Followers--
			Leaders++
			// brainswap auditor
			Followers--
			Audits++
			WaitForAllNodes(state0)
		})

		t.Run("Verify Network", func(t *testing.T) {
			CheckAuthoritySet(t)
			WaitBlocks(state0, 3)
			Halt(t)
		})

	})
}
