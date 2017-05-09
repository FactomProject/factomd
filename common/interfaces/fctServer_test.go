package interfaces_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

func TestFCTServer(t *testing.T) {
	rbool := func() bool {
		if random.RandIntBetween(0, 2) == 1 {
			return true
		}
		return false
	}

	for i := 0; i < 5; i++ {
		c := primitives.RandomHash()
		n := random.RandomString()
		o := rbool()
		r := primitives.RandomHash()

		f := new(FctServer)
		f.ChainID = c
		f.Name = n
		f.Online = o
		f.Replace = r

		sf := new(FctServer)
		sf.ChainID = c
		sf.Name = n
		sf.Online = o
		sf.Replace = r
		if f.String() != sf.String() {
			t.Error("String should be same")
		}

		if !f.GetChainID().IsSameAs(c) {
			t.Error("Should be same chain")
		}

		if !(f.GetName() == n) {
			t.Error("Should be same name")
		}

		if f.IsOnline() != o {
			t.Error("Should be same online")
		}

		f.SetOnline(!o)
		if f.IsOnline() == o {
			t.Error("Should be different online")
		}

		if !f.LeaderToReplace().IsSameAs(r) {
			t.Error("Should be same replace")
		}

		f.SetReplace(primitives.RandomHash())
		if f.LeaderToReplace().IsSameAs(r) {
			t.Error("Should be different replace")
		}
	}
}
