package main

import (
	"fmt"
	"github.com/FactomProject/electiontesting/controller"
	"github.com/FactomProject/electiontesting/election"
	"github.com/FactomProject/electiontesting/messages"
)

func main() {
	recurse(3, 3, 40)
}

// newElections will return an array of elections (1 per leader) and an array
// of volunteers messages to kick things off.
//		Params:
//			feds int   Number of Federated Nodes
//			auds int   Number of Volunteers
//			noDisplay  Passing a true here will reduce memory consumption, as it is a debugging tool
//
//		Returns:
//			controller *Controller  This can used for debugging (Printing votes)
//			elections []*election   Nodes you can execute on (returns msg, statchange)
//			volmsgs   []*VoluntMsg	Volunteer msgs you can start things with
func newElections(feds, auds int, noDisplay bool) (*controller.Controller, []*election.Election, []*messages.VolunteerMessage) {
	con := controller.NewController(feds, auds)

	if noDisplay {
		for _, e := range con.Elections {
			e.Display = nil
		}
		con.GlobalDisplay = nil
	}

	return con, con.Elections, con.Volunteers
}

var breath = 0

func dive(msgs []int, leaders int, depth int, limit int) {
	depth++
	if depth > limit {
		fmt.Println("Breath ", breath)
		breath++
		return
	}

	for d, v := range msgs {
		msgs2 := append(msgs[0:d], msgs[d+1:]...)
		ml2 := len(msgs2)
		for i := 0; i < leaders-1; i++ {
			msgs2 = append(msgs2, v)
			dive(msgs2, leaders, depth, limit)
			msgs2 = msgs2[:ml2]
		}
	}
}

func recurse(msg int, leaders int, limit int) {
	var msgs []int
	for i := 0; i < msg; i++ {
		for j := 0; j < leaders-1; j++ {
			msgs = append(msgs, i*1024+j)
		}
	}
	dive(msgs, leaders, 0, limit)
}
