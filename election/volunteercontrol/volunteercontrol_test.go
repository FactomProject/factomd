package volunteercontrol_test

import (
	"fmt"
	"testing"

	. "github.com/FactomProject/electiontesting/election/volunteercontrol"
)

var _ = NewVolunteerControl
var _ = fmt.Println

//func TestSimpleVolunteerControl(t *testing.T) {
//	as := testhelper.NewAuthSetHelper(3, 3)
//	me := as.NextIdentity()
//
//	var loc primitives.ProcessListLocation
//	eom := messages.NewEomMessage(as.GetAuds()[0], loc)
//	vol := messages.NewVolunteerMessage(eom, as.GetAuds()[0])
//	vf := as.NewVoteFactory(vol)
//	var _ = vf
//
//	vc := NewVolunteerControl(me, as.GetAuthSet())
//
//	for i := 0; i < as.Majority()-1; i++ {
//		f := as.NextIdentity()
//		l := messages.NewLeaderLevelMessage(f, 0, 1, vol)
//		msg := vc.Execute(&l)
//		if msg != nil && msg.(*messages.LeaderLevelMessage).Signer != f {
//			t.Error("Do not expect any msgs to be returned")
//		}
//	}
//
//	l := messages.NewLeaderLevelMessage(as.NextIdentity(), 0, 1, vol)
//	result := vc.Execute(&l)
//	if result == nil {
//		t.Error("Expected a message back")
//	} else {
//		ll := result.(*messages.LeaderLevelMessage)
//		if ll.Rank != 1 {
//			t.Errorf("Expect rank 1, got %d", ll.Rank)
//		}
//	}
//}

func TestSort(t *testing.T) {
	arr := []int{8, 2, 1, 54, 6, 12, 6, 8, 2, 34, 0, 43, 12, 2, 3465, 68, 123}
	for i := 1; i < len(arr); i++ {
		for j := 0; j < len(arr)-i; j++ {
			if arr[j] < arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
	prev := arr[0]
	for _, a := range arr[1:] {
		if a > prev {
			t.Errorf("Sort bad")
		}
		t.Log(a)
		prev = a
	}
}
