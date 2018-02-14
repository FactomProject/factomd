package controller

import (
	"fmt"
	"github.com/FactomProject/electiontesting/election"
	"github.com/FactomProject/electiontesting/imessage"
)

var _ = fmt.Println

// Router helps keep route patterns for returning messages
// Each election has a queue, and each step consumes 1 message from the queue
// and adds to some number of queues
type Router struct {
	ElectionQueues []chan imessage.IMessage
	Elections      []*election.Election
}

func NewRouter(elections []*election.Election) *Router {
	r := new(Router)
	r.Elections = elections
	r.ElectionQueues = make([]chan imessage.IMessage, len(elections))

	for i := range r.ElectionQueues {
		r.ElectionQueues[i] = make(chan imessage.IMessage, 10000)
	}

	return r
}

func (r *Router) Run() {
	for !r.Step() {
		r.Step()
	}
}

func (r *Router) Step() bool {
	for i := range r.ElectionQueues {
		if len(r.ElectionQueues[i]) > 0 {
			msg := <-r.ElectionQueues[i]
			resp, _ := r.Elections[i].Execute(msg)
			r.route(i, resp)
		}
	}

	// Check all for complete
	for _, e := range r.Elections {
		if !e.Committed {
			return false
		}
	}

	// All complete!
	return true
}

func (r *Router) route(from int, msg imessage.IMessage) {
	// TODO: Adust routing behavior
	for _, c := range r.ElectionQueues {
		c <- msg
	}
}
