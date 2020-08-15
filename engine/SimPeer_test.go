package engine_test

import (
	"math"
	"testing"

	. "github.com/PaulSnow/factom2d/engine"
)

var fnodes []*FactomNode

func SimPeerTest(t *testing.T) {
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
		t.Errorf("Should have %d nodes", cnt)
	}
}
