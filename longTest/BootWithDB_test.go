package longtest

import (
	"testing"
	"time"

	. "github.com/FactomProject/factomd/testHelper"
)

var givenNodes string = "LLLLLLLLFFFFFF"

// Run this first to configure network and initialize databases
func TestSetupBootWithoutDB(t *testing.T) {
	// FIXME: later versions of factomd test suite will create a home directory based on test name
	// this step should be relocated into TestBootWithDB function
	params := map[string]string{
		"--db":  "LDB",
		"--net": "alot+",
	}
	state0 := SetupSim(givenNodes, params, 10, 0, 0, t)
	WaitBlocks(state0, 1)
}

/*
Subsequent runs after network is setup, can be re-run to check behavior when booting w/ existing DB's

Replicates behavior of
factomd  --network=LOCAL --fastsaverate=100 --checkheads=false --count=15 --net=alot+ --blktime=600 --faulttimeout=12 --enablenet=false --startdelay=2 $@ > out.txt 2> err.txt
*/
func TestBootWithDB(t *testing.T) {
	params := map[string]string{
		"--db":           "LDB",
		"--fastsaverate": "100",
		"--net":          "alot+",
		"--blktime":      "30",
		"--faulttimeout": "12",
		"--startdelay":   "2",
	}
	state0 := StartSim(givenNodes, params)

	// adjust simulation parameters
	RunCmd("s")
	RunCmd("Re")
	RunCmd("r")
	RunCmd("S10")
	RunCmd("F500")

	// REVIEW it's possible changing timing after boot can induce issues
	//RunCmd("T600") // already set

	RunCmd("R5")

	time.Sleep(time.Second * 300)
	t.Logf("LLHT: %v<=>%v", state0.GetDBHeightAtBoot(), state0.GetDBHeightComplete())
}
