package simtest

import (
	"github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
	"time"
)

func LaggingAuditElection(t *testing.T, lag int, recovery int) {
	state0 := SetupSim("LAFL", map[string]string{ "--blktime": "30", "--falttimeout": "30"}, 15, 0, 0, t)
	state1 := engine.GetFnodes()[1].State

	WaitForBlock(state0, 6)
	WaitForAllNodes(state0)

	RunCmd("1")
	RunCmd("x") // take out audit

	time.Sleep(120*time.Second) // make audit lag behind

	RunCmd("0")
	RunCmd("x") // take out a leader

	time.Sleep(120*time.Second) // make audit lag behind

	RunCmd("1")
	RunCmd("x") // bring back audit

	WaitForAllNodes(state1)

	RunCmd("0")
	RunCmd("x") // bring back leader-should become Audit

	WaitBlocks(state1, recovery) // give time to come back

	AssertAuthoritySet(t, "ALFL")
	Halt(t)
}

// test electing an audit that is 1 block behind
func TestLaggingAuditElection1(t *testing.T) {
	LaggingAuditElection(t, 1, 2) // KLUDGE: fails when recovery = 2
}

// test electing an audit that is 2 blocks behind
func TestLaggingAuditElection2(t *testing.T) {
	LaggingAuditElection(t, 2, 2)
}
