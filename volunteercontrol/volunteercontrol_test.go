package volunteercontrol_test

import (
	"fmt"
	"github.com/FactomProject/electiontesting/messages"
	"github.com/FactomProject/electiontesting/primitives"
	"github.com/FactomProject/electiontesting/testhelper"
	. "github.com/FactomProject/electiontesting/volunteercontrol"
	"testing"
)

var _ = fmt.Println

func TestSimpleVolunteerControl(t *testing.T) {
	as := testhelper.NewAuthSetHelper(3, 3)
	me := as.NextIdentity()

	var loc primitives.ProcessListLocation
	eom := messages.NewEomMessage(as.GetAuds()[0], loc)
	vol := messages.NewVolunteerMessage(eom, as.GetAuds()[0])
	vf := as.NewVoteFactory(vol)
	var _ = vf

	vc := NewVolunteerControl(me, as.GetAuthSet())

	for i := 0; i < as.Majority()-1; i++ {
		msg := vc.Execute(messages.NewLeaderLevelMessage(as.NextIdentity(), 0, 1, vol))
		if len(msg) != 0 {
			t.Error("Do not expect any msgs to be returned")
		}
	}

	result := vc.Execute(messages.NewLeaderLevelMessage(as.NextIdentity(), 0, 1, vol))
	if len(result) == 0 {
		t.Error("Expected a message back")
	} else {
		ll := result[0].(messages.LeaderLevelMessage)
		if ll.Rank != 1 {
			t.Errorf("Expect rank 1, got %d", ll.Rank)
		}
	}

}
