// +build all 

package simtest

import (
	"testing"

	"github.com/FactomProject/factomd/engine"

	. "github.com/FactomProject/factomd/testHelper"
)

/*
This test is useful to exercise reboot behavior
here we copy a db and boot up an additional follower
*/
func TestAddFNode(t *testing.T) {
	ResetSimHome(t)          // clear out old test home
	for i := 0; i < 6; i++ { // build config files for the test
		WriteConfigFile(i, i, "", t) // just write the minimal config
	}
	state0 := SetupSim("LLLLLAA", map[string]string{"--db": "LDB"}, 25, 1, 1, t)
	WaitForBlock(state0, 7)
	CloneFnodeData(2, 7, t)
	AddFNode()
	state7 := engine.GetFnodes()[7].State // Get new node
	WaitForBlock(state7, 7)
	AssertAuthoritySet(t, "LLLLLAAF")
	ShutDownEverything(t)
}
