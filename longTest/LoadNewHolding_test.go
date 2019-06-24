package longtest

import (
	"github.com/FactomProject/factomd/engine"
	"testing"
	"time"

	. "github.com/FactomProject/factomd/testHelper"
)

// authority node configuration
var nodesLoadNewHolding string = "LLLLLLLLFFFFFF"

/*
1st Part - Deletes old test data and re-initializes a new network
*/
func TestSetupLoadNewHolding(t *testing.T) {
	homeDir := GetLongTestHome(t)
	ResetTestHome(homeDir, t)

	params := map[string]string{
		"--db":         "LDB",
		"--net":        "alot+",
		"--factomhome": homeDir,
	}
	state0 := SetupSim(nodesLoadNewHolding, params, 10, 0, 0, t)
	WaitBlocks(state0, 1)
}

/*
2nd Part Subsequent runs after network is setup

can be re-run to check behavior when booting w/ existing DB's
*/
func TestLoadNewHolding(t *testing.T) {
	params := map[string]string{
		"--db": "LDB",
		//"--fastsaverate": "100",
		"--net":          "alot+",
		"--blktime":      "30",
		"--faulttimeout": "12",
		"--startdelay":   "0",
		"--debuglog":     ".",
		"--factomhome":   GetLongTestHome(t),
	}
	state0 := StartSim(nodesLoadNewHolding, params)

	WaitForAllNodes(state0)

	// adjust simulation parameters
	RunCmd("s")  // show node state summary
	RunCmd("Re") // keep reloading EC wallet on 'tight' schedule (only small amounts)
	//RunCmd("r")  // rotate wsapi
	//RunCmd("S10")  // message drop rate 1%
	//RunCmd("F500") // add 500 ms delay to all messages

	// REVIEW: results in a stall if load starts before network is up

	LogStuck := func(comment string) {
		for _, fnode := range engine.GetFnodes() {
			s := fnode.State
			for _, h := range s.Hold.Messages() {
				for _, m := range h {
					s.LogMessage("newholding", comment, m)
				}
			}
		}
	}

	for x := 0; x < 5; x++ { // 5 iterations
		startHt := state0.GetDBHeightComplete()
		time.Sleep(time.Second * 20)  // wait network to be up
		RunCmd("R5")                  // Load 5 tx/sec
		time.Sleep(time.Second * 260) // wait for rebound

		LogStuck("held_during_load")

		RunCmd("R0")                 // Load 0 tx/sec
		time.Sleep(time.Second * 20) // wait for rebound

		LogStuck("stuck_after_load")

		endHt := state0.GetDBHeightComplete()

		delta := endHt - startHt
		// show progress made during this run
		t.Logf("LLHT: %v<=>%v moved %v", startHt, endHt, delta)
		if delta < 9 {
			t.Fatalf("only moved %v blocks", delta)
			panic("FAILED")
		}
	}

}
