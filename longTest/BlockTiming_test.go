package longtest

import (
	"fmt"
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"
	"time"
)

/*
send consistent load to simulator ramping up over 5 iterations.

NOTE: must run this test with a large timeout such as -timeout=9999h
*/
func TestBlockTiming(t *testing.T) {
	simulation.ResetSimHome(t) // ditch the old data

	params := map[string]string{
		"--blktime":      "30",
		"--faulttimeout": "12",
		"--startdelay":   "0",
		//"--db":           "LDB", // XXX using the db incurs heavy IO
		//"--debuglog":     ".", // enable logs cause max ~ 50 TPS
	}
	state0 := simulation.SetupSim("LLLFF", params, 60, 0, 0, t) // start 6L 8F

	// adjust simulation parameters
	simulation.RunCmd("s")  // show node state summary
	simulation.RunCmd("Re") // keep reloading EC wallet on 'tight' schedule (only small amounts)

	incrementLoad := 10 // tx
	setLoad := 10       // tx/sec

	for x := 0; x < 5; x++ {
		simulation.RunCmd(fmt.Sprintf("R%v", setLoad)) // Load tx/sec
		startHt := state0.GetDBHeightComplete()
		time.Sleep(time.Second * 300) // test 300s (5min) increments

		endHt := state0.GetDBHeightComplete()
		delta := endHt - startHt

		// ramp up load
		setLoad = setLoad + incrementLoad

		// show progress made during this run
		t.Logf("LLHT: %v<=>%v moved %v", startHt, endHt, delta)
		if delta < 9 { // 30 sec blocks - height should move at least 9 blocks each 5min period
			t.Fatalf("only moved %v blocks", delta)
		}
	}
}
