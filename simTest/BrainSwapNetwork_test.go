package simtest

import (
	"os"
	"strconv"
	"testing"

	. "github.com/FactomProject/factomd/testHelper"
)

func TestBrainSwapNetwork(t *testing.T) {

	t.Run("Create Authority Set", func(t *testing.T) {
		givenNodes := os.Getenv("GIVEN_NODES")
		factomHome := os.Getenv("FACTOM_HOME")
		maxBlocks, _ := strconv.ParseInt(os.Getenv("MAX_BLOCKS"), 10, 64)
		peers := os.Getenv("PEERS")

		if factomHome == "" {
			factomHome = ".sim/network"
		}

		if maxBlocks == 0 {
			maxBlocks = 30
		}

		if peers == "" {
			peers = "127.0.0.1:37003"
		}

		if givenNodes == "" {
			givenNodes = "LLLLAAA"
		}

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
			"--logPort":          "38000",
			"--port":             "38001",
			"--controlpanelport": "38002",
			"--networkport":      "38003",
			"--peers":            peers,
			"--factomhome":       factomHome,
		}

		state0 := SetupSim(givenNodes, params, int(maxBlocks), 0, 0, t)

		t.Run("Wait For Identity Swap", func(t *testing.T) {
			WaitForBlock(state0, 12)
			// brainswap leader
			Followers++
			Leaders--
			// brainswap auditor
			Followers++
			Audits--
			WaitForAllNodes(state0)
		})

		t.Run("Verify Network", func(t *testing.T) {
			CheckAuthoritySet(t)
			WaitBlocks(state0, 3)
			Halt(t)
		})

	})
}
