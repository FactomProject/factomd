package simtest

import (
	"testing"

	. "github.com/PaulSnow/factom2d/testHelper"
)

// create Stub DBs & configs for DevNet Testing
func TestInitDevNet(t *testing.T) {
	home := ResetSimHome(t) // clear out old test home
	givenNodes := "FAALL"
	state0 := SetupSim(givenNodes, map[string]string{"--blktime": "15", "--db": "LDB"}, 12, 0, 0, t)
	WaitForAllNodes(state0)

	addSpecialPeers := `LocalSpecialPeers                     = "factomd-0-0.factomd:8110 factomd-1-0.factomd:8110 factomd-2-0.factomd:8110 factomd-3-0.factomd:8110 factomd-4-0.factomd:8110"`
	// write identity keys out to config
	for i := 0; i < len(givenNodes); i++ { // build config files for the test
		if i == 0 {
			// use spare identity for fnode 0
			WriteConfigFile(len(givenNodes)+1, i, addSpecialPeers, t)
		} else {
			// use default identities for other nodes
			WriteConfigFile(i, i, addSpecialPeers, t)
		}

	}

	// KLUDGE make fnode0 an audit
	RunCmd("0")
	RunCmd("o")
	WaitBlocks(state0, 1)

	AssertAuthoritySet(t, givenNodes)
	ShutDownEverything(t)
	t.Logf("generated DB's & config here: %s/.factom", home)
}
