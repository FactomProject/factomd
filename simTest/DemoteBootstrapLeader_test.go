package simtest

import (
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
)

func TestDemoteBootstrap(t *testing.T) {
	state0 := SetupSim("FLA", map[string]string{"--blktime": "15"}, 12, 0, 0, t)
	WaitForAllNodes(state0)
	AssertAuthoritySet(t, "FLA")
	ShutDownEverything(t)
}
