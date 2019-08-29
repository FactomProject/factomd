package simtest

import (
	"testing"

	. "github.com/FactomProject/factomd/testHelper"
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

	ResetSimHome(t)
	WriteConfigFile(9, 0, "", t)
	WriteConfigFile(8, 1, "", t)

	params := map[string]string{
		"--db":               "LDB",
		"--network":          "LOCAL",
		"--net":              "alot+",
		"--enablenet":        "true",
		"--blktime":          "30",
		"--logPort":          "37000",
		"--port":             "37001",
		"--controlpanelport": "37002",
		"--networkport":      "37003",
		"--peers":            peers,
	}

	state0 := SetupSim(givenNodes, params, int(maxBlocks), 0, 0, t)

	WaitForAllNodes(state0)
	WriteConfigFile(1, 0, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between L2 and F4
	WriteConfigFile(4, 1, "ChangeAcksHeight = 10\n", t) // Setup A brain swap between L2 and F4
	WaitForBlock(state0, 9)
	RunCmd("1") // make sure the follower is lagging the audit so he doesn't beat the auditor to the ID change and produce a heartbeat that will kill him
	RunCmd("x")
	WaitForBlock(state0, 10) // wait till should have brainswapped
	RunCmd("x")
	AdjustAuthoritySet(outputNodes)

	WaitBlocks(state0, 3)
	AssertAuthoritySet(t, outputNodes)
	WaitBlocks(state0, 1)
	Halt(t)
}
