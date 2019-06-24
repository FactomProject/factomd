package longtest

import (
	"github.com/FactomProject/factomd/engine"
	"testing"

	. "github.com/FactomProject/factomd/testHelper"
)

/*
 */
func TestNewHoldingStall(t *testing.T) {

	params := map[string]string{
		"--db": "LDB",
		//"--fastsaverate": "100",
		"--blktime":      "30",
		"--faulttimeout": "12",
		"--startdelay":   "0",
		"--debuglog":     ".",
	}
	state0 := SetupSim("LLLLLLFFFFFFFF", params, 17, 0, 0, t) // start 6L 8F

	// adjust simulation parameters
	RunCmd("s")  // show node state summary
	RunCmd("Re") // keep reloading EC wallet on 'tight' schedule (only small amounts)

	RunCmd("R5") // Set Load 10 tx/sec

	state3 := engine.GetFnodes()[3].State

	WaitForBlock(state3, 16) // node 3 stalls at block 15

	// FIXME: test should timeout
	_ = state0
}
