package simtest

import (
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
)

func LaggingAuditElection(t *testing.T, lag int, recovery int) {
	state0 := SetupSim("LLLALFL", map[string]string{}, 15, 0, 0, t)
	WaitForBlock(state0, 6)
	WaitForAllNodes(state0)

	RunCmd("3")
	RunCmd("x")
	WaitBlocks(state0, lag) // make audit lag behind

	RunCmd("4")
	RunCmd("x") // take out a leader

	// REVIEW: should the be relocated?
	RunCmd("3")
	RunCmd("x") // bring back audit

	WaitMinutes(state0, 2) // wait for audit to be elected

	RunCmd("4")
	RunCmd("x") // bring back leader-should become Audit

	// Do we need to wait by timeclock? rather than waitblocks
	WaitBlocks(state0, recovery) // give time to come back

	//WaitForAllNodes(state0) // REVIEW: is this desired? rather than using WaitBlocks as above?

	AssertAuthoritySet(t, "LLLLAFL") // leader
	Halt(t)
}

// test electing an audit that is 1 block behind
func TestLaggingAuditElection1(t *testing.T) {
	LaggingAuditElection(t, 1, 3) // KLUDGE: fails when recovery = 2
}

// test electing an audit that is 2 blocks behind
func TestLaggingAuditElection2(t *testing.T) {
	LaggingAuditElection(t, 2, 2)
}
