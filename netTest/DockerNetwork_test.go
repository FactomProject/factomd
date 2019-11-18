package nettest

import (
	"fmt"
	"testing"
)

// test using a network built w/ ./support/dev/docker-compose.json
func TestDockerNetwork(t *testing.T) {

	// address hardcoded to point a docker network
	n := SetupNode(DOCKER_NETWORK, 1, t)
	_ = n

	fmt.Printf("remote nodes: %v", n.fnodes)

	n.fnodes[0].RunCmd("R10")
	n.fnodes[0].WaitBlocks(1)
	n.fnodes[0].RunCmd("R0")

	//t.Logf("%v", n.fnodes)
}
