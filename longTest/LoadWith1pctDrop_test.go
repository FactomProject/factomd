package longtest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/PaulSnow/factom2d/testHelper"
)

// authority node configuration
var nodesLoadWith1pctDrop string = "LLLLLLLLFFFFFF"

/*
1st Part - Deletes old test data and re-initializes a new network
*/
func TestSetupLoadWith1pctDrop(t *testing.T) {
	homeDir := GetLongTestHome(t)
	ResetTestHome(homeDir, t)

	params := map[string]string{
		"--db":         "LDB",
		"--net":        "alot+",
		"--factomhome": homeDir,
	}
	state0 := SetupSim(nodesLoadWith1pctDrop, params, 10, 0, 0, t)
	WaitBlocks(state0, 1)
}

/*
2nd Part Subsequent runs after network is setup

can be re-run to check behavior when booting w/ existing DB's

Replicates behavior of
factomd  --network=LOCAL --fastsaverate=100 --checkheads=false --count=15 --net=alot+ --blktime=600 --faulttimeout=12 --enablenet=false --startdelay=2 $@ > out.txt 2> err.txt
*/
func TestLoadWith1pctDrop(t *testing.T) {
	params := map[string]string{
		"--db":           "LDB",
		"--fastsaverate": "100",
		"--net":          "alot+",
		"--blktime":      "30",
		"--faulttimeout": "12",
		"--startdelay":   "2",
		"--factomhome":   GetLongTestHome(t),
	}
	state0 := StartSim(len(nodesLoadWith1pctDrop), params)

	// adjust simulation parameters
	RunCmd("s")    // show node state summary
	RunCmd("Re")   // keep reloading EC wallet on 'tight' schedule (only small amounts)
	RunCmd("r")    // reset all nodes in the simulation (maybe not needed)
	RunCmd("S10")  // message drop rate 1%
	RunCmd("F500") // add 500 ms delay to all messages
	RunCmd("R5")   // Load 5 tx/sec

	time.Sleep(time.Second * 300) // wait 5 min
	startHt := state0.GetDBHeightAtBoot()
	endHt := state0.GetDBHeightComplete()
	t.Logf("LLHT: %v<=>%v", startHt, endHt)

	// normally without load we expect to create 10 blocks over the span of 5 min
	assert.True(t, endHt-startHt >= 5) // check that we created at least 1 block per min
}
