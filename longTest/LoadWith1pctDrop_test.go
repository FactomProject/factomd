package longtest

import (
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// authority node configuration
var nodesLoadWith1pctDrop string = "LLLLLLLLFFFFFF"

/*
1st Part - Deletes old test data and re-initializes a new network
*/
func TestSetupLoadWith1pctDrop(t *testing.T) {
	homeDir := simulation.GetLongTestHome(t)
	simulation.ResetTestHome(homeDir, t)

	params := map[string]string{
		"--db":         "LDB",
		"--net":        "alot+",
		"--factomhome": homeDir,
	}
	state0 := simulation.SetupSim(nodesLoadWith1pctDrop, params, 10, 0, 0, t)
	simulation.WaitBlocks(state0, 1)
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
		"--factomhome":   simulation.GetLongTestHome(t),
	}
	state0 := simulation.StartSim(len(nodesLoadWith1pctDrop), params)

	// adjust simulation parameters
	simulation.RunCmd("s")    // show node state summary
	simulation.RunCmd("Re")   // keep reloading EC wallet on 'tight' schedule (only small amounts)
	simulation.RunCmd("r")    // reset all nodes in the simulation (maybe not needed)
	simulation.RunCmd("S10")  // message drop rate 1%
	simulation.RunCmd("F500") // add 500 ms delay to all messages
	simulation.RunCmd("R5")   // Load 5 tx/sec

	time.Sleep(time.Second * 300) // wait 5 min
	startHt := state0.GetDBHeightAtBoot()
	endHt := state0.GetDBHeightComplete()
	t.Logf("LLHT: %v<=>%v", startHt, endHt)

	// normally without load we expect to create 10 blocks over the span of 5 min
	assert.True(t, endHt-startHt >= 5) // check that we created at least 1 block per min
}
