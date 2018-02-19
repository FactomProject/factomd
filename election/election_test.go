package election_test

import (
	"testing"

	"fmt"

	. "github.com/FactomProject/electiontesting/election"
	"github.com/FactomProject/electiontesting/messages"
	"github.com/FactomProject/electiontesting/primitives"
	"github.com/FactomProject/electiontesting/testhelper"
)

// TestElectionConstructor is dumb. I just have it for imports and test copy setup
func TestElectionConstructor(t *testing.T) {
	a := testhelper.NewAuthSetHelper(5, 5)
	var loc primitives.ProcessListLocation
	e := NewElection(a.GetFeds()[0], a.GetAuthSet(), loc)

	var _ = e
}

func TestElectionRanks(t *testing.T) {
	a := testhelper.NewAuthSetHelper(5, 5)
	var loc primitives.ProcessListLocation
	e := NewElection(a.GetFeds()[0], a.GetAuthSet(), loc)

	var _ = e
	messages.NewVolunteerMessage(messages.NewEomMessage(a.GetAuds()[0], loc), a.GetAuds()[0])
}

func TestElectionCopy(t *testing.T) {
	var loc primitives.ProcessListLocation
	au := testhelper.NewAuthSetHelper(5, 5)
	a := NewElection(au.GetFeds()[0], au.GetAuthSet(), loc)

	vol := messages.NewVolunteerMessage(messages.NewEomMessage(a.GetAuds()[0], loc), a.GetAuds()[0])
	l := messages.NewLeaderLevelMessage(0, 0, 0, vol)
	a.MsgListIn = append(a.MsgListIn, &l)
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
