package election_test

import (
	"testing"

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
