package simulation_test

import (
	"math"
	"testing"

	"github.com/FactomProject/factomd/fnode"
	. "github.com/FactomProject/factomd/simulation"
)

var fnodes []*fnode.FactomNode

func TestSimPeer(t *testing.T) {
	t.Skip("deprecated test")
	cnt := 40
	side := int(math.Sqrt(float64(cnt)))

	for i := 0; i < side; i++ {
		AddSimPeer(fnodes, i*side, (i+1)*side-1)
		AddSimPeer(fnodes, i, side*(side-1)+i)
		for j := 0; j < side; j++ {
			if j < side-1 {
				AddSimPeer(fnodes, i*side+j, i*side+j+1)
			}
			AddSimPeer(fnodes, i*side+j, ((i+1)*side)+j)
		}
	}

	if len(fnodes) != cnt {
		t.Errorf("Should have %d nodes found %v", cnt, len(fnodes))
	}
}
