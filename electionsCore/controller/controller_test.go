package controller_test

import (
	"fmt"
	"testing"

	. "github.com/FactomProject/factomd/electionsCore/controller"
	. "github.com/FactomProject/factomd/electionsCore/errorhandling"
)

func TestSimpleController(t *testing.T) {
	StartUnitTestErrorHandling(t)

	con := NewController(3, 3)
	all := []int{0, 1, 2}
	con.RouteVolunteerMessage(1, all)

}

func TestElectionDisplay(t *testing.T) {
	//StartUnitTestErrorHandling(t)
	//
	//con := NewController(3, 3)
	//all := []int{0, 1, 2}
	//con.RouteVolunteerMessage(1, all)
	//fmt.Println(con.GlobalDisplay.String())
	//
	//ExpMsg(con.RouteLeaderSetLevelMessage(all, 0, all))
	//fmt.Println(con.GlobalDisplay.String())
	//
	//fmt.Println(con.Elections[0].Display.String())
	//
	//ExpMsg(con.RouteLeaderLevelMessage(1, 0, all))
	//ExpMsg(con.RouteLeaderLevelMessage(2, 0, all))
	//
	//ExpMsg(con.RouteLeaderLevelMessage(0, 2, all))

	//fmt.Println(con.GlobalDisplay.String())
}

// TestElectionSimpleScenario will test 100% consensus
func TestElectionSimpleScenario(t *testing.T) {
	return
	StartUnitTestErrorHandling(t)

	con := NewController(3, 3)
	all := []int{0, 1, 2}
	con.SendOutputsToRouter(true)
	con.RouteVolunteerMessage(1, all)
	con.RouteLeaderSetVoteMessage(all, 1, all)
	con.RouteLeaderSetLevelMessage(all, 1, all)
	con.RouteLeaderSetLevelMessage(all, 2, all)

	loop := con.GlobalDisplay.DetectVerticalLoop(con.Elections[0].Self)
	if loop {
		t.Errorf("Detected a vert loop when there is not one")
	}

	runToComplete(con, t)
}

func TestFlipFlop(t *testing.T) {
	StartUnitTestErrorHandling(t)

	con := NewController(3, 3)
	all := []int{0, 1, 2}
	left := []int{0, 1}
	right := []int{1, 2}
	mid := []int{1}
	fright := []int{2} // far right

	// <setup>
	con.RouteVolunteerMessage(1, all)
	con.RouteVolunteerMessage(2, all)

	//(Leader 0)
	//  Lvl  L0  L1  L2
	//  0:   1   1
	//  1: 0.1
	//(Leader 1)
	//  Lvl  L0  L1  L2
	//  0:   1   1
	//  1:     0.1
	con.RouteLeaderSetVoteMessage(left, 1, left)
	fmt.Println(con.ElectionStatus(-1))

	// We need L1 to switch to 2
	//(Leader 1)
	//  Lvl  L0  L1  L2
	//  0:   1  12   2
	//  1:     0.1
	//  2:     0.2
	//
	//(Leader 2)
	//  Lvl  L0  L1  L2
	//  0:       2   2
	//  1:         0.2
	con.RouteLeaderSetVoteMessage(right, 2, right)

	// Let's flop 1 the other way
	//(Global)
	//Lvl  L0  L1  L2
	//0:   1  12   2
	//1: 0.1 0.1 0.2
	//2: 1.1 0.2
	//3:     1.1
	//
	//(Leader 0)
	//Lvl  L0  L1  L2
	//0:   1   1
	//1: 0.1 0.1
	//2: 1.1
	//
	//(Leader 1)
	//Lvl  L0  L1  L2
	//0:   1  12   2
	//1: 0.1 0.1
	//2:     0.2
	//3:     1.1
	//
	//(Leader 2)
	//Lvl  L0  L1  L2
	//0:       2   2
	//1:         0.2
	con.RouteLeaderSetLevelMessage(left, 1, left)

	//(Global)
	//Lvl  L0  L1  L2
	//0:   1  12   2
	//1: 0.1 0.1 0.2
	//2: 1.1 0.2 1.2
	//3:     1.1
	//4:     1.2
	//
	//(Leader 0)
	//Lvl  L0  L1  L2
	//0:   1   1
	//1: 0.1 0.1
	//2: 1.1
	//
	//(Leader 1)
	//Lvl  L0  L1  L2
	//0:   1  12   2
	//1: 0.1 0.1 0.2
	//2:     0.2
	//3:     1.1
	//4:     1.2
	//
	//(Leader 2)
	//Lvl  L0  L1  L2
	//0:       2   2
	//1:         0.2
	//2:     0.2 1.2

	con.RouteLeaderSetLevelMessage(mid, 2, right)
	con.RouteLeaderSetLevelMessage(fright, 1, mid)
	// </setup>

	con.SendOutputsToRouter(true)

	// <resolve>
	con.RouteLeaderLevelMessage(2, 1, []int{0})
	con.RouteLeaderLevelMessage(1, 2, []int{0})
	con.RouteLeaderLevelMessage(1, 4, []int{0})

	con.RouteLeaderLevelMessage(2, 2, []int{0})
	// </resolve>

	con.AddLeaderSetLevelMessageToRouter(all, 2)
	con.AddLeaderSetLevelMessageToRouter(all, 3)
	con.AddLeaderSetLevelMessageToRouter(all, 4)
	runToComplete(con, t)
}

func TestStrange(t *testing.T) {
	StartUnitTestErrorHandling(t)

	con := NewController(3, 3)
	all := []int{0, 1, 2}
	span := []int{0, 1, 2, 1, 0}
	con.SendOutputsToRouter(true)
	con.RouteVolunteerMessage(0, span)
	con.RouteVolunteerMessage(1, span)
	con.RouteVolunteerMessage(2, span)

	con.RouteLeaderSetVoteMessage(all, 0, span)
	con.RouteLeaderSetVoteMessage(all, 1, span)
	con.RouteLeaderSetVoteMessage(all, 2, []int{0, 2})

	con.RouteLeaderSetLevelMessage(all, 1, []int{0})
	con.RouteLeaderSetLevelMessage(all, 2, []int{0})
	con.RouteLeaderSetLevelMessage(all, 3, []int{0})
	con.RouteLeaderLevelMessage(2, 3, []int{0})

	t.Log(con.ElectionStatus(-1))
	t.Log(con.ElectionStatus(0))
	//t.Log(con.ElectionStatus(1))
	//t.Log(con.ElectionStatus(2))

	//con.RouteLeaderSetLevelMessage(all, 2, all)

	runToCompleteNoPrint(con, t)
}

func getVerticalFlipFlop(t *testing.T) *Controller {
	StartUnitTestErrorHandling(t)

	con := NewController(3, 3)
	all := []int{0, 1, 2, 1, 0}
	con.SendOutputsToRouter(true)
	con.RouteVolunteerMessage(0, all)
	con.RouteVolunteerMessage(1, all)
	con.RouteVolunteerMessage(2, all)

	con.RouteLeaderSetVoteMessage(all, 0, all)
	con.RouteLeaderSetVoteMessage(all, 1, all)
	con.RouteLeaderSetVoteMessage(all, 2, all)

	con.RouteLeaderSetLevelMessage(all, 1, all)
	con.RouteLeaderSetLevelMessage(all, 2, all)
	con.RouteLeaderSetLevelMessage(all, 3, all)
	loop := con.GlobalDisplay.DetectVerticalLoop(con.Elections[0].Self)
	if !loop {
		t.Errorf("Did not detect vertical loop when there was")
	}

	con.RouteLeaderSetLevelMessage(all, 4, all)
	con.RouteLeaderSetLevelMessage(all, 5, all)
	con.RouteLeaderSetLevelMessage(all, 6, all)

	//t.Log(con.ElectionStatus(-1))
	//t.Log(con.ElectionStatus(0))
	//t.Log(con.ElectionStatus(1))
	//t.Log(con.ElectionStatus(2))

	loop = con.GlobalDisplay.DetectVerticalLoop(con.Elections[0].Self)
	if !loop {
		t.Errorf("Did not detect vertical loop when there was")
	}
	return con
}

func TestVerticalFlipFlop(t *testing.T) {
	all := []int{0, 1, 2, 1, 0}
	con := getVerticalFlipFlop(t)
	con.RouteLeaderSetLevelMessage(all, 2, all)
	con.AddLeaderSetLevelMessageToRouter(all, 7)
	runToComplete(con, t)
}

func TestNormalizedStates(t *testing.T) {
	con := getVerticalFlipFlop(t)
	cl := con.Elections[0].Copy()

	if string(cl.StateString()) != string(con.Elections[0].StateString()) {
		t.Errorf("Normalized strings should be the same on clones")
	}

	for i := 0; i < 2; i++ {
		for j := 1; j < 3; j++ {
			if string(con.Elections[i].StateString()) != string(con.Elections[j].StateString()) {
				t.Errorf("State strings should be the same in this case")
			}
			if string(con.Elections[i].NormalizedString()) != string(con.Elections[j].NormalizedString()) {
				t.Errorf("State strings should be the same in this case")
			}
		}
	}

	t.Log(string(con.ElectionStatus(0)))
	t.Log(string(con.Elections[0].NormalizedString()))
	t.Log(string(con.Elections[0].StateString()))
}

// BenchmarkNormalizedString-8   	  200000	      6762 ns/op

func BenchmarkNormalizedString(b *testing.B) {
	var t testing.T
	con := getVerticalFlipFlop(&t)
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		str := con.Elections[n%3].StateString()
		var _ = str
	}
}

func Fib(n int) int {
	if n < 2 {
		return n
	}
	return Fib(n-1) + Fib(n-2)
}

func runToCompleteNoPrint(con *Controller, t *testing.T) {
	con.Router.StepN(100)
	if !con.Complete() {
		t.Errorf("Did not complete")
	}
}

func runToComplete(con *Controller, t *testing.T) {
	con.Router.StepN(100)
	t.Log(con.GlobalDisplay.String())
	if !con.Complete() {
		t.Errorf("Did not complete")
	}
	t.Log(con.GlobalDisplay.String())
}

func expmsg(found bool, t *testing.T) {
	if !found {
		t.Errorf("Expected message, got nil")
	}
}
