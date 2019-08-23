package nettest

import (
	"testing"
)

func TestAnElection(t *testing.T) {
	// Assume network has booted up w/ FLALL roles
	n := SetupNode(DEV_NET, 0, t)

	target := 1

	n.StatusEveryMinute()
	n.WaitBlocks(2)

	n.fnodes[0].WaitMinutes(1)

	// remove the last leader
	n.fnodes[target].RunCmd("x")

	// wait for the election
	n.WaitMinutes(2)

	//bring him back
	n.fnodes[target].RunCmd("x")

	// wait for him to update via dbstate and become an audit
	n.WaitBlocks(2)
	n.WaitMinutes(1)

	//// PrintOneStatus(0, 0)
	//if GetFnodes()[2].State.Leader {
	//	t.Fatalf("Node 2 should not be a leader")
	//}
	//if !GetFnodes()[3].State.Leader && !GetFnodes()[4].State.Leader {
	//	t.Fatalf("Node 3 or 4  should be a leader")
	//}
}
