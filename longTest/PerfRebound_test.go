package longtest

import (
	"testing"
	"time"

	. "github.com/FactomProject/factomd/testHelper"
)

/*
send periodic load to sim network

NOTE: must run this test with a large timeout such as -timeout=9999h
*/
func TestPerfRebound(t *testing.T) {

	params := map[string]string{
		"--db": "LDB",
		//"--fastsaverate": "100",
		"--blktime":      "30",
		"--faulttimeout": "12",
		"--startdelay":   "0",
		"--debuglog":     ".",
	}
	state0 := SetupSim("LLLLLLFFFFFFFF", params, 60, 0, 0, t) // start 6L 8F

	// adjust simulation parameters
	RunCmd("s")  // show node state summary
	RunCmd("Re") // keep reloading EC wallet on 'tight' schedule (only small amounts)

	for x := 0; x < 5; x++ { // 5 iterations
		// 300s (5min) increments of load w/ +20s quiet at each end
		startHt := state0.GetDBHeightComplete()

		time.Sleep(time.Second * 20) // give some lead time
		RunCmd("R5")                 // Set Load 10 tx/sec

		time.Sleep(time.Second * 260) // Send Load

		RunCmd("R0")                 // Load 0 tx/sec
		time.Sleep(time.Second * 20) // quiet / rebound

		endHt := state0.GetDBHeightComplete()
		delta := endHt - startHt

		// show progress made during this run
		t.Logf("LLHT: %v<=>%v moved %v", startHt, endHt, delta)
		if delta < 9 { // 30 sec blocks - height should move at least 9 blocks each 5min period
			t.Fatalf("only moved %v blocks", delta)
			panic("FAILED")
		}
	}
}
