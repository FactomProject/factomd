package simtest

import (
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"
)

/*
This test is part of a Network/Follower pair of tests used to test
brainswapping between 2 different versions of factomd

If you boot this simulator by itself - the simulation will not progress and will eventually time out
*/
func TestBrainSwapB(t *testing.T) {

	maxBlocks := 30
	peers := "127.0.0.1:38003"
	// this sim starts with identities 8 & 9
	givenNodes := "FF"
	outputNodes := "LA"

	simulation.ResetSimHome(t)
	simulation.WriteConfigFile(9, 0, "", t)
	simulation.WriteConfigFile(8, 1, "", t)

	params := map[string]string{
		"--db":               "LDB",
		"--network":          "LOCAL",
		"--nodename":         "TestB",
		"--net":              "alot+",
		"--enablenet":        "true",
		"--blktime":          "30",
		"--logPort":          "37000",
		"--port":             "37001",
		"--controlpanelport": "37002",
		"--networkport":      "37003",
		"--peers":            peers,
	}

	state0 := simulation.SetupSim(givenNodes, params, int(maxBlocks), 0, 0, t)

	simulation.WaitForAllNodes(state0)
	simulation.WriteConfigFile(1, 0, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between L2 and F4
	simulation.WriteConfigFile(4, 1, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between L2 and F4
	simulation.WaitForBlock(state0, 9)
	simulation.RunCmd("1") // make sure the follower is lagging the audit so he doesn't beat the auditor to the ID change and produce a heartbeat that will kill him
	simulation.RunCmd("x")
	simulation.WaitForBlock(state0, 10) // wait till should have brainswapped
	simulation.RunCmd("x")
	simulation.AdjustAuthoritySet(outputNodes)

	simulation.WaitBlocks(state0, 3)
	simulation.AssertAuthoritySet(t, outputNodes)
	simulation.WaitBlocks(state0, 1)
	simulation.Halt(t)
}
