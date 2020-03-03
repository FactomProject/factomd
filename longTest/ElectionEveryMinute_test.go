package longtest

import (
	"fmt"
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"

	"github.com/FactomProject/factomd/fnode"
)

func TestElectionEveryMinute(t *testing.T) {
	//							  01234567890123456789012345678901
	state0 := simulation.SetupSim("LLLLLLLLLLLLLLLLLLLLLAAAAAAAAAAF", map[string]string{"--blktime": "60"}, 20, 10, 1, t)

	simulation.StatusEveryMinute(state0)
	s := fnode.Get(1).State
	simulation.WaitMinutes(s, 1) // wait for start of next minute on fnode01
	// knock followers off one per minute
	start := s.CurrentMinute
	for i := 0; i < 10; i++ {
		s := fnode.Get(i + 1).State
		simulation.RunCmd(fmt.Sprintf("%d", i+1))
		simulation.WaitForMinute(s, (start+i+1)%10) // wait for selected minute
		simulation.RunCmd("x")
	}
	simulation.WaitMinutes(state0, 1)
	// bring them all back
	for i := 0; i < 10; i++ {
		simulation.RunCmd(fmt.Sprintf("%d", i+1))
		simulation.RunCmd("x")
	}

	simulation.WaitForAllNodes(state0) /// wait till everyone catches up
	simulation.ShutDownEverything(t)
}
