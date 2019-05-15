package longtest

import (
	"testing"
	"time"

	. "github.com/FactomProject/factomd/testHelper"
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
	state0 := StartSim(nodesLoadWith1pctDrop, params)

	// adjust simulation parameters
	RunCmd("s")
	RunCmd("Re")
	RunCmd("r")
	RunCmd("S10")
	RunCmd("F500")
	RunCmd("R5")

	time.Sleep(time.Second * 300) // wait 5 min
	t.Logf("LLHT: %v<=>%v", state0.GetDBHeightAtBoot(), state0.GetDBHeightComplete())
}
