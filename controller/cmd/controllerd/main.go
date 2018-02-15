package main

import (
	"fmt"

	. "github.com/FactomProject/electiontesting/controller"
)

func main() {
	con := NewController(3, 3)
	all := []int{0, 1, 2}
	left := []int{0, 1}
	right := []int{1, 2}
	mid := []int{1}
	fright := []int{2} // far right

	con.SendOutputsToRouter(true)
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

	// Print end result
	fmt.Println(con.GlobalDisplay.String())
	fmt.Println(con.ElectionStatus(0))
	fmt.Println(con.ElectionStatus(1))
	fmt.Println(con.ElectionStatus(2))

	con.Shell()
	//con.Router.Run()
	fmt.Println("DONE!")
}

func expmsg(found bool) {
	if !found {
		fmt.Println("Expected message, got nil")
	}
}
