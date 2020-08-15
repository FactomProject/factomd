package testhelper_test

import (
	"fmt"
	"testing"

	"github.com/PaulSnow/factom2d/electionsCore/messages"
	"github.com/PaulSnow/factom2d/electionsCore/primitives"
	. "github.com/PaulSnow/factom2d/electionsCore/testhelper"
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
		t.Errorf("majority should be 1, found %d", ah.Majority())
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

	// This code was removed
}
