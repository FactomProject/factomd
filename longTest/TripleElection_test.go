package longtest

import (
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"

	"github.com/FactomProject/factomd/state"
)

func TestTripleElections(t *testing.T) {
	state.MMR_enable = false // No MMR for you!

	simulation.RanSimTest = true
	//                            0123456789AB
	state0 := simulation.SetupSim("LALALALLLLFF", map[string]string{"--debuglog": ".", "--blktime": "20"}, 360, 30, 30, t)

	for minute := 0; minute < 10; minute += 2 {
		simulation.WaitForMinute(state0, minute)
		simulation.RunCmd("2")            // select 1
		simulation.RunCmd("x")            // off the net
		simulation.RunCmd("4")            // select 2
		simulation.RunCmd("x")            // off the net
		simulation.RunCmd("6")            // select 3
		simulation.RunCmd("x")            // off the net
		simulation.WaitMinutes(state0, 2) // wait for elections
		simulation.RunCmd("2")            // select 1
		simulation.RunCmd("x")            // on the net
		simulation.RunCmd("4")            // select 2
		simulation.RunCmd("x")            // on the net
		simulation.RunCmd("6")            // select 3
		simulation.RunCmd("x")            // on the net
		simulation.WaitBlocks(state0, 2)  // wait till nodes should have updated by dbstate

		simulation.WaitForMinute(state0, minute+1)
		simulation.RunCmd("1")            // select 1
		simulation.RunCmd("x")            // off the net
		simulation.RunCmd("3")            // select 2
		simulation.RunCmd("x")            // off the net
		simulation.RunCmd("5")            // select 3
		simulation.RunCmd("x")            // off the net
		simulation.WaitMinutes(state0, 2) // wait for elections
		simulation.RunCmd("1")            // select 1
		simulation.RunCmd("x")            // on the net
		simulation.RunCmd("3")            // select 2
		simulation.RunCmd("x")            // on the net
		simulation.RunCmd("5")            // select 3
		simulation.RunCmd("x")            // on the net
		simulation.WaitBlocks(state0, 2)  // wait till nodes should have updated by dbstate

	}
	simulation.WaitForAllNodes(state0)
	simulation.ShutDownEverything(t)
} // TestTripleElections(){...}
