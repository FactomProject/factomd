package longtest

import (
	"fmt"
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"
)

func TestMultiple7Election(t *testing.T) {
	state0 := simulation.SetupSim("LLLLLLLLLFLLFLFLLLFLAAFAAAAFA", map[string]string{"--blktime": "60"}, 10, 7, 7, t)

	simulation.WaitForMinute(state0, 2)

	// Take 7 nodes off line
	for i := 1; i < 8; i++ {
		simulation.RunCmd(fmt.Sprintf("%d", i))
		simulation.RunCmd("x")
	}
	// force them all to be faulted
	simulation.WaitMinutes(state0, 1)

	// bring them back online
	for i := 1; i < 8; i++ {
		simulation.RunCmd(fmt.Sprintf("%d", i))
		simulation.RunCmd("x")
	}

	// Wait till they should have updated by DBSTATE
	simulation.WaitBlocks(state0, 2)
	simulation.WaitMinutes(state0, 1)
	simulation.WaitForAllNodes(state0)
	simulation.ShutDownEverything(t)
}
