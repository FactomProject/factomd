package longtest

import (
	"testing"

	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
)

func TestTripleElections(t *testing.T) {
	state.MMR_enable = false // No MMR for you!

	RanSimTest = true
	//                            0123456789AB
	state0 := SetupSim("LALALALLLLFF", map[string]string{"--debuglog": ".", "--blktime": "20"}, 360, 30, 30, t)

	for minute := 0; minute < 10; minute += 2 {
		WaitForMinute(state0, minute)
		RunCmd("2")            // select 1
		RunCmd("x")            // off the net
		RunCmd("4")            // select 2
		RunCmd("x")            // off the net
		RunCmd("6")            // select 3
		RunCmd("x")            // off the net
		WaitMinutes(state0, 2) // wait for elections
		RunCmd("2")            // select 1
		RunCmd("x")            // on the net
		RunCmd("4")            // select 2
		RunCmd("x")            // on the net
		RunCmd("6")            // select 3
		RunCmd("x")            // on the net
		WaitBlocks(state0, 2)  // wait till nodes should have updated by dbstate

		WaitForMinute(state0, minute+1)
		RunCmd("1")            // select 1
		RunCmd("x")            // off the net
		RunCmd("3")            // select 2
		RunCmd("x")            // off the net
		RunCmd("5")            // select 3
		RunCmd("x")            // off the net
		WaitMinutes(state0, 2) // wait for elections
		RunCmd("1")            // select 1
		RunCmd("x")            // on the net
		RunCmd("3")            // select 2
		RunCmd("x")            // on the net
		RunCmd("5")            // select 3
		RunCmd("x")            // on the net
		WaitBlocks(state0, 2)  // wait till nodes should have updated by dbstate

	}
	WaitForAllNodes(state0)
	ShutDownEverything(t)
} // TestTripleElections(){...}
