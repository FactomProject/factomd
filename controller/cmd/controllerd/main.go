package main

import (
	"fmt"

	. "github.com/FactomProject/electiontesting/controller"
)

func main() {
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

	con.RouteLeaderSetLevelMessage(all, 4, all)
	con.RouteLeaderSetLevelMessage(all, 5, all)
	con.RouteLeaderSetLevelMessage(all, 6, all)
	con.AddLeaderSetLevelMessageToRouter(all, 5)
	con.AddLeaderSetLevelMessageToRouter(all, 6)

	//t.Log(con.ElectionStatus(-1))
	//t.Log(con.ElectionStatus(0))
	//t.Log(con.ElectionStatus(1))
	//t.Log(con.ElectionStatus(2))

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
