package controller

import (
	"fmt"

	. "github.com/PaulSnow/factom2d/electionsCore/ET2/directedmessage"
	"github.com/PaulSnow/factom2d/electionsCore/election"
	"github.com/PaulSnow/factom2d/electionsCore/imessage"
	priminterpreter "github.com/PaulSnow/factom2d/electionsCore/interpreter/primitives"
	"github.com/PaulSnow/factom2d/electionsCore/messages"
	"github.com/PaulSnow/factom2d/electionsCore/primitives"
	"github.com/PaulSnow/factom2d/electionsCore/testhelper"
)

var _ = fmt.Println

type ControllerInterpreter struct {
	*Controller

	*priminterpreter.Primitives
}

// Controller will be able to route messages to a set of nodes and control various
// communication patterns
type Controller struct {
	AuthSet *testhelper.AuthSetHelper

	RoutingElections []*election.RoutingElection
	Elections        []*election.Election
	Volunteers       []*messages.VolunteerMessage

	feds []primitives.Identity
	auds []primitives.Identity

	// Can be all 0s for now
	primitives.ProcessListLocation

	Buffer        *MessageBuffer
	GlobalDisplay *election.Display

	Router          *Router
	OutputsToRouter bool

	BufferingMessages bool
	BufferedMessages  []*DirectedMessage

	PrintingTrace bool

	History []string
}

func NewControllerInterpreter(feds, auds int) *ControllerInterpreter {
	c := new(ControllerInterpreter)
	c.Controller = NewController(feds, auds)
	c.InitInterpreter()

	return c
}

// NewController creates all the elections and initial volunteer messages
func NewController(feds, auds int) *Controller {
	c := new(Controller)
	c.AuthSet = testhelper.NewAuthSetHelper(feds, auds)
	fedlist := c.AuthSet.GetFeds()
	c.Elections = make([]*election.Election, len(fedlist))
	c.RoutingElections = make([]*election.RoutingElection, len(fedlist))

	for i, f := range fedlist {
		c.Elections[i] = c.newElection(f)
		if i == 0 {
			c.GlobalDisplay = election.NewDisplay(c.Elections[0], nil)
			c.GlobalDisplay.Identifier = "Global"
		}
		c.Elections[i].AddDisplay(c.GlobalDisplay)
		c.RoutingElections[i] = election.NewRoutingElection(c.Elections[i])
	}

	audlist := c.AuthSet.GetAuds()
	c.Volunteers = make([]*messages.VolunteerMessage, len(audlist))

	for _, a := range audlist {
		c.Volunteers[c.Elections[0].GetVolunteerPriority(a)] = c.newVolunteer(a)
	}

	c.Buffer = NewMessageBuffer()
	c.feds = fedlist
	c.auds = audlist

	c.Router = NewRouter(c.RoutingElections)
	return c
}

func (c *Controller) Complete() bool {
	for _, e := range c.Elections {
		if !e.Committed {
			return false
		}
	}
	return true
}

func (c *Controller) SendOutputsToRouter(set bool) {
	c.OutputsToRouter = set
}

func (c *Controller) ElectionStatus(node int) string {
	if node == -1 {
		return c.GlobalDisplay.String()
	}
	return c.Elections[node].Display.String()
}

func (c *Controller) AddLeaderSetLevelMessageToRouter(from []int, level int) bool {
	for _, f := range from {
		if !c.AddLeaderLevelMessageToRouter(f, level) {
			return false
		}
	}
	return true
}

func (c *Controller) AddLeaderLevelMessageToRouter(from int, level int) bool {
	msg := c.Buffer.RetrieveLeaderLevelMessageByLevel(c.indexToFedID(from), level)
	if msg == nil {
		return false
	}
	c.routeToRouter(msg, from)
	return true
}

func (c *Controller) RouteLeaderSetLevelMessage(from []int, level int, to []int) bool {
	for _, f := range from {
		if !c.RouteLeaderLevelMessage(f, level, to) {
			return false
		}
	}
	return true
}

func (c *Controller) RouteLeaderLevelMessage(from int, level int, to []int) bool {
	msg := c.Buffer.RetrieveLeaderLevelMessageByLevel(c.indexToFedID(from), level)
	if msg == nil {
		return false
	}
	c.RouteMessage(msg, to)
	return true
}

func (c *Controller) RouteLeaderSetVoteMessage(from []int, vol int, to []int) bool {
	for _, f := range from {
		if !c.RouteLeaderVoteMessage(f, vol, to) {
			return false
		}
	}
	return true
}

func (c *Controller) RouteLeaderVoteMessage(from int, vol int, to []int) bool {
	msg := c.Buffer.RetrieveLeaderVoteMessage(c.indexToFedID(from), c.indexToAudID(vol))
	if msg == nil {
		return false
	}
	c.RouteMessage(msg, to)
	return true
}

func (c *Controller) RouteVolunteerMessage(vol int, nodes []int) {
	c.RouteMessage(c.Volunteers[vol], nodes)
}

// RouteMessage will execute message on all nodes given and add the messages returned to the buffer
func (c *Controller) RouteMessage(msg imessage.IMessage, nodes []int) {
	for _, n := range nodes {
		c.routeSingleNode(msg, n)
	}
}

func (c *Controller) routeSingleNode(msg imessage.IMessage, node int) {
	if c.BufferingMessages {
		c.BufferedMessages = append(c.BufferedMessages, &DirectedMessage{node, msg})
		if c.PrintingTrace {
			str := fmt.Sprintf("L%d: ", node)
			str += fmt.Sprintf(" Buffered(%s)", c.Elections[node].Display.FormatMessage(msg))
			fmt.Println(str)
		}
		return
	}

	var resp imessage.IMessage = nil
	var statechange bool = false
	if c.OutputsToRouter {
		resp, statechange = c.RoutingElections[node].Execute(msg)
	} else {
		resp, statechange = c.Elections[node].Execute(msg, 0)
	}

	if c.OutputsToRouter {
		// Outputs get sent to Router so we can hit "run"
		//f := messages.GetSigner(msg)
		//c.Router.route(c.fedIDtoIndex(f), msg)
		c.Router.route(node, msg)
	}
	c.Buffer.Add(resp)

	if c.PrintingTrace {
		str := fmt.Sprintf("L%d: ", node)
		str += fmt.Sprintf(" Consumed(%s)", c.Elections[node].Display.FormatMessage(msg))
		str += fmt.Sprintf(" Generated(%s) StateChange: %t", c.Elections[node].Display.FormatMessage(resp), statechange)
		fmt.Println(str)
	}
}

func (c *Controller) routeToRouter(msg imessage.IMessage, node int) {
	c.Router.acceptIncoming(node, msg)
}

// indexToAudID will take the human legible "Audit 1" and get the correct identity.
func (c *Controller) indexToAudID(index int) primitives.Identity {
	// TODO: Actually implement some logic if this changes
	return c.AuthSet.PriorityToIdentityMap[index]

}

func (c *Controller) fedIDtoIndex(id primitives.Identity) int {
	for i, f := range c.feds {
		if f == id {
			return i
		}
	}
	return -1
}

// indexToFedID will take the human legible "Leader 1" and get the correct identity
func (c *Controller) indexToFedID(index int) primitives.Identity {
	// TODO: Actually implement some logic if this changes
	return c.feds[index]
}

func (c *Controller) newVolunteer(id primitives.Identity) *messages.VolunteerMessage {
	vol := messages.NewVolunteerMessage(messages.NewEomMessage(id, c.ProcessListLocation), id)
	return &vol
}

func (c *Controller) newElection(id primitives.Identity) *election.Election {
	return election.NewElection(id, c.AuthSet.GetAuthSet())
}
