// +build all 

package election_test

import (
	"testing"

	"fmt"

	. "github.com/FactomProject/factomd/electionsCore/election"
	"github.com/FactomProject/factomd/electionsCore/messages"
	"github.com/FactomProject/factomd/electionsCore/primitives"
	"github.com/FactomProject/factomd/electionsCore/testhelper"
)

// TestElectionConstructor is dumb. I just have it for imports and test copy setup
func TestElectionConstructor(t *testing.T) {
	a := testhelper.NewAuthSetHelper(5, 5)
	e := NewElection(a.GetFeds()[0], a.GetAuthSet())

	var _ = e
}

func TestElectionRanks(t *testing.T) {
	a := testhelper.NewAuthSetHelper(5, 5)
	var loc primitives.ProcessListLocation
	e := NewElection(a.GetFeds()[0], a.GetAuthSet())

	var _ = e
	messages.NewVolunteerMessage(messages.NewEomMessage(a.GetAuds()[0], loc), a.GetAuds()[0])
}

func TestElectionCopy(t *testing.T) {
	var loc primitives.ProcessListLocation
	au := testhelper.NewAuthSetHelper(5, 5)
	a := NewElection(au.GetFeds()[0], au.GetAuthSet())

	vol := messages.NewVolunteerMessage(messages.NewEomMessage(a.GetAuds()[0], loc), a.GetAuds()[0])
	l := messages.NewLeaderLevelMessage(primitives.NewIdentityFromInt(0), 0, 0, vol)

	a.MsgListIn = append(a.MsgListIn, NewDepthLeaderLevel(&l, 1))
	var _ = l
	b := a.Copy()

	ainp := fmt.Sprintf("%p", a.MsgListIn)
	aoutp := fmt.Sprintf("%p", a.MsgListOut)

	binp := fmt.Sprintf("%p", b.MsgListIn)
	boutp := fmt.Sprintf("%p", b.MsgListOut)

	if ainp == binp && len(a.MsgListIn) > 0 {
		t.Errorf("Same pointer for msgin")
	}

	if aoutp == boutp && len(a.MsgListOut) > 0 {
		t.Errorf("Same pointer for msg out")
	}
}

func TestLeaderLevelMsgSort(t *testing.T) {
	ls := make([]*messages.LeaderLevelMessage, 10)
	for i := range ls {
		ls[i] = new(messages.LeaderLevelMessage)
	}

	ls[2].Rank = 1
	ls[2].VolunteerPriority = 1

	ls[3].Rank = 2
	ls[3].VolunteerPriority = 1

	ls[5].Rank = 1
	ls[5].VolunteerPriority = 2

	ls[6].Rank = 7
	ls[6].VolunteerPriority = 2

	ls[7].Rank = 5
	ls[7].VolunteerPriority = 2

	ls[8].Rank = 0
	ls[8].VolunteerPriority = 2

	BubbleSortLeaderLevelMsg(ls)

	for i := range ls {
		fmt.Println(ls[i].Rank, ".", ls[i].VolunteerPriority)
	}
	fmt.Println()

	for i := range ls[:len(ls)-1] {
		j := i + 1

		if ls[j-1].Less(ls[j]) {
			t.Errorf("Leader sort bad")
		}
	}
}
