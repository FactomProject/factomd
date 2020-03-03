package simtest

import (
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"
)

func TestDemoteBootstrap(t *testing.T) {
	state0 := simulation.SetupSim("FLA", map[string]string{"--blktime": "15"}, 12, 0, 0, t)
	simulation.WaitForAllNodes(state0)
	simulation.AssertAuthoritySet(t, "FLA")
	simulation.ShutDownEverything(t)
}
