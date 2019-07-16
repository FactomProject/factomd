package engine_test

import (
	"fmt"
	"testing"
	"time"

	. "github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
)

func TestSaveState1(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	// remove all the old database files
	SystemCall("find  test/.factom/m2 -name LOCAL | xargs rm -rvf ")

	state0 := SetupSim("LAFL", map[string]string{"--fastsaverate": "4", "--db": "LDB", "--factomhome": "test"}, 12, 0, 0, t)
	StatusEveryMinute(state0)
	WaitMinutes(state0, 2)
	RunCmd("R5")
	WaitForBlock(state0, 11)
	WaitMinutes(state0, 1)
	WaitForAllNodes(state0)
	PrintOneStatus(0, 0)
	ShutDownEverything(t)

	for _, x := range GetFnodes() {
		newState := x.State

		fmt.Println("FactomNodeName: ", newState.FactomNodeName)
		fmt.Println("	IdentityChainID: ", newState.IdentityChainID)
		fmt.Println("	ServerPrivKey: ", newState.LocalServerPrivKey)
		fmt.Println("	ServerPublicKey: ", newState.ServerPubKey)
	}
	time.Sleep(10 * time.Second)
}

func TestCreateDB_LLLLLLAAAAAFFFF(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	// remove all the old database files
	SystemCall("find  test/.factom/m2 -name LOCAL | xargs rm -rvf ")
	state0 := SetupSim("LLLLLLAAAAAFFFF", map[string]string{"--db": "LDB", "--factomhome": "test", "--network": "CUSTOM", "--customnet": "devnet"}, 6, 0, 0, t)
	WaitForAllNodes(state0)
	PrintOneStatus(0, 0)
	ShutDownEverything(t)

	for _, x := range GetFnodes() {
		newState := x.State

		fmt.Println("FactomNodeName: ", newState.FactomNodeName)
		fmt.Println("	IdentityChainID: ", newState.IdentityChainID)
		fmt.Println("	ServerPrivKey: ", newState.LocalServerPrivKey)
		fmt.Println("	ServerPublicKey: ", newState.ServerPubKey)
	}
	time.Sleep(10 * time.Second)
}

func TestSaveState2(t *testing.T) {
	if RanSimTest {
		return
	}
	RanSimTest = true

	// remove fnode02's fastboot and fnode01's whole database
	SystemCall("rm -vfr test/.factom/m2/local-database/ldb/Sim02/LOCAL/  test/.factom/m2/local-database/ldb/Sim01/FastBoot_LOCAL_v10.db")
	state0 := SetupSim("FFFF", map[string]string{"--fastsaverate": "4", "--db": "LDB", "--factomhome": "test", "--blktime": "20"}, 20, 0, 0, t)

	// check we booted from database to the right state
	Audits = 1
	Leaders = 2
	Followers = 1
	CheckAuthoritySet(t)

	StatusEveryMinute(state0)
	WaitForBlock(state0, 10)
	WaitForAllNodes(state0)
	PrintOneStatus(0, 0)

	CheckAuthoritySet(t)
	ShutDownEverything(t)
}
