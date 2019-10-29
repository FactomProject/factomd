package simtest

import (
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
)

/*
This test is the part A of a Network/Follower A/B pair of tests used to test
Just boots to test that Leader can sync over a network
*/
func TestNetSyncA(t *testing.T) {

	peers := "127.0.0.1:37003"
	ResetSimHome(t)

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

	state0 := SetupSim("L", params, 13, 0, 0, t)
	WaitForBlock(state0, 13)
	Halt(t)
}
