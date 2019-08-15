package nettest

import (
	"fmt"
	"testing"
)

func TestDebugApi(t *testing.T) {

	// address hardcoded to point a docker network
	n := SetupNode("10.7.0.1:8110", t)
	_ = n

	// KLUDGE: waiting on factomd_0
	// needs to be  more explicit
	fmt.Printf("%v", n.fnodes)
	//assert.Equal(t,"http://10.7.0.1:8088/debug", n.fnodes[0].getAPIUrl())

	//n.fnodes[0].WaitBlocks(1)
	//t.Logf("%v", n.fnodes)
}
