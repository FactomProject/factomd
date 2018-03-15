package testhelper_test

import (
	"fmt"
	"testing"

	"github.com/FactomProject/electiontesting/messages"
	"github.com/FactomProject/electiontesting/primitives"
	. "github.com/FactomProject/electiontesting/testhelper"
)

var _ = fmt.Println

func TestMajority(t *testing.T) {
	ah := NewAuthSetHelper(5, 5)
	if ah.Majority() != 3 {
		t.Errorf("majority should be 3, found %d", ah.Majority())
	}

	ah = NewAuthSetHelper(100, 5)
	if ah.Majority() != 51 {
		t.Errorf("majority should be 51, found %d", ah.Majority())
	}

	ah = NewAuthSetHelper(1, 5)
	if ah.Majority() != 1 {
		t.Error("majority should be 1, found %d", ah.Majority())
	}
}

func TestVoteFactory(t *testing.T) {
	ah := NewAuthSetHelper(5, 5)
	var loc primitives.ProcessListLocation
	eom := messages.NewEomMessage(ah.GetFeds()[0], loc)
	vol := messages.NewVolunteerMessage(eom, ah.GetAuds()[0])
	vf := ah.NewVoteFactory(vol)
	for i := 0; i < len(ah.GetFeds()); i++ {
		vote := vf.NextVote()
		if vote.Volunteer != vol {
			t.Error("Different volunteer that was passed")
		}
	}

	if ah.Majority() != vf.Majority() {
		t.Error("Majorties should be the same")
	}

	if len(vf.VotesListWithMajority()) != ah.Majority() {
		t.Errorf("Number of votes in majority is not correct. Exp %d, got %d", ah.Majority(), len(vf.VotesListWithMajority()))
	}

	if len(vf.MajorityDecisionListWithMajority()) != ah.Majority() {
		t.Errorf("Number of votes in majority is not correct. Exp %d, got %d", ah.Majority(), len(vf.MajorityDecisionListWithMajority()))
	}

	if len(vf.InsistenceListWithMajority()) != ah.Majority() {
		t.Errorf("Number of votes in majority is not correct. Exp %d, got %d", ah.Majority(), len(vf.InsistenceListWithMajority()))
	}

	fmt.Println(vf.NextPublish().MajorityIAckMessages)
	if len(vf.NextPublish().MajorityIAckMessages) != ah.Majority() {
		t.Errorf("Number of votes in majority is not correct. Exp %d, got %d", ah.Majority(), len(vf.NextPublish().MajorityIAckMessages))
	}
}
