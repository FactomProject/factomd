package nettest

import (
	"testing"
)

func TestLoad(t *testing.T) {
	n := SetupNode(DEV_NET, 0, t)

	// NOTE: this step can take awhile if devnet has been running for awhile
	n.WaitBlocks(2) // make sure local node is progressing

	n.fnodes[2].RunCmd("R30") // start for 30 EPS on node2
	n.WaitBlocks(10)
	n.fnodes[2].RunCmd("R0") // stop load

	// TODO: improve to check health of network
}
