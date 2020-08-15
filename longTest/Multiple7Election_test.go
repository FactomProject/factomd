package longtest

import (
	"fmt"
	"testing"

	. "github.com/PaulSnow/factom2d/testHelper"
)

func TestMultiple7Election(t *testing.T) {
	state0 := SetupSim("LLLLLLLLLFLLFLFLLLFLAAFAAAAFA", map[string]string{"--blktime": "60"}, 10, 7, 7, t)

	WaitForMinute(state0, 2)

	// Take 7 nodes off line
	for i := 1; i < 8; i++ {
		RunCmd(fmt.Sprintf("%d", i))
		RunCmd("x")
	}
	// force them all to be faulted
	WaitMinutes(state0, 1)

	// bring them back online
	for i := 1; i < 8; i++ {
		RunCmd(fmt.Sprintf("%d", i))
		RunCmd("x")
	}

	// Wait till they should have updated by DBSTATE
	WaitBlocks(state0, 2)
	WaitMinutes(state0, 1)
	WaitForAllNodes(state0)
	ShutDownEverything(t)
}
