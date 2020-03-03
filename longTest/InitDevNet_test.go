package longtest

import (
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"
)

// create Stub DBs & configs for DevNet Testing
func TestInitDevNet(t *testing.T) {
	home := simulation.ResetSimHome(t) // clear out old test home
	givenNodes := "FFFFL"
	state0 := simulation.SetupSim(givenNodes, map[string]string{"--blktime": "15", "--db": "LDB"}, 12, 0, 0, t)
	simulation.WaitForAllNodes(state0)

	addSpecialPeers := `LocalSpecialPeers                     = "factomd-0-0.factomd:8110 factomd-1-0.factomd:8110 factomd-2-0.factomd:8110 factomd-3-0.factomd:8110 factomd-4-0.factomd:8110"`
	// write identity keys out to config
	for i := 0; i < len(givenNodes); i++ { // build config files for the test
		if i == 0 {
			// use spare identity for fnode 0
			simulation.WriteConfigFile(len(givenNodes)+1, i, addSpecialPeers, t)
		} else {
			// use default identities for other nodes
			simulation.WriteConfigFile(i, i, addSpecialPeers, t)
		}

	}

	// wait one more block
	simulation.WaitBlocks(state0, 2)
	simulation.AssertAuthoritySet(t, givenNodes)
	simulation.ShutDownEverything(t)
	t.Logf("generated DB's & config here: %s/.factom", home)
}
