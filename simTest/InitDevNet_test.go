package simtest

import (
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
)

// create Stub DBs & configs for DevNet Testing
func TestInitDevNet(t *testing.T) {
	ResetSimHome(t) // clear out old test home
	givenNodes := "FAALL"
	state0 := SetupSim(givenNodes, map[string]string{"--blktime": "15", "--db": "LDB"}, 12, 0, 0, t)
	WaitForAllNodes(state0)
	AssertAuthoritySet(t, givenNodes)
	ShutDownEverything(t)

	// write identity keys out to config
	for i := 0; i < len(givenNodes); i++ { // build config files for the test
		if i == 0 {
			// use spare identity for fnode 0
			WriteConfigFile(len(givenNodes)+1, i, "", t)
		} else {
			// use default identities for other nodes
			WriteConfigFile(i, i, "", t)
		}

	}
}
