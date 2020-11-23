package controller

import (
	"fmt"

	"github.com/FactomProject/factomd/electionsCore/election"
	"github.com/FactomProject/factomd/electionsCore/imessage"
)

// Router helps keep route patterns for returning messages
// Each election has a queue, and each step consumes 1 message from the queue
// and adds to some number of queues
type Router struct {
	// Number of messages consumed
	Consumed       []int
	Generated      []int
	ElectionQueues []chan imessage.IMessage
	RepeatedFilter []map[imessage.IMessage]struct{}
	Elections      []*election.RoutingElection

	Printing bool
}

func NewRouter(elections []*election.RoutingElection) *Router {
	r := new(Router)
	r.Elections = elections
	r.ElectionQueues = make([]chan imessage.IMessage, len(elections))
	r.Consumed = make([]int, len(elections))
	r.Generated = make([]int, len(elections))
	r.RepeatedFilter = make([]map[imessage.IMessage]struct{}, len(elections))

	for i := range r.ElectionQueues {
		r.ElectionQueues[i] = make(chan imessage.IMessage, 10000)
		r.RepeatedFilter[i] = make(map[imessage.IMessage]struct{})
	}

	return r
}

func (r *Router) StepN(step int) {
	for i := 0; i < step; i++ {
		r.Step()
	}
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

// NodeStack will return the messages in the node's input queue
func (r *Router) NodeStack(node int) string {
	str := fmt.Sprintf("-- Node %d --\n", node)
	msgs := make([]imessage.IMessage, len(r.ElectionQueues[node]))
	i := 0
ForLoop:
	for {
		select {
		case m := <-r.ElectionQueues[node]:
			str += fmt.Sprintf("%d: %s\n", i, r.Elections[0].Display.FormatMessage(m))
			msgs[i] = m
			i++
		default:
			break ForLoop
		}
	}

	for _, m := range msgs {
		r.ElectionQueues[node] <- m
	}
	return str
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
	oneChange := false
	for i := range r.ElectionQueues {
		if len(r.ElectionQueues[i]) > 0 {
			oneChange = true
			r.Consumed[i]++
			msg := <-r.ElectionQueues[i]
			resp, _ := r.Elections[i].Execute(msg)
			if resp != nil {
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

	if !oneChange {
		return true
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

// routeIncoming forwards all incoming messages to all connected nodes. The incoming router needs
// to not forward repeated messages to stop infinite recursive behavior
func (r *Router) routeIncoming(from int, msg imessage.IMessage) {
	if msg == nil {
		return
	}
	// TODO: Adjust routing behavior
	for i := range r.ElectionQueues {
		r.acceptIncoming(i, msg)
	}
}

// acceptIncoming takes a msg to a particular node. It will handle the onincoming behavior
func (r *Router) acceptIncoming(to int, msg imessage.IMessage) {
	if msg == nil {
		return
	}

	if _, ok := r.RepeatedFilter[to][msg]; ok {
		return
	}

	r.RepeatedFilter[to][msg] = struct{}{}
	r.ElectionQueues[to] <- msg

	r.routeIncoming(to, msg)
}

func (r *Router) route(from int, msg imessage.IMessage) {
	if msg == nil {
		return
	}
	// TODO: Adust routing behavior
	for i := range r.ElectionQueues {
		r.acceptIncoming(i, msg)
	}
}
