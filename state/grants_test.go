package state

import (
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
)

func makeExpected(grants []HardGrant) []interfaces.ITransAddress {
	var rval []interfaces.ITransAddress
	for _, g := range grants {
		rval = append(rval, factoid.NewOutAddress(g.Address, g.Amount))
	}
	return rval
}

func TestImmutableGrantTable(t *testing.T) {
	globals.Params.NetworkName = "LOCAL"
	constants.SetLocalCoinBaseConstants()

	grants1 := GetHardCodedGrants()
	grants2 := GetHardCodedGrants()

	if fmt.Sprintf("%p", grants1) == fmt.Sprintf("%p", grants2) {
		t.Errorf("Expected unique address for grant lists")
	}
	if fmt.Sprintf("%p", &grants1[0]) == fmt.Sprintf("%p", &grants2[0]) {
		t.Errorf("Expected unique address for grants")
	}
}

func TestGetGrantPayoutsFor(t *testing.T) {
	globals.Params.NetworkName = "LOCAL"
	constants.SetLocalCoinBaseConstants()

	grants := GetHardCodedGrants()

	// find all the heights we care about
	heights := map[uint32][]HardGrant{}
	min := uint32(9999999)
	max := uint32(0)
	for _, g := range grants {
		heights[g.DBh] = append(heights[g.DBh], g)
		if min > g.DBh {
			min = g.DBh
		}
		if max < g.DBh {
			max = g.DBh
		}
	}
	// loop thru the dbheights and make sure the payouts get returned
	for dbh := uint32(min - 1); dbh <= uint32(max+1); dbh++ {
		expected := makeExpected(heights[dbh])
		gotGrants := GetGrantPayoutsFor(dbh)
		if len(expected) != len(gotGrants) {
			t.Errorf("Expected %d grants but found %d", len(expected), len(gotGrants))
		}
		for i, p := range expected {
			if expected[i].GetAddress() == gotGrants[i].GetAddress() &&
				expected[i].GetAmount() == gotGrants[i].GetAmount() &&
				expected[i].GetUserAddress() == gotGrants[i].GetUserAddress() {
				t.Errorf("Expected: %v ", expected[i])
				t.Errorf("but found %v for grant #%d at %d", gotGrants[i], i, dbh)
			}
			fmt.Println(p.GetAmount(), p.GetUserAddress())
		}
	}

}
