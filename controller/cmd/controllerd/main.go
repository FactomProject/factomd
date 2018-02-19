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

	// <setup>
	con.RouteVolunteerMessage(1, all)
	con.RouteVolunteerMessage(2, all)

	con.RouteLeaderSetVoteMessage(left, 1, left)
	con.RouteLeaderSetVoteMessage(right, 2, right)
	con.RouteLeaderSetLevelMessage(left, 1, left)
	con.RouteLeaderSetLevelMessage(mid, 2, right)
	con.RouteLeaderSetLevelMessage(fright, 1, mid)

	// </setup>

	// <resolve>
	con.SendOutputsToRouter(true)

	con.RouteLeaderLevelMessage(2, 1, []int{0})
	con.RouteLeaderLevelMessage(1, 2, []int{0})
	con.RouteLeaderLevelMessage(1, 4, []int{0})

	con.RouteLeaderLevelMessage(2, 2, []int{0})

	con.Shell()

	con.AddLeaderSetLevelMessageToRouter(all, 2)
	con.AddLeaderSetLevelMessageToRouter(all, 3)
	con.AddLeaderSetLevelMessageToRouter(all, 4)
	con.AddLeaderSetLevelMessageToRouter(all, 5)

	con.Shell()

	con.Router.Run()
	con.Router.StepN(1000)

	fmt.Println(con.ElectionStatus(-1))
	fmt.Println(con.Complete())
	//con.Shell()
}

func expmsg(found bool) {
	if !found {
		fmt.Println("Expected message, got nil")
	}
}
