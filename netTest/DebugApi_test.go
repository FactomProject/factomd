package nettest

import (
	"encoding/json"
	"testing"
)

func TestDebugApi(t *testing.T) {

	n := SetupNode(t)
	_ = n

	// KLUDGE: waiting on factomd_0
	// needs to be  more explicit
	n.WaitBlocks(1)
	peers := n.GetPeers()
	d, _ := json.Marshal(peers)
	t.Logf("%s", d)

}
