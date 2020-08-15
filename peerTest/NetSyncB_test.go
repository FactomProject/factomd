package simtest

import (
	"testing"

	. "github.com/PaulSnow/factom2d/testHelper"
)

/*
This test is the part B of a Network/Follower A/B pair of tests used to test
Just boots to test that follower can sync over a network
*/
func TestSyncB(t *testing.T) {

	peers := "127.0.0.1:38003"
	ResetSimHome(t)

	// write config file from identity9 to fnode0
	WriteConfigFile(9, 0, "", t)

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

	state0 := SetupSim("F", params, 10, 0, 0, t)
	WaitForBlock(state0, 9)
	Halt(t)
}
