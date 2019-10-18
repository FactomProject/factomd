package nettest

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAnElection(t *testing.T) {
	n := SetupNode(DEV_NET, 0, t)

	target := 0

	// Find Leader to demote
	for i := 0; i < 5; i++ {
		if n.fnodes[i].NetworkInfo().Role == "Leader" {
			target = i
			break
		}
	}

	assert.Equal(t, "Leader", n.fnodes[target].NetworkInfo().Role) // assert we target a Leader

	// NOTE: this step can take awhile if devnet has been running for awhile
	n.WaitBlocks(2) // make sure local node is progressing

	n.fnodes[target].RunCmd("x") // x-out the targeted leader

	// REVIEW: perhaps we need to assert that node is in isolation mode?
	n.WaitBlocks(2) // wait for election & chain to progress

	n.fnodes[target].RunCmd("x") //bring him back
	n.WaitBlocks(2)              // wait of leader to update via dbstate and become an audit

	assert.Equal(t, "Audit", n.fnodes[target].NetworkInfo().Role)
}
