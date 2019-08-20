package nettest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDockerNetwork(t *testing.T) {

	// address hardcoded to point a docker network
	n := SetupNode("10.7.0.1:8110", 1, t)
	_ = n

	// KLUDGE: waiting on factomd_0
	// needs to be  more explicit
	fmt.Printf("remote nodes: %v", n.fnodes)

	/*
	n.fnodes[0].RunCmd("R10")
	n.fnodes[0].WaitBlocks(1)
	n.fnodes[0].RunCmd("R0")
	*/
	//t.Logf("%v", n.fnodes)
}

func TestNetworkInfo(t *testing.T) {
	n := SetupNode("127.0.0.1:39001", 0, t)
	r := n.fnodes[0].NetworkInfo()
	assert.Equal(t, "Follower", r.Role)
}
