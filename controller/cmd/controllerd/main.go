package main

import (
	"fmt"

	. "github.com/FactomProject/electiontesting/controller"
)

func main() {
	con := NewController(3, 3)
	all := []int{0, 1, 2}
	con.SendOutputsToRouter(true)
	con.RouteVolunteerMessage(1, all)

	con.Router.Run()
	fmt.Println(con.ElectionStatus(-1))
	fmt.Println(con.Complete())
	//con.Shell()
}

func expmsg(found bool) {
	if !found {
		fmt.Println("Expected message, got nil")
	}
}
