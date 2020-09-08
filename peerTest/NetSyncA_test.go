package simtest

import (
	"testing"

	"github.com/FactomProject/factomd/testHelper/simulation"
)

/*
This test is the part A of a Network/Follower A/B pair of tests used to test
Just boots to test that Leader can sync over a network
*/
func TestNetSyncA(t *testing.T) {

	peers := "127.0.0.1:37003"
	simulation.ResetSimHome(t)

	params := map[string]string{
		"--db":               "LDB",
		"--network":          "LOCAL",
		"--nodename":         "TestA",
		"--net":              "alot+",
		"--enablenet":        "true",
		"--blktime":          "15",
		"--logPort":          "38000",
		"--port":             "38001",
		"--controlpanelport": "38002",
		"--networkport":      "38003",
		"--peers":            peers,
	}

	state0 := simulation.SetupSim("L", params, 7, 0, 0, t)

	simulation.WaitForBlock(state0, 6)
	simulation.Halt(t)
}
