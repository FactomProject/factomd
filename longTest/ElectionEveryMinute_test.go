package longtest

import (
	"fmt"
	"testing"

	. "github.com/PaulSnow/factom2d/engine"
	. "github.com/PaulSnow/factom2d/testHelper"
)

func TestElectionEveryMinute(t *testing.T) {
	//							  01234567890123456789012345678901
	state0 := SetupSim("LLLLLLLLLLLLLLLLLLLLLAAAAAAAAAAF", map[string]string{"--blktime": "60"}, 20, 10, 1, t)

	StatusEveryMinute(state0)
	s := GetFnodes()[1].State
	WaitMinutes(s, 1) // wait for start of next minute on fnode01
	// knock followers off one per minute
	start := s.CurrentMinute
	for i := 0; i < 10; i++ {
		s := GetFnodes()[i+1].State
		RunCmd(fmt.Sprintf("%d", i+1))
		WaitForMinute(s, (start+i+1)%10) // wait for selected minute
		RunCmd("x")
	}
	WaitMinutes(state0, 1)
	// bring them all back
	for i := 0; i < 10; i++ {
		RunCmd(fmt.Sprintf("%d", i+1))
		RunCmd("x")
	}

	WaitForAllNodes(state0) /// wait till everyone catches up
	ShutDownEverything(t)
}
