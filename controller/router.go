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
	// Number of messages consumed
	Consumed       []int
	Generated      []int
	ElectionQueues []chan imessage.IMessage
	Elections      []*election.Election

	Printing bool
}

func NewRouter(elections []*election.Election) *Router {
	r := new(Router)
	r.Elections = elections
	r.ElectionQueues = make([]chan imessage.IMessage, len(elections))
	r.Consumed = make([]int, len(elections))
	r.Generated = make([]int, len(elections))

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

func pad(row string) string {
	return fmt.Sprintf(" %4s", fmt.Sprintf("%s:", row))
}

func center(str string) string {
	return fmt.Sprintf("%-4s", fmt.Sprintf("%4s", str))
}

func (r *Router) PrintMode(active bool) {
	r.Printing = active
}

func (r *Router) Status() string {
	str := pad("")
	for i := range r.Elections {
		str += center(fmt.Sprintf("L%d", i))
	}
	str += "\n" + pad("Q:")
	for _, q := range r.ElectionQueues {
		str += center(fmt.Sprintf("%d", len(q)))
	}
	str += "\n" + pad("C:")
	for i := range r.Consumed {
		str += center(fmt.Sprintf("%d", r.Consumed[i]))
	}
	str += "\n" + pad("G:")
	for i := range r.Generated {
		str += center(fmt.Sprintf("%d", r.Generated[i]))
	}

	return str
}

var c int = 0

func (r *Router) Step() bool {
	for i := range r.ElectionQueues {
		if len(r.ElectionQueues[i]) > 0 {
			r.Consumed[i]++
			msg := <-r.ElectionQueues[i]
			resp, _ := r.Elections[i].Execute(msg)
			if resp == nil {
				r.Generated[i]++
				r.route(i, resp)
			}

			if r.Printing {
				str := fmt.Sprintf("L%d: ", i)
				str += fmt.Sprintf(" Consumed(%s)", r.Elections[0].Display.FormatMessage(msg))
				str += fmt.Sprintf(" Generated(%s)", r.Elections[0].Display.FormatMessage(resp))
				fmt.Println(str)
			}
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
	if msg == nil {
		return
	}
	// TODO: Adust routing behavior
	for _, c := range r.ElectionQueues {
		// fmt.Println(len(c))
		c <- msg
	}
}
