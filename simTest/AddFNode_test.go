package simtest

import (
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"

	"github.com/FactomProject/factomd/fnode"
)

/*
This test is useful to exercise reboot behavior
here we copy a db and boot up an additional follower
*/
func TestAddFNode(t *testing.T) {
	simulation.ResetSimHome(t) // clear out old test home
	for i := 0; i < 6; i++ {   // build config files for the test
		simulation.WriteConfigFile(i, i, "", t) // just write the minimal config
	}
	state0 := simulation.SetupSim("LLLLLAA", map[string]string{"--db": "LDB"}, 25, 1, 1, t)
	simulation.WaitForBlock(state0, 7)
	simulation.CloneFnodeData(2, 7, t)
	simulation.AddFNode()
	state7 := fnode.Get(7).State // Get new node
	simulation.WaitForBlock(state7, 7)
	simulation.AssertAuthoritySet(t, "LLLLLAAF")
	simulation.ShutDownEverything(t)
}
