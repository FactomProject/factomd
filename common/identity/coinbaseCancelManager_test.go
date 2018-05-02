package identity_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/identity"
)

func TestCancelGC(t *testing.T) {
	dbheight := uint32(10)

	c := NewCoinbaseCancelManager()
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
	c := NewCoinbaseCancelManager()
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
	c := NewCoinbaseCancelManager()
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
	c := NewCoinbaseCancelManager()
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

	c = NewCoinbaseCancelManager()
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
