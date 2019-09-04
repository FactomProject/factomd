package simtest

import (
	"github.com/FactomProject/factomd/engine"
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
)

func LaggingAuditElection(t *testing.T, lag int, recovery int) {
	state0 := SetupSim("LAFL", map[string]string{ "--blktime": "30"}, 15, 0, 0, t)
	state1 := engine.GetFnodes()[1].State

	WaitForBlock(state0, 6)
	WaitForAllNodes(state0)

	RunCmd("1")
	RunCmd("x") // take out audit

    WaitBlocks(state0, lag) // make audit lag behind

	RunCmd("0")
	RunCmd("x") // take out a leader

	RunCmd("1")
	RunCmd("x") // bring back audit

	WaitForAllNodes(state1)

	WaitBlocks(state1, recovery) // give time to come back

	RunCmd("0")
	RunCmd("x") // bring back leader-should become Audit

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
