package identity_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/identityEntries"
)

func TestCancelTally(t *testing.T) {
	im := RandomIdentityManagerWithCounts(5, 5)
	c := NewCoinbaseCancelManager(im)

	feds := make([]*Authority, 0)
	for _, a := range im.Authorities {
		if a.Status == constants.IDENTITY_FEDERATED_SERVER {
			feds = append(feds, a)
		}
	}

	// Has options
	addAndTestWithOptions := func(auths []int, h uint32, i []uint32, exp []int, t *testing.T, duplicates bool, gc bool) {
		var outs []uint32
		for ci, count := range auths {
			index := i[ci]
			for ai := 0; ai < count; ai++ {
				c.AddCancel(newCoinbaseCancel(feds[ai], h, index))
				if duplicates {
					c.AddCancel(newCoinbaseCancel(feds[ai], h, index))
				}
			}
		}

		outs = c.CanceledOutputs(h)
		if len(outs) != len(exp) {
			t.Errorf("Exp %d cancelled, got %d", len(exp), len(outs))
			return
		}

		for i := range outs {
			if outs[i] != uint32(exp[i]) {
				t.Errorf("Exp %d at %d, found %d", exp[i], i, outs[i])
			}
		}

		// Reset at the end
		if gc {
			c.GC(h + constants.COINBASE_DECLARATION*2)
		}
	}

	al := func(list ...int) []int {
		return list
	}

	addAndTest := func(auths []int, h uint32, i []uint32, exp []int, t *testing.T) {
		addAndTestWithOptions(auths, h, i, exp, t, false, true)
	}

	// Test GC
	addAndTest(al(5), 1, []uint32{1}, []int{1}, t)
	addAndTest(al(1), 1, []uint32{1}, []int{}, t)

	// Test 0 case
	addAndTest(al(), 1, []uint32{}, []int{}, t)
	addAndTest(al(), 2, []uint32{}, []int{}, t)

	// Test no majority
	addAndTest(al(1), 1, []uint32{1}, []int{}, t)
	addAndTest(al(2), 1, []uint32{1}, []int{}, t)
	addAndTest(al(1, 1), 1, []uint32{1, 2}, []int{}, t)
	addAndTest(al(1, 1, 2, 1, 2, 1), 1, []uint32{1, 2, 3, 4, 5, 6}, []int{}, t)

	// Test majority
	addAndTest(al(3), 1, []uint32{1}, []int{1}, t)
	addAndTest(al(4), 1, []uint32{1}, []int{1}, t)
	addAndTest(al(5), 1, []uint32{1}, []int{1}, t)
	addAndTest(al(1, 2, 3, 4, 5), 1, []uint32{1, 2, 3, 4, 5}, []int{3, 4, 5}, t)

	// Test duplicates on 1 minute
	addAndTestWithOptions(al(0, 3, 3, 1, 5), 1, []uint32{1, 2, 2, 4, 5}, []int{2, 5}, t, true, false)

	// Test output order
	addAndTest(al(3, 3, 3, 3, 3, 3), 1, []uint32{6, 5, 4, 3, 2, 1}, []int{1, 2, 3, 4, 5, 6}, t)
	addAndTest(al(3, 3, 3, 3, 3, 3), 1, []uint32{1, 3, 5, 2, 4, 6}, []int{1, 2, 3, 4, 5, 6}, t)
	addAndTest(al(3, 3, 3, 3, 3, 3), 1, []uint32{2, 4, 6, 1, 3, 5}, []int{1, 2, 3, 4, 5, 6}, t)

	// Test edges
	addAndTest(al(3), 1, []uint32{0}, []int{0}, t)
	addAndTest(al(3), 0, []uint32{65}, []int{65}, t)

	// Test panic cases
	addAndTestWithOptions(al(0), 0, []uint32{1}, []int{}, t, false, true)
	if c.IsCoinbaseCancelled(10, 1) ||
		c.IsCoinbaseCancelled(1000, 0) ||
		c.IsCoinbaseCancelled(0, 100) {
		t.Errorf("Should not be cancelled")
	}

	addAndTestWithOptions(al(0), 1000, []uint32{1}, []int{}, t, false, true)
	if c.IsCoinbaseCancelled(0, 0) ||
		c.IsCoinbaseCancelled(9, 9) ||
		c.IsCoinbaseCancelled(100, 100) {
		t.Errorf("Should not be cancelled")
	}

}

func newCoinbaseCancel(id *Authority, h, i uint32) identityEntries.NewCoinbaseCancelStruct {
	cc := new(identityEntries.NewCoinbaseCancelStruct)
	cc.RootIdentityChainID = id.AuthorityChainID
	cc.CoinbaseDescriptorHeight = h
	cc.CoinbaseDescriptorIndex = i
	return *cc
}

func TestCancelGC(t *testing.T) {
	dbheight := uint32(10)

	c := NewCoinbaseCancelManager(nil)
	c.MarkAdminBlockRecorded(dbheight, 1)
	if !c.IsAdminBlockRecorded(dbheight, 1) {
		t.Errorf("Should be true, found false")
	}

	c.GC(dbheight + constants.COINBASE_DECLARATION + 1)
	if c.IsAdminBlockRecorded(dbheight, 1) {
		t.Errorf("Should be false, found true")
	}

	if _, ok := c.AdminBlockRecord[10]; ok {
		t.Errorf("Should be deleted")
	}

	if _, ok := c.Proposals[10]; ok {
		t.Errorf("Should be deleted")
	}
}

func TestAdminBlockRecorded(t *testing.T) {
	c := NewCoinbaseCancelManager(nil)
	if c.IsAdminBlockRecorded(0, 1) {
		t.Errorf("Should be false, found true")
	}

	c.MarkAdminBlockRecorded(10, 1)
	if !c.IsAdminBlockRecorded(10, 1) {
		t.Errorf("Should be true, found false")
	}

	if c.IsAdminBlockRecorded(10, 2) {
		t.Errorf("Should be false, found true")
	}

}

func TestAddProposalRadom(t *testing.T) {
	c := NewCoinbaseCancelManager(nil)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 1000; i++ {
		c.AddNewProposalHeight(rand.Uint32())
		err := checkList(c.ProposalsList, i+1)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestAddProposalVector(t *testing.T) {
	c := NewCoinbaseCancelManager(nil)
	// Check 0
	err := checkList(c.ProposalsList, 0)
	if err != nil {
		t.Error(err)
	}

	// Check prepend on list size 1
	c.AddNewProposalHeight(100)
	err = checkList(c.ProposalsList, 1)
	if err != nil {
		t.Error(err)
	}

	c.AddNewProposalHeight(2)
	err = checkList(c.ProposalsList, 2)
	if err != nil {
		t.Error(err)
	}

	c = NewCoinbaseCancelManager(nil)
	// Check append on list size 1
	c.AddNewProposalHeight(2)
	err = checkList(c.ProposalsList, 1)
	if err != nil {
		t.Error(err)
	}

	c.AddNewProposalHeight(100)
	err = checkList(c.ProposalsList, 2)
	if err != nil {
		t.Error(err)
	}

	// Check same value
	c.AddNewProposalHeight(100)
	err = checkList(c.ProposalsList, 3)
	if err != nil {
		t.Error(err)
	}

}

func checkList(list []uint32, l int) error {
	err := checkListLength(list, l)
	if err != nil {
		return err
	}
	return checkListSorted(list)
}

func checkListLength(list []uint32, l int) error {
	if len(list) != l {
		return fmt.Errorf("Expect length of %d,found %d", l, len(list))
	}
	return nil
}

func checkListSorted(list []uint32) error {
	last := 0
	for v := range list {
		if last > v {
			return fmt.Errorf("%d found before %d. Not in order", last, v)
		}
	}
	return nil
}
