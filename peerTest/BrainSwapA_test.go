package simtest

import (
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"
)

/*
This test is part of a Network/Follower pair of tests used to test
brainswapping between 2 different versions of factomd

If you boot this simulator by itself - the tests will fail
*/
func TestBrainSwapA(t *testing.T) {

	maxBlocks := 30
	peers := "127.0.0.1:37003"
	// nodes usage 0123456 nodes 8 and 9 are in a separate sim of TestBrainSwapB
	givenNodes := "LLLLAAA"
	outputNodes := "LFLLFAA"

	simulation.ResetSimHome(t)

	// build config files for the test
	for i := 0; i < len(givenNodes); i++ {
		simulation.WriteConfigFile(i, i, "", t)
	}

	params := map[string]string{
		"--db":               "LDB",
		"--network":          "LOCAL",
		"--net":              "alot+",
		"--enablenet":        "true",
		"--blktime":          "30",
		"--logPort":          "38000",
		"--port":             "38001",
		"--controlpanelport": "38002",
		"--networkport":      "38003",
		"--peers":            peers,
	}

	state0 := simulation.SetupSim(givenNodes, params, int(maxBlocks), 0, 0, t)

	simulation.WaitForAllNodes(state0)
	simulation.WriteConfigFile(9, 1, "ChangeAcksHeight = 10\n", t)
	simulation.WriteConfigFile(8, 4, "ChangeAcksHeight = 10\n", t)
	simulation.WaitForBlock(state0, 10)
	simulation.AdjustAuthoritySet(outputNodes)

	simulation.WaitBlocks(state0, 3)
	simulation.AssertAuthoritySet(t, outputNodes)
	simulation.WaitBlocks(state0, 1)
	simulation.Halt(t)

}
